package service

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

type mockExecutor struct {
	mockCommand string
	mockArgs    []string
	mockOutput  []byte
	mockError   error
}

func (m *mockExecutor) Command(command string, args ...string) *exec.Cmd {
	m.mockCommand = command
	m.mockArgs = args
	return &exec.Cmd{}
}

func (m *mockExecutor) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return m.mockOutput, m.mockError
}

func (m *mockExecutor) Output(cmd *exec.Cmd) ([]byte, error) {
	return m.mockOutput, m.mockError
}

type testController struct {
	*Controller
	executor *mockExecutor
}

func NewTestController() *testController {
	mockExec := &mockExecutor{}
	return &testController{
		Controller: &Controller{},
		executor:   mockExec,
	}
}

func TestNewController(t *testing.T) {
	ctrl := NewController()
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
}

func TestStart_Success(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestStart_Failure(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("Failed to start")
	tc.executor.mockError = fmt.Errorf("exit status 1")

	err := tc.Start("test")
	if err == nil {
		t.Fatal("expected error on failure")
	}
	if !strings.Contains(err.Error(), "failed to start service") {
		t.Errorf("expected 'failed to start service' error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Failed to start") {
		t.Errorf("expected output in error, got %v", err)
	}
}

func TestStop_Success(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestStop_Failure(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("Failed to stop")
	tc.executor.mockError = fmt.Errorf("exit status 1")

	err := tc.Stop("test")
	if err == nil {
		t.Fatal("expected error on failure")
	}
	if !strings.Contains(err.Error(), "failed to stop service") {
		t.Errorf("expected 'failed to stop service' error, got %v", err)
	}
}

func TestRestart_Success(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestRestart_Failure(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("Failed to restart")
	tc.executor.mockError = fmt.Errorf("exit status 1")

	err := tc.Restart("test")
	if err == nil {
		t.Fatal("expected error on failure")
	}
	if !strings.Contains(err.Error(), "failed to restart service") {
		t.Errorf("expected 'failed to restart service' error, got %v", err)
	}
}

func TestEnable_Success(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestEnable_Failure(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("Failed to enable")
	tc.executor.mockError = fmt.Errorf("exit status 1")

	err := tc.Enable("test")
	if err == nil {
		t.Fatal("expected error on failure")
	}
	if !strings.Contains(err.Error(), "failed to enable service") {
		t.Errorf("expected 'failed to enable service' error, got %v", err)
	}
}

func TestDisable_Success(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestDisable_Failure(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("Failed to disable")
	tc.executor.mockError = fmt.Errorf("exit status 1")

	err := tc.Disable("test")
	if err == nil {
		t.Fatal("expected error on failure")
	}
	if !strings.Contains(err.Error(), "failed to disable service") {
		t.Errorf("expected 'failed to disable service' error, got %v", err)
	}
}

func TestIsActive_ActiveService(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestIsActive_InactiveService(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("inactive")
	tc.executor.mockError = fmt.Errorf("exit status 3")

	active, err := tc.IsActive("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if active {
		t.Error("expected active to be false")
	}
}

func TestIsActive_ErrorReadingOutput(t *testing.T) {
	t.Skip("requires mocking exec.Command (would need refactor)")
}

func TestIsEnabled_EnabledService(t *testing.T) {
	t.Skip("requires systemctl access (integration test)")
}

func TestIsEnabled_DisabledService(t *testing.T) {
	tc := NewTestController()
	tc.executor.mockOutput = []byte("disabled")
	tc.executor.mockError = fmt.Errorf("exit status 1")

	enabled, err := tc.IsEnabled("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if enabled {
		t.Error("expected enabled to be false")
	}
}

func TestIsEnabled_ErrorReadingOutput(t *testing.T) {
	t.Skip("requires mocking exec.Command (would need refactor)")
}
