package service

import (
	"fmt"
	"os/exec"
	"strings"
)

// Controller handles systemd service operations
type Controller struct{}

// NewController creates a new service controller
func NewController() *Controller {
	return &Controller{}
}

// Start starts a factorio service
func (c *Controller) Start(name string) error {
	cmd := exec.Command("sudo", "systemctl", "start", "factorio@"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service: %w, output: %s", err, string(output))
	}
	return nil
}

// Stop stops a factorio service
func (c *Controller) Stop(name string) error {
	cmd := exec.Command("sudo", "systemctl", "stop", "factorio@"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service: %w, output: %s", err, string(output))
	}
	return nil
}

// Restart restarts a factorio service
func (c *Controller) Restart(name string) error {
	cmd := exec.Command("sudo", "systemctl", "restart", "factorio@"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service: %w, output: %s", err, string(output))
	}
	return nil
}

// Enable enables autostart for a factorio service
func (c *Controller) Enable(name string) error {
	cmd := exec.Command("sudo", "systemctl", "enable", "factorio@"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable service: %w, output: %s", err, string(output))
	}
	return nil
}

// Disable disables autostart for a factorio service
func (c *Controller) Disable(name string) error {
	cmd := exec.Command("sudo", "systemctl", "disable", "factorio@"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disable service: %w, output: %s", err, string(output))
	}
	return nil
}

// IsActive checks if a service is currently running
func (c *Controller) IsActive(name string) (bool, error) {
	cmd := exec.Command("sudo", "systemctl", "is-active", "factorio@"+name)
	output, err := cmd.Output()
	if err != nil {
		// systemctl is-active returns non-zero if not active
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "active", nil
}

// IsEnabled checks if a service is enabled for autostart
func (c *Controller) IsEnabled(name string) (bool, error) {
	cmd := exec.Command("sudo", "systemctl", "is-enabled", "factorio@"+name)
	output, err := cmd.Output()
	if err != nil {
		// systemctl is-enabled returns non-zero if not enabled
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "enabled", nil
}
