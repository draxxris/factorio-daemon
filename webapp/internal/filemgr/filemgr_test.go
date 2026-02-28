package filemgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/draxxris/factorio-webapp/internal/testutil"
)

func TestStageFile_ValidModList(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	data := []byte(`{"mods": []}`)

	err := manager.StageFile("test", "mod-list.json", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stagingPath := filepath.Join(stagingDir, "test", "mod-list.json")
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Fatal("expected mod-list.json to be staged")
	}
}

func TestStageFile_ValidModSettings(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	data := []byte("mod-settings content")

	err := manager.StageFile("test", "mod-settings.dat", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stagingPath := filepath.Join(stagingDir, "test", "mod-settings.dat")
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Fatal("expected mod-settings.dat to be staged")
	}
}

func TestStageFile_ValidZipFile(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	data := []byte("zip content")

	err := manager.StageFile("test", "save.zip", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stagingPath := filepath.Join(stagingDir, "test", "main.zip")
	if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
		t.Fatal("expected main.zip to be staged")
	}
}

func TestStageFile_CaseInsensitiveZip(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	testCases := []string{"save.ZIP", "SAVE.Zip", "SaVe.ZiP"}

	for _, filename := range testCases {
		data := []byte("zip content")
		err := manager.StageFile("test", filename, data)
		if err != nil {
			t.Fatalf("expected no error for %s, got %v", filename, err)
		}

		stagingPath := filepath.Join(stagingDir, "test", "main.zip")
		if _, err := os.Stat(stagingPath); os.IsNotExist(err) {
			t.Fatalf("expected main.zip to be staged for %s", filename)
		}

		os.Remove(stagingPath)
	}
}

func TestStageFile_InvalidFileType(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	data := []byte("content")

	err := manager.StageFile("test", "invalid.txt", data)
	if err == nil {
		t.Fatal("expected error for invalid file type")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected 'not allowed' error, got %v", err)
	}
}

func TestGetStagedFiles_EmptyStaging(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	files, err := manager.GetStagedFiles("nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestGetStagedFiles_MultipleFiles(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	instanceStaging := filepath.Join(stagingDir, "test")
	testutil.CreateDir(t, instanceStaging)
	testutil.WriteFile(t, filepath.Join(instanceStaging, "mod-list.json"), "content1")
	testutil.WriteFile(t, filepath.Join(instanceStaging, "main.zip"), "content2")

	files, err := manager.GetStagedFiles("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestClearStagedFiles_ExistingFiles(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	instanceStaging := filepath.Join(stagingDir, "test")
	testutil.CreateDir(t, instanceStaging)
	testutil.WriteFile(t, filepath.Join(instanceStaging, "file.txt"), "content")

	err := manager.ClearStagedFiles("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(instanceStaging); !os.IsNotExist(err) {
		t.Fatal("expected staging directory to be removed")
	}
}

func TestClearStagedFiles_EmptyStaging(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	err := manager.ClearStagedFiles("nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeployFiles_Success(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	instanceStaging := filepath.Join(stagingDir, "test")
	testutil.CreateDir(t, instanceStaging)
	testutil.WriteFile(t, filepath.Join(instanceStaging, "mod-list.json"), "content")

	err := manager.DeployFiles("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	deployedPath := filepath.Join(baseDir, "test", "mod-list.json")
	if _, err := os.Stat(deployedPath); os.IsNotExist(err) {
		t.Fatal("expected file to be deployed")
	}

	stagedFiles, err := manager.GetStagedFiles("test")
	if err != nil {
		t.Fatalf("expected no error getting staged files, got %v", err)
	}
	if len(stagedFiles) != 0 {
		t.Errorf("expected no staged files after deploy, got %d", len(stagedFiles))
	}
}

func TestDeployFiles_NoStagedFiles(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	err := manager.DeployFiles("test")
	if err == nil {
		t.Fatal("expected error when no files to deploy")
	}
	if !strings.Contains(err.Error(), "no staged files") {
		t.Errorf("expected 'no staged files' error, got %v", err)
	}
}

func TestBackupSave_Success(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	savesDir := filepath.Join(baseDir, "test", "saves")
	testutil.CreateDir(t, savesDir)
	testutil.WriteFile(t, filepath.Join(savesDir, "main.zip"), "save content")

	err := manager.BackupSave("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	backupPath := filepath.Join(backupDir, "test")
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 backup file, got %d", len(entries))
	}

	if !strings.Contains(entries[0].Name(), "_main.zip") {
		t.Errorf("expected backup file to end with _main.zip, got %s", entries[0].Name())
	}
}

func TestBackupSave_NoSaveFile(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	err := manager.BackupSave("test")
	if err == nil {
		t.Fatal("expected error when no save file")
	}
	if !strings.Contains(err.Error(), "no save file found") {
		t.Errorf("expected 'no save file found' error, got %v", err)
	}
}

func TestListBackups_EmptyList(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	backups, err := manager.ListBackups("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

func TestListBackups_MultipleBackups(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	backupPath := filepath.Join(backupDir, "test")
	testutil.CreateDir(t, backupPath)

	oldTime := time.Now().Add(-2 * time.Hour).Format("2006-01-02_150405")
	newTime := time.Now().Add(-1 * time.Hour).Format("2006-01-02_150405")

	testutil.WriteFile(t, filepath.Join(backupPath, oldTime+"_main.zip"), "old content")
	testutil.WriteFile(t, filepath.Join(backupPath, newTime+"_main.zip"), "new content")
	testutil.WriteFile(t, filepath.Join(backupPath, "other.txt"), "other content")

	backups, err := manager.ListBackups("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(backups) != 2 {
		t.Errorf("expected 2 backups, got %d", len(backups))
	}

	if !backups[0].Timestamp.After(backups[1].Timestamp) {
		t.Error("expected backups to be sorted by timestamp descending")
	}
}

func TestRestoreBackup_Success(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	backupPath := filepath.Join(backupDir, "test")
	testutil.CreateDir(t, backupPath)
	testutil.WriteFile(t, filepath.Join(backupPath, "2026-01-01_120000_main.zip"), "backup content")

	err := manager.RestoreBackup("test", "2026-01-01_120000_main.zip")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	savePath := filepath.Join(baseDir, "test", "saves", "main.zip")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatal("expected save file to be restored")
	}

	content := testutil.ReadFile(t, savePath)
	if content != "backup content" {
		t.Errorf("expected backup content, got %s", content)
	}
}

func TestRestoreBackup_MissingBackup(t *testing.T) {
	baseDir := testutil.TempDir(t)
	stagingDir := testutil.TempDir(t)
	backupDir := testutil.TempDir(t)
	manager := NewManager(baseDir, stagingDir, backupDir)

	err := manager.RestoreBackup("test", "nonexistent_backup.zip")
	if err == nil {
		t.Fatal("expected error for missing backup")
	}
	if !strings.Contains(err.Error(), "backup file not found") {
		t.Errorf("expected 'backup file not found' error, got %v", err)
	}
}
