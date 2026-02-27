package rcon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorcon/rcon"
)

// Client wraps an RCON connection
type Client struct {
	conn     *rcon.Conn
	instance string
	baseDir  string
}

// NewClient creates a new RCON client for an instance
func NewClient(baseDir, instanceName string) (*Client, error) {
	instanceDir := filepath.Join(baseDir, instanceName)

	// Read RCON port
	portPath := filepath.Join(instanceDir, "rcon-port")
	portData, err := os.ReadFile(portPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rcon-port file: %w", err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(string(portData)))
	if err != nil {
		return nil, fmt.Errorf("invalid rcon port: %w", err)
	}

	// Read RCON password
	passwdPath := filepath.Join(instanceDir, "rcon-passwd")
	passwdData, err := os.ReadFile(passwdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rcon-passwd file: %w", err)
	}
	password := strings.TrimSpace(string(passwdData))
	if password == "" {
		return nil, fmt.Errorf("rcon password is empty")
	}

	// Connect to RCON server with timeout
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := rcon.Dial(address, password, rcon.SetDialTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RCON server: %w", err)
	}

	return &Client{
		conn:     conn,
		instance: instanceName,
		baseDir:  baseDir,
	}, nil
}

// GetServerTime returns current server time
func (c *Client) GetServerTime() (string, error) {
	response, err := c.conn.Execute("/time")
	if err != nil {
		return "", fmt.Errorf("failed to get server time: %w", err)
	}
	return response, nil
}

// GetPlayerList returns list of online players
func (c *Client) GetPlayerList() ([]string, error) {
	response, err := c.conn.Execute("/players online")
	if err != nil {
		return nil, fmt.Errorf("failed to get player list: %w", err)
	}

	// Parse player list from response
	// Format: "Online Players (X):\nplayer1\nplayer2\n..."
	lines := strings.Split(response, "\n")
	var players []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Online") && !strings.HasPrefix(line, "No") {
			players = append(players, line)
		}
	}

	return players, nil
}

// AddAdmin adds a player as admin
func (c *Client) AddAdmin(playerName string) error {
	response, err := c.conn.Execute(fmt.Sprintf("/admin %s", playerName))
	if err != nil {
		return fmt.Errorf("failed to add admin: %w", err)
	}

	// Check for error in response
	if strings.Contains(response, "Error") || strings.Contains(response, "error") {
		return fmt.Errorf("failed to add admin: %s", response)
	}

	return nil
}

// Close closes RCON connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
