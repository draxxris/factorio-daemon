//go:build integration

package modloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegration_RealDownload(t *testing.T) {
	username := os.Getenv("FACTORIO_USERNAME")
	token := os.Getenv("FACTORIO_TOKEN")

	if username == "" || token == "" {
		t.Skip("FACTORIO_USERNAME and FACTORIO_TOKEN must be set for integration tests")
	}

	tmpDir := t.TempDir()
	oldModsDir := ModsDir
	ModsDir = tmpDir
	defer func() { ModsDir = oldModsDir }()

	d := NewDownloader(username, token)

	modListFile := filepath.Join(tmpDir, "mod-list.json")
	testContent := `{
		"mods": [
			{"name": "base", "enabled": true},
			{"name": "simple-evolution-combinator", "enabled": true}
		]
	}`

	if err := os.WriteFile(modListFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test mod-list.json: %v", err)
	}

	modList, err := d.LoadModList(modListFile)
	if err != nil {
		t.Fatalf("failed to load mod list: %v", err)
	}

	enabledMods := d.GetEnabledMods(modList)
	if len(enabledMods) != 1 {
		t.Errorf("expected 1 enabled mod, got %d", len(enabledMods))
	}

	if enabledMods[0] != "simple-evolution-combinator" {
		t.Errorf("expected mod 'simple-evolution-combinator', got %s", enabledMods[0])
	}

	modInfo, err := d.FetchModInfo("simple-evolution-combinator")
	if err != nil {
		t.Fatalf("failed to fetch mod info: %v", err)
	}

	if modInfo.Name != "simple-evolution-combinator" {
		t.Errorf("expected mod name 'simple-evolution-combinator', got %s", modInfo.Name)
	}

	latest := d.GetLatestRelease(modInfo)
	if latest == nil {
		t.Fatal("expected latest release, got nil")
	}

	if latest.FileName == "" {
		t.Error("expected file name to be set")
	}

	if err := d.DownloadMod("simple-evolution-combinator", latest.DownloadURL, latest.FileName); err != nil {
		t.Fatalf("failed to download mod: %v", err)
	}

	outputPath := filepath.Join(tmpDir, latest.FileName)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("expected mod file to exist at %s", outputPath)
	}

	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat downloaded file: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("expected downloaded file to have non-zero size")
	}

	t.Logf("Successfully downloaded mod: %s (%d bytes)", latest.FileName, fileInfo.Size())
}

func TestIntegration_RealDownloadMultipleMods(t *testing.T) {
	username := os.Getenv("FACTORIO_USERNAME")
	token := os.Getenv("FACTORIO_TOKEN")

	if username == "" || token == "" {
		t.Skip("FACTORIO_USERNAME and FACTORIO_TOKEN must be set for integration tests")
	}

	tmpDir := t.TempDir()
	oldModsDir := ModsDir
	ModsDir = tmpDir
	defer func() { ModsDir = oldModsDir }()

	d := NewDownloader(username, token)

	modListFile := filepath.Join(tmpDir, "mod-list.json")
	testContent := `{
		"mods": [
			{"name": "base", "enabled": true},
			{"name": "simple-evolution-combinator", "enabled": true},
			{"name": "helmod", "enabled": true}
		]
	}`

	if err := os.WriteFile(modListFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test mod-list.json: %v", err)
	}

	modList, err := d.LoadModList(modListFile)
	if err != nil {
		t.Fatalf("failed to load mod list: %v", err)
	}

	success, failed, skipped, latestVersions, err := d.ProcessMods(modList)
	if err != nil {
		t.Fatalf("failed to process mods: %v", err)
	}

	t.Logf("Processed mods: success=%d, failed=%d, skipped=%d", success, failed, skipped)

	if failed > 0 {
		t.Errorf("expected no failed downloads, got %d", failed)
	}

	files, _ := filepath.Glob(filepath.Join(tmpDir, "*.zip"))
	t.Logf("Downloaded %d mod files", len(files))

	if len(files) != 2 {
		t.Errorf("expected 2 downloaded mod files, got %d", len(files))
	}

	for _, file := range files {
		filename := filepath.Base(file)
		modName, _, ok := ParseModFilename(filename)
		if !ok {
			t.Errorf("failed to parse filename: %s", filename)
			continue
		}

		if modName == "simple-evolution-combinator" || modName == "helmod" {
			t.Logf("Verified downloaded mod: %s", filename)
		}
	}

	enabledModNames := d.GetEnabledModNames(modList)
	deleted, kept := d.CleanupOldMods(enabledModNames, latestVersions)
	t.Logf("Cleanup: deleted=%d, kept=%d", deleted, kept)

	if kept != len(files) {
		t.Errorf("expected %d kept files, got %d", len(files), kept)
	}
}
