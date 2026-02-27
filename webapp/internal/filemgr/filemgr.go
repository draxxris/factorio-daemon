package filemgr

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Manager handles file operations for staging, backup, and deployment
type Manager struct {
	baseDir    string
	stagingDir string
	backupDir  string
}

// NewManager creates a new file manager
func NewManager(baseDir, stagingDir, backupDir string) *Manager {
	return &Manager{
		baseDir:    baseDir,
		stagingDir: stagingDir,
		backupDir:  backupDir,
	}
}

// Backup represents a save file backup
type Backup struct {
	Filename  string    `json:"filename"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
}

// StageFile saves an uploaded file to the staging directory
func (m *Manager) StageFile(instance, filename string, data []byte) error {
	// Validate filename
	allowedFiles := map[string]bool{
		"mod-list.json":    true,
		"mod-settings.dat": true,
	}
	
	// Check if it's a zip file (save)
	isZip := strings.HasSuffix(strings.ToLower(filename), ".zip")
	
	if !allowedFiles[filename] && !isZip {
		return fmt.Errorf("file type not allowed: %s", filename)
	}
	
	// Create staging directory for instance
	instanceStaging := filepath.Join(m.stagingDir, instance)
	if err := os.MkdirAll(instanceStaging, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	
	// Determine target filename
	targetFilename := filename
	if isZip {
		targetFilename = "main.zip"
	}
	
	// Write file
	targetPath := filepath.Join(instanceStaging, targetFilename)
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write staged file: %w", err)
	}
	
	return nil
}

// GetStagedFiles returns list of staged files for an instance
func (m *Manager) GetStagedFiles(instance string) ([]string, error) {
	instanceStaging := filepath.Join(m.stagingDir, instance)
	
	entries, err := os.ReadDir(instanceStaging)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read staging directory: %w", err)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// ClearStagedFiles removes all staged files for an instance
func (m *Manager) ClearStagedFiles(instance string) error {
	instanceStaging := filepath.Join(m.stagingDir, instance)
	
	if err := os.RemoveAll(instanceStaging); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear staging directory: %w", err)
	}
	
	return nil
}

// DeployFiles moves staged files to the instance directory
// IMPORTANT: The service must be stopped before calling this
func (m *Manager) DeployFiles(instance string) error {
	instanceStaging := filepath.Join(m.stagingDir, instance)
	instanceDir := filepath.Join(m.baseDir, instance)
	
	// Check if there are staged files
	stagedFiles, err := m.GetStagedFiles(instance)
	if err != nil {
		return err
	}
	if len(stagedFiles) == 0 {
		return fmt.Errorf("no staged files to deploy")
	}
	
	// Ensure instance directory exists
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}
	
	// Move each staged file
	for _, file := range stagedFiles {
		src := filepath.Join(instanceStaging, file)
		dst := filepath.Join(instanceDir, file)
		
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s: %w", file, err)
		}
	}
	
	return nil
}

// BackupSave creates a backup of the main.zip save file
func (m *Manager) BackupSave(instance string) error {
	savePath := filepath.Join(m.baseDir, instance, "saves", "main.zip")
	
	// Check if save file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return fmt.Errorf("no save file found to backup")
	}
	
	// Create backup directory for instance
	instanceBackup := filepath.Join(m.backupDir, instance)
	if err := os.MkdirAll(instanceBackup, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_150405")
	backupFilename := fmt.Sprintf("%s_main.zip", timestamp)
	backupPath := filepath.Join(instanceBackup, backupFilename)
	
	// Copy save file to backup
	if err := copyFile(savePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	return nil
}

// ListBackups returns list of save backups for an instance
func (m *Manager) ListBackups(instance string) ([]Backup, error) {
	instanceBackup := filepath.Join(m.backupDir, instance)
	
	entries, err := os.ReadDir(instanceBackup)
	if os.IsNotExist(err) {
		return []Backup{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// Parse timestamp from filename (format: 2006-01-02_150405_main.zip)
		name := entry.Name()
		if !strings.HasSuffix(name, "_main.zip") {
			continue
		}
		
		timestampStr := strings.TrimSuffix(name, "_main.zip")
		timestamp, err := time.Parse("2006-01-02_150405", timestampStr)
		if err != nil {
			continue // Skip files with invalid timestamp format
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		backups = append(backups, Backup{
			Filename:  name,
			Timestamp: timestamp,
			Size:      info.Size(),
		})
	}
	
	// Sort by timestamp descending (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})
	
	return backups, nil
}

// RestoreBackup restores a save file from backup
func (m *Manager) RestoreBackup(instance, backupFilename string) error {
	backupPath := filepath.Join(m.backupDir, instance, backupFilename)
	saveDir := filepath.Join(m.baseDir, instance, "saves")
	savePath := filepath.Join(saveDir, "main.zip")
	
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFilename)
	}
	
	// Ensure saves directory exists
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create saves directory: %w", err)
	}
	
	// Copy backup to save location
	if err := copyFile(backupPath, savePath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Preserve permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}
