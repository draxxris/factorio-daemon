package instance

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Instance represents a Factorio server instance
type Instance struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Port            int    `json:"port"`
	NonBlockingSave bool   `json:"non_blocking_save"`
	Running         bool   `json:"running"`
	Enabled         bool   `json:"enabled"`
}

// Manager handles instance discovery and configuration
type Manager struct {
	baseDir string
}

// NewManager creates a new instance manager
func NewManager(baseDir string) *Manager {
	return &Manager{baseDir: baseDir}
}

// ListInstances discovers all instances by scanning for env-* files
func (m *Manager) ListInstances() ([]Instance, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	var instances []Instance
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "env-") {
			// Skip env-example as it's a sample file
			if name == "env-example" {
				continue
			}
			instanceName := strings.TrimPrefix(name, "env-")
			instance, err := m.GetInstance(instanceName)
			if err != nil {
				// Log error but continue with other instances
				fmt.Printf("Warning: failed to parse %s: %v\n", name, err)
				continue
			}
			instances = append(instances, *instance)
		}
	}

	return instances, nil
}

// GetInstance reads and parses an env file for a specific instance
func (m *Manager) GetInstance(name string) (*Instance, error) {
	envPath := filepath.Join(m.baseDir, "env-"+name)
	
	file, err := os.Open(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	instance := &Instance{Name: name}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "NAME":
			instance.Name = value
		case "VERSION":
			instance.Version = value
		case "TITLE":
			instance.Title = value
		case "DESCRIPTION":
			instance.Description = value
		case "PORT":
			fmt.Sscanf(value, "%d", &instance.Port)
		case "NON_BLOCKING_SAVE":
			instance.NonBlockingSave = strings.ToLower(value) == "true"
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	return instance, nil
}

// CreateInstance creates a new env file for an instance
func (m *Manager) CreateInstance(instance Instance) error {
	// Validate instance name (alphanumeric and hyphens only)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, instance.Name)
	if !matched {
		return fmt.Errorf("instance name must contain only alphanumeric characters and hyphens")
	}

	envPath := filepath.Join(m.baseDir, "env-"+instance.Name)
	
	// Check if already exists
	if _, err := os.Stat(envPath); err == nil {
		return fmt.Errorf("instance %s already exists", instance.Name)
	}

	content := fmt.Sprintf(`NAME=%s
VERSION=%s
TITLE=%q
DESCRIPTION=%q
PORT=%d
NON_BLOCKING_SAVE=%t
`, instance.Name, instance.Version, instance.Title, instance.Description, instance.Port, instance.NonBlockingSave)

	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create env file: %w", err)
	}

	// Create instance directory
	instanceDir := filepath.Join(m.baseDir, instance.Name)
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Create saves and mods subdirectories
	os.MkdirAll(filepath.Join(instanceDir, "saves"), 0755)
	os.MkdirAll(filepath.Join(instanceDir, "mods"), 0755)

	return nil
}

// DeleteInstance removes the env file (does not delete data directory)
func (m *Manager) DeleteInstance(name string) error {
	envPath := filepath.Join(m.baseDir, "env-"+name)
	
	if err := os.Remove(envPath); err != nil {
		return fmt.Errorf("failed to remove env file: %w", err)
	}
	
	return nil
}
