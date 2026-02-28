package testutil

import (
	"io"
	"os"
	"path/filepath"
)

type T interface {
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Cleanup(func())
}

func TempDir(t T) string {
	dir, err := os.MkdirTemp("", "factorio-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

func WriteFile(t T, path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func ReadFile(t T, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(data)
}

func CopyFile(t T, src, dst string) {
	s, err := os.Open(src)
	if err != nil {
		t.Fatalf("failed to open %s: %v", src, err)
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		t.Fatalf("failed to create %s: %v", dst, err)
	}
	defer d.Close()

	if _, err := io.Copy(d, s); err != nil {
		t.Fatalf("failed to copy %s to %s: %v", src, dst, err)
	}
}

func CreateDir(t T, path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", path, err)
	}
}

func CreateEnvFile(t T, path string, instanceName string) {
	content := `NAME=` + instanceName + `
VERSION=latest
TITLE=Test Instance
DESCRIPTION=Test Description
PORT=34197
NON_BLOCKING_SAVE=false
`
	WriteFile(t, path, content)
}

func CreateRconFiles(t T, instanceDir string, port int, password string) {
	WriteFile(t, filepath.Join(instanceDir, "rcon-port"), string(rune(port)))
	WriteFile(t, filepath.Join(instanceDir, "rcon-passwd"), password)
}
