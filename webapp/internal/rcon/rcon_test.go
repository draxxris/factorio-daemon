package rcon

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/draxxris/factorio-webapp/internal/testutil"
)

func createTestInstanceDir(t *testing.T, baseDir, instanceName string, port int, password string) {
	instanceDir := filepath.Join(baseDir, instanceName)
	testutil.CreateDir(t, instanceDir)
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-port"), fmt.Sprintf("%d", port))
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-passwd"), password)
}

func TestNewClient_ValidConfiguration(t *testing.T) {
	baseDir := testutil.TempDir(t)
	createTestInstanceDir(t, baseDir, "test", 27015, "testpass")

	client, err := NewClient(baseDir, "test")
	if err == nil {
		client.Close()
		t.Fatal("expected error (connection failure) in test environment")
	}
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("expected connection error, got %v", err)
	}
}

func TestNewClient_MissingPortFile(t *testing.T) {
	baseDir := testutil.TempDir(t)
	instanceDir := filepath.Join(baseDir, "test")
	testutil.CreateDir(t, instanceDir)
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-passwd"), "testpass")

	_, err := NewClient(baseDir, "test")
	if err == nil {
		t.Fatal("expected error for missing port file")
	}
	if !strings.Contains(err.Error(), "rcon-port") {
		t.Errorf("expected rcon-port error, got %v", err)
	}
}

func TestNewClient_InvalidPort(t *testing.T) {
	baseDir := testutil.TempDir(t)
	instanceDir := filepath.Join(baseDir, "test")
	testutil.CreateDir(t, instanceDir)
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-port"), "invalid")
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-passwd"), "testpass")

	_, err := NewClient(baseDir, "test")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if !strings.Contains(err.Error(), "invalid rcon port") {
		t.Errorf("expected invalid port error, got %v", err)
	}
}

func TestNewClient_MissingPasswordFile(t *testing.T) {
	baseDir := testutil.TempDir(t)
	instanceDir := filepath.Join(baseDir, "test")
	testutil.CreateDir(t, instanceDir)
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-port"), "27015")

	_, err := NewClient(baseDir, "test")
	if err == nil {
		t.Fatal("expected error for missing password file")
	}
	if !strings.Contains(err.Error(), "rcon-passwd") {
		t.Errorf("expected rcon-passwd error, got %v", err)
	}
}

func TestNewClient_EmptyPassword(t *testing.T) {
	baseDir := testutil.TempDir(t)
	instanceDir := filepath.Join(baseDir, "test")
	testutil.CreateDir(t, instanceDir)
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-port"), "27015")
	testutil.WriteFile(t, filepath.Join(instanceDir, "rcon-passwd"), "")

	_, err := NewClient(baseDir, "test")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected empty password error, got %v", err)
	}
}

func TestGetServerTime_Success(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestGetServerTime_RCONError(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestGetPlayerList_Success(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestGetPlayerList_EmptyList(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestGetPlayerList_RCONError(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestAddAdmin_Success(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestAddAdmin_RCONError(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestAddAdmin_ErrorInResponse(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestAddAdmin_LowercaseErrorInResponse(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestClose_Success(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}

func TestClose_AlreadyClosed(t *testing.T) {
	client := &Client{
		conn:     nil,
		instance: "test",
		baseDir:  "/tmp",
	}

	err := client.Close()
	if err != nil {
		t.Errorf("expected no error for nil connection, got %v", err)
	}
}

func TestClient_ConnectionTimeout(t *testing.T) {
	baseDir := testutil.TempDir(t)
	createTestInstanceDir(t, baseDir, "test", 27015, "testpass")

	startTime := time.Now()
	_, err := NewClient(baseDir, "test")
	duration := time.Since(startTime)

	if err == nil {
		t.Fatal("expected error (connection failure)")
	}

	if duration > 6*time.Second {
		t.Errorf("expected connection timeout within 5 seconds, took %v", duration)
	}
}

func TestGetPlayerList_ParsesCorrectly(t *testing.T) {
	t.Skip("requires RCON connection (integration test)")
}
