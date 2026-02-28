package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := Load()

	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected Server.Host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Factorio.BaseDir != "/opt/factorio" {
		t.Errorf("expected Factorio.BaseDir /opt/factorio, got %s", cfg.Factorio.BaseDir)
	}
	expectedStaging := filepath.Join("/opt/factorio", "webapp/data/staging")
	if cfg.Factorio.StagingDir != expectedStaging {
		t.Errorf("expected Factorio.StagingDir %s, got %s", expectedStaging, cfg.Factorio.StagingDir)
	}
	expectedBackup := filepath.Join("/opt/factorio", "webapp/data/backups")
	if cfg.Factorio.BackupDir != expectedBackup {
		t.Errorf("expected Factorio.BackupDir %s, got %s", expectedBackup, cfg.Factorio.BackupDir)
	}
	if cfg.Logs.PollInterval != 2 {
		t.Errorf("expected Logs.PollInterval 2, got %d", cfg.Logs.PollInterval)
	}
	if cfg.Logs.MaxLines != 1000 {
		t.Errorf("expected Logs.MaxLines 1000, got %d", cfg.Logs.MaxLines)
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	os.Clearenv()

	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("FACTORIO_BASE_DIR", "/custom/factorio")
	os.Setenv("STAGING_DIR", "/custom/staging")
	os.Setenv("BACKUP_DIR", "/custom/backups")
	os.Setenv("LOG_POLL_INTERVAL", "5")
	os.Setenv("LOG_MAX_LINES", "500")
	defer os.Clearenv()

	cfg := Load()

	if cfg.Server.Port != 9000 {
		t.Errorf("expected Server.Port 9000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected Server.Host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Factorio.BaseDir != "/custom/factorio" {
		t.Errorf("expected Factorio.BaseDir /custom/factorio, got %s", cfg.Factorio.BaseDir)
	}
	if cfg.Factorio.StagingDir != "/custom/staging" {
		t.Errorf("expected Factorio.StagingDir /custom/staging, got %s", cfg.Factorio.StagingDir)
	}
	if cfg.Factorio.BackupDir != "/custom/backups" {
		t.Errorf("expected Factorio.BackupDir /custom/backups, got %s", cfg.Factorio.BackupDir)
	}
	if cfg.Logs.PollInterval != 5 {
		t.Errorf("expected Logs.PollInterval 5, got %d", cfg.Logs.PollInterval)
	}
	if cfg.Logs.MaxLines != 500 {
		t.Errorf("expected Logs.MaxLines 500, got %d", cfg.Logs.MaxLines)
	}
}

func TestLoad_InvalidEnvironmentValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("SERVER_PORT", "invalid")
	os.Setenv("LOG_POLL_INTERVAL", "not-a-number")
	os.Setenv("LOG_MAX_LINES", "also-invalid")
	defer os.Clearenv()

	cfg := Load()

	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port to default to 8080 on invalid input, got %d", cfg.Server.Port)
	}
	if cfg.Logs.PollInterval != 2 {
		t.Errorf("expected Logs.PollInterval to default to 2 on invalid input, got %d", cfg.Logs.PollInterval)
	}
	if cfg.Logs.MaxLines != 1000 {
		t.Errorf("expected Logs.MaxLines to default to 1000 on invalid input, got %d", cfg.Logs.MaxLines)
	}
}

func TestGetEnv(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	result := getEnv("TEST_VAR", "default")
	if result != "default" {
		t.Errorf("expected default value, got %s", result)
	}

	os.Setenv("TEST_VAR", "custom")
	result = getEnv("TEST_VAR", "default")
	if result != "custom" {
		t.Errorf("expected custom value, got %s", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	result := getEnvInt("TEST_INT", 42)
	if result != 42 {
		t.Errorf("expected default value 42, got %d", result)
	}

	os.Setenv("TEST_INT", "100")
	result = getEnvInt("TEST_INT", 42)
	if result != 100 {
		t.Errorf("expected custom value 100, got %d", result)
	}

	os.Setenv("TEST_INT", "invalid")
	result = getEnvInt("TEST_INT", 42)
	if result != 42 {
		t.Errorf("expected default value 42 on invalid input, got %d", result)
	}
}
