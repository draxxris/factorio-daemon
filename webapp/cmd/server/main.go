package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/draxxris/factorio-webapp/internal/api"
	"github.com/draxxris/factorio-webapp/internal/config"
	"github.com/draxxris/factorio-webapp/internal/filemgr"
	"github.com/draxxris/factorio-webapp/internal/instance"
	"github.com/draxxris/factorio-webapp/internal/logstream"
	"github.com/draxxris/factorio-webapp/internal/service"
)

//go:embed static
var staticFS embed.FS

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize components
	instanceMgr := instance.NewManager(cfg.Factorio.BaseDir)
	serviceCtrl := service.NewController()
	fileMgr := filemgr.NewManager(cfg.Factorio.BaseDir, cfg.Factorio.StagingDir, cfg.Factorio.BackupDir)
	logStreamer := logstream.NewStreamer(cfg.Logs.PollInterval, cfg.Logs.MaxLines)
	handlers := api.NewHandlers(instanceMgr, serviceCtrl, fileMgr, logStreamer, cfg.Factorio.BaseDir)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// API routes
	apiGroup := r.Group("/api")
	{
		// Instance management
		apiGroup.GET("/instances", handlers.ListInstances)
		apiGroup.GET("/instances/:name", handlers.GetInstance)
		apiGroup.POST("/instances", handlers.CreateInstance)
		apiGroup.DELETE("/instances/:name", handlers.DeleteInstance)

		// Service control
		apiGroup.POST("/instances/:name/start", handlers.StartInstance)
		apiGroup.POST("/instances/:name/stop", handlers.StopInstance)
		apiGroup.POST("/instances/:name/restart", handlers.RestartInstance)
		apiGroup.POST("/instances/:name/enable", handlers.EnableInstance)
		apiGroup.POST("/instances/:name/disable", handlers.DisableInstance)

		// Logs
		apiGroup.GET("/instances/:name/logs", handlers.GetLogs)
		apiGroup.GET("/instances/:name/logs/stream", handlers.StreamLogs)

		// File management
		apiGroup.POST("/instances/:name/upload", handlers.UploadFile)
		apiGroup.GET("/instances/:name/staged", handlers.GetStagedFiles)
		apiGroup.DELETE("/instances/:name/staged", handlers.ClearStagedFiles)
		apiGroup.POST("/instances/:name/deploy", handlers.DeployFiles)

		// Backup management
		apiGroup.POST("/instances/:name/backup", handlers.BackupSave)
		apiGroup.GET("/instances/:name/backups", handlers.ListBackups)
		apiGroup.POST("/instances/:name/backups/:filename/restore", handlers.RestoreBackup)

		// RCON management
		apiGroup.GET("/instances/:name/rcon/time", handlers.GetServerTime)
		apiGroup.GET("/instances/:name/rcon/players", handlers.GetPlayerList)
		apiGroup.POST("/instances/:name/rcon/admin", handlers.AddAdmin)
	}

	// Serve static files
	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to load static files: %v", err)
	}
	r.StaticFS("/static", http.FS(staticContent))

	// Serve index.html for root
	r.GET("/", func(c *gin.Context) {
		data, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			c.String(500, "Failed to load index.html")
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
