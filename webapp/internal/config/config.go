package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Factorio FactorioConfig
	Logs     LogsConfig
}

type ServerConfig struct {
	Port int
	Host string
}

type FactorioConfig struct {
	BaseDir    string
	StagingDir string
	BackupDir  string
}

type LogsConfig struct {
	PollInterval int // seconds
	MaxLines     int
}

func Load() *Config {
	baseDir := getEnv("FACTORIO_BASE_DIR", "/opt/factorio")

	return &Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Factorio: FactorioConfig{
			BaseDir:    baseDir,
			StagingDir: getEnv("STAGING_DIR", filepath.Join(baseDir, "webapp/data/staging")),
			BackupDir:  getEnv("BACKUP_DIR", filepath.Join(baseDir, "webapp/data/backups")),
		},
		Logs: LogsConfig{
			PollInterval: getEnvInt("LOG_POLL_INTERVAL", 2),
			MaxLines:     getEnvInt("LOG_MAX_LINES", 1000),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
