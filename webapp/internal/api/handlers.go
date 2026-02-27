package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/draxxris/factorio-webapp/internal/filemgr"
	"github.com/draxxris/factorio-webapp/internal/instance"
	"github.com/draxxris/factorio-webapp/internal/logstream"
	"github.com/draxxris/factorio-webapp/internal/rcon"
	"github.com/draxxris/factorio-webapp/internal/service"
)

// Handlers holds all API handlers
type Handlers struct {
	instances *instance.Manager
	services  *service.Controller
	files     *filemgr.Manager
	logs      *logstream.Streamer
	baseDir   string
}

// NewHandlers creates a new Handlers instance
func NewHandlers(instances *instance.Manager, services *service.Controller, files *filemgr.Manager, logs *logstream.Streamer, baseDir string) *Handlers {
	return &Handlers{
		instances: instances,
		services:  services,
		files:     files,
		logs:      logs,
		baseDir:   baseDir,
	}
}

// ListInstances returns all instances with their status
func (h *Handlers) ListInstances(c *gin.Context) {
	instances, err := h.instances.ListInstances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get status for each instance
	for i := range instances {
		running, _ := h.services.IsActive(instances[i].Name)
		enabled, _ := h.services.IsEnabled(instances[i].Name)
		instances[i].Running = running
		instances[i].Enabled = enabled
	}

	c.JSON(http.StatusOK, instances)
}

// GetInstance returns details for a specific instance
func (h *Handlers) GetInstance(c *gin.Context) {
	name := c.Param("name")

	inst, err := h.instances.GetInstance(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Get status
	inst.Running, _ = h.services.IsActive(name)
	inst.Enabled, _ = h.services.IsEnabled(name)

	c.JSON(http.StatusOK, inst)
}

// CreateInstanceRequest is the request body for creating an instance
type CreateInstanceRequest struct {
	Name            string `json:"name" binding:"required"`
	Version         string `json:"version"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Port            int    `json:"port"`
	NonBlockingSave bool   `json:"non_blocking_save"`
}

// CreateInstance creates a new instance
func (h *Handlers) CreateInstance(c *gin.Context) {
	var req CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inst := instance.Instance{
		Name:            req.Name,
		Version:         req.Version,
		Title:           req.Title,
		Description:     req.Description,
		Port:            req.Port,
		NonBlockingSave: req.NonBlockingSave,
	}

	// Set defaults
	if inst.Version == "" {
		inst.Version = "latest"
	}
	if inst.Port == 0 {
		inst.Port = 34197
	}

	if err := h.instances.CreateInstance(inst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inst)
}

// DeleteInstance deletes an instance (env file only)
func (h *Handlers) DeleteInstance(c *gin.Context) {
	name := c.Param("name")

	// Check if instance exists
	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Stop and disable first
	h.services.Stop(name)
	h.services.Disable(name)

	if err := h.instances.DeleteInstance(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance deleted"})
}

// StartInstance starts an instance
func (h *Handlers) StartInstance(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.services.Start(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance started"})
}

// StopInstance stops an instance
func (h *Handlers) StopInstance(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.services.Stop(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance stopped"})
}

// RestartInstance restarts an instance
func (h *Handlers) RestartInstance(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.services.Restart(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance restarted"})
}

// EnableInstance enables autostart for an instance
func (h *Handlers) EnableInstance(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.services.Enable(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "autostart enabled"})
}

// DisableInstance disables autostart for an instance
func (h *Handlers) DisableInstance(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.services.Disable(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "autostart disabled"})
}

// GetLogs returns recent logs for an instance
func (h *Handlers) GetLogs(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	lines := 100
	if l := c.Query("lines"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			lines = parsed
		}
	}

	logs, err := h.logs.GetLogs(name, lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// StreamLogs streams logs via Server-Sent Events
func (h *Handlers) StreamLogs(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Get initial logs
	logs, err := h.logs.GetLogs(name, 50)
	if err == nil {
		for _, entry := range logs {
			fmt.Fprintf(c.Writer, "data: {\"timestamp\":\"%s\",\"message\":\"%s\"}\n\n",
				entry.Timestamp.Format(time.RFC3339), entry.Message)
		}
		c.Writer.(http.Flusher).Flush()
	}

	// Stream new logs
	ctx := c.Request.Context()
	ch, err := h.logs.StreamLogsFollow(ctx, name)
	if err != nil {
		return
	}

	for entry := range ch {
		fmt.Fprintf(c.Writer, "data: {\"timestamp\":\"%s\",\"message\":\"%s\"}\n\n",
			entry.Timestamp.Format(time.RFC3339), entry.Message)
		c.Writer.(http.Flusher).Flush()
	}
}

// UploadFile uploads a file to staging
func (h *Handlers) UploadFile(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	// Read file content
	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	if err := h.files.StageFile(name, header.Filename, data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded", "filename": header.Filename})
}

// GetStagedFiles returns list of staged files
func (h *Handlers) GetStagedFiles(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	files, err := h.files.GetStagedFiles(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// ClearStagedFiles clears all staged files
func (h *Handlers) ClearStagedFiles(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.files.ClearStagedFiles(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "staged files cleared"})
}

// DeployFiles deploys staged files to instance
func (h *Handlers) DeployFiles(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Check if instance is stopped
	running, _ := h.services.IsActive(name)
	if running {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance must be stopped before deploying"})
		return
	}

	if err := h.files.DeployFiles(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "files deployed"})
}

// BackupSave creates a backup of the save file
func (h *Handlers) BackupSave(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if err := h.files.BackupSave(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "save backed up"})
}

// ListBackups returns list of save backups
func (h *Handlers) ListBackups(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	backups, err := h.files.ListBackups(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, backups)
}

// RestoreBackup restores a save from backup
func (h *Handlers) RestoreBackup(c *gin.Context) {
	name := c.Param("name")
	filename := c.Param("filename")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Check if instance is stopped
	running, _ := h.services.IsActive(name)
	if running {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance must be stopped before restoring"})
		return
	}

	if err := h.files.RestoreBackup(name, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "backup restored"})
}

// GetServerTime returns the server time via RCON
func (h *Handlers) GetServerTime(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	client, err := rcon.NewClient(h.baseDir, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer client.Close()

	time, err := client.GetServerTime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"time": time})
}

// GetPlayerList returns the list of online players via RCON
func (h *Handlers) GetPlayerList(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	client, err := rcon.NewClient(h.baseDir, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer client.Close()

	players, err := client.GetPlayerList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"players": players})
}

// AddAdminRequest is the request body for adding an admin
type AddAdminRequest struct {
	Player string `json:"player" binding:"required"`
}

// AddAdmin adds a player as admin via RCON
func (h *Handlers) AddAdmin(c *gin.Context) {
	name := c.Param("name")

	if _, err := h.instances.GetInstance(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	var req AddAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := rcon.NewClient(h.baseDir, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer client.Close()

	if err := client.AddAdmin(req.Player); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "admin added", "player": req.Player})
}
