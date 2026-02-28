package instance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/draxxris/factorio-webapp/internal/testutil"
)

func TestListInstances_EmptyDirectory(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	instances, err := manager.ListInstances()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}

func TestListInstances_MultipleInstances(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	testutil.CreateEnvFile(t, filepath.Join(baseDir, "env-instance1"), "instance1")
	testutil.CreateEnvFile(t, filepath.Join(baseDir, "env-instance2"), "instance2")
	testutil.CreateEnvFile(t, filepath.Join(baseDir, "env-example"), "example")

	instances, err := manager.ListInstances()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestListInstances_InvalidEnvFiles(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	testutil.CreateEnvFile(t, filepath.Join(baseDir, "env-valid"), "valid")

	instances, err := manager.ListInstances()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 1 {
		t.Errorf("expected 1 valid instance, got %d", len(instances))
	}
	if instances[0].Name != "valid" {
		t.Errorf("expected instance name 'valid', got %s", instances[0].Name)
	}
}

func TestGetInstance_ValidInstance(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	testutil.CreateEnvFile(t, filepath.Join(baseDir, "env-test"), "test")

	instance, err := manager.GetInstance("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if instance.Name != "test" {
		t.Errorf("expected name 'test', got %s", instance.Name)
	}
	if instance.Version != "latest" {
		t.Errorf("expected version 'latest', got %s", instance.Version)
	}
	if instance.Title != "Test Instance" {
		t.Errorf("expected title 'Test Instance', got %s", instance.Title)
	}
	if instance.Description != "Test Description" {
		t.Errorf("expected description 'Test Description', got %s", instance.Description)
	}
	if instance.Port != 34197 {
		t.Errorf("expected port 34197, got %d", instance.Port)
	}
	if instance.NonBlockingSave != false {
		t.Errorf("expected NonBlockingSave false, got %t", instance.NonBlockingSave)
	}
}

func TestGetInstance_MissingInstance(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	_, err := manager.GetInstance("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing instance")
	}
	if !strings.Contains(err.Error(), "failed to open env file") {
		t.Errorf("expected env file error, got %v", err)
	}
}

func TestGetInstance_MalformedEnvFile(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	malformedContent := `NAME=test
invalid line without equals
PORT=not-a-number
NON_BLOCKING_SAVE=maybe
`
	testutil.WriteFile(t, filepath.Join(baseDir, "env-test"), malformedContent)

	instance, err := manager.GetInstance("test")
	if err != nil {
		t.Fatalf("expected no error parsing malformed file, got %v", err)
	}
	if instance.Name != "test" {
		t.Errorf("expected name 'test', got %s", instance.Name)
	}
	if instance.Port != 0 {
		t.Errorf("expected port 0 for invalid value, got %d", instance.Port)
	}
	if instance.NonBlockingSave != false {
		t.Errorf("expected NonBlockingSave false for invalid value, got %t", instance.NonBlockingSave)
	}
}

func TestCreateInstance_ValidInstance(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	inst := Instance{
		Name:            "test-instance",
		Version:         "1.2.3",
		Title:           "Test Title",
		Description:     "Test Description",
		Port:            34198,
		NonBlockingSave: true,
	}

	err := manager.CreateInstance(inst)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	envPath := filepath.Join(baseDir, "env-test-instance")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("expected env file to be created")
	}

	content := testutil.ReadFile(t, envPath)
	if !strings.Contains(content, "NAME=test-instance") {
		t.Error("expected NAME in env file")
	}
	if !strings.Contains(content, "VERSION=1.2.3") {
		t.Error("expected VERSION in env file")
	}
	if !strings.Contains(content, "PORT=34198") {
		t.Error("expected PORT in env file")
	}

	instanceDir := filepath.Join(baseDir, "test-instance")
	if _, err := os.Stat(instanceDir); os.IsNotExist(err) {
		t.Fatal("expected instance directory to be created")
	}

	savesDir := filepath.Join(instanceDir, "saves")
	if _, err := os.Stat(savesDir); os.IsNotExist(err) {
		t.Fatal("expected saves directory to be created")
	}

	modsDir := filepath.Join(instanceDir, "mods")
	if _, err := os.Stat(modsDir); os.IsNotExist(err) {
		t.Fatal("expected mods directory to be created")
	}
}

func TestCreateInstance_InvalidName(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	testCases := []string{
		"test instance",
		"test@instance",
		"test/instance",
		"test.instance",
		"",
	}

	for _, name := range testCases {
		inst := Instance{Name: name}
		err := manager.CreateInstance(inst)
		if err == nil {
			t.Errorf("expected error for invalid name '%s'", name)
		}
		if !strings.Contains(err.Error(), "alphanumeric") {
			t.Errorf("expected alphanumeric error for '%s', got %v", name, err)
		}
	}
}

func TestCreateInstance_DuplicateInstance(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	inst := Instance{Name: "duplicate"}

	err := manager.CreateInstance(inst)
	if err != nil {
		t.Fatalf("expected no error on first create, got %v", err)
	}

	err = manager.CreateInstance(inst)
	if err == nil {
		t.Fatal("expected error for duplicate instance")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got %v", err)
	}
}

func TestCreateInstance_Defaults(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	inst := Instance{
		Name: "test-defaults",
	}

	err := manager.CreateInstance(inst)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	instance, err := manager.GetInstance("test-defaults")
	if err != nil {
		t.Fatalf("expected no error getting instance, got %v", err)
	}

	if instance.Name != "test-defaults" {
		t.Errorf("expected name 'test-defaults', got %s", instance.Name)
	}
	if instance.Version != "" {
		t.Errorf("expected empty version, got %s", instance.Version)
	}
	if instance.Port != 0 {
		t.Errorf("expected port 0 (not set), got %d", instance.Port)
	}
	if instance.NonBlockingSave != false {
		t.Errorf("expected NonBlockingSave false, got %t", instance.NonBlockingSave)
	}
}

func TestDeleteInstance_ValidInstance(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	inst := Instance{Name: "to-delete"}
	err := manager.CreateInstance(inst)
	if err != nil {
		t.Fatalf("expected no error creating instance, got %v", err)
	}

	envPath := filepath.Join(baseDir, "env-to-delete")
	instanceDir := filepath.Join(baseDir, "to-delete")

	err = manager.DeleteInstance("to-delete")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		t.Fatal("expected env file to be deleted")
	}

	if _, err := os.Stat(instanceDir); os.IsNotExist(err) {
		t.Fatal("expected instance directory to NOT be deleted")
	}
}

func TestDeleteInstance_MissingInstance(t *testing.T) {
	baseDir := testutil.TempDir(t)
	manager := NewManager(baseDir)

	err := manager.DeleteInstance("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing instance")
	}
	if !strings.Contains(err.Error(), "failed to remove env file") {
		t.Errorf("expected env file removal error, got %v", err)
	}
}
