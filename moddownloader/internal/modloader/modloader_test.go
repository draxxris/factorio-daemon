package modloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseModFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantName string
		wantVer  string
		wantOk   bool
	}{
		{
			name:     "valid mod filename",
			filename: "example-mod_1.0.0.zip",
			wantName: "example-mod",
			wantVer:  "1.0.0",
			wantOk:   true,
		},
		{
			name:     "valid mod filename with complex name",
			filename: "factorio-mod-with-dashes_2.3.4.zip",
			wantName: "factorio-mod-with-dashes",
			wantVer:  "2.3.4",
			wantOk:   true,
		},
		{
			name:     "missing zip extension",
			filename: "example-mod_1.0.0",
			wantName: "",
			wantVer:  "",
			wantOk:   false,
		},
		{
			name:     "no underscore",
			filename: "examplemod.zip",
			wantName: "",
			wantVer:  "",
			wantOk:   false,
		},
		{
			name:     "empty name",
			filename: "_1.0.0.zip",
			wantName: "",
			wantVer:  "",
			wantOk:   false,
		},
		{
			name:     "empty version",
			filename: "example-mod_.zip",
			wantName: "",
			wantVer:  "",
			wantOk:   false,
		},
		{
			name:     "multiple underscores",
			filename: "example_mod_name_1.0.0.zip",
			wantName: "example_mod_name",
			wantVer:  "1.0.0",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVer, gotOk := ParseModFilename(tt.filename)
			if gotName != tt.wantName || gotVer != tt.wantVer || gotOk != tt.wantOk {
				t.Errorf("ParseModFilename() = (%v, %v, %v), want (%v, %v, %v)",
					gotName, gotVer, gotOk, tt.wantName, tt.wantVer, tt.wantOk)
			}
		})
	}
}

func TestLoadModList(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "mod-list.json")

	testContent := `{
		"mods": [
			{"name": "base", "enabled": true},
			{"name": "example-mod", "enabled": true},
			{"name": "disabled-mod", "enabled": false}
		]
	}`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	d := NewDownloader("testuser", "testtoken")
	modList, err := d.LoadModList(testFile)
	if err != nil {
		t.Fatalf("LoadModList() error = %v", err)
	}

	if len(modList.Mods) != 3 {
		t.Errorf("expected 3 mods, got %d", len(modList.Mods))
	}

	if modList.Mods[0].Name != "base" {
		t.Errorf("expected first mod name 'base', got %s", modList.Mods[0].Name)
	}
}

func TestLoadModList_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "mod-list.json")

	testContent := `{"mods": [invalid json]}`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	d := NewDownloader("testuser", "testtoken")
	_, err := d.LoadModList(testFile)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadModList_FileNotFound(t *testing.T) {
	d := NewDownloader("testuser", "testtoken")
	_, err := d.LoadModList("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestGetEnabledMods(t *testing.T) {
	modList := &ModList{
		Mods: []Mod{
			{Name: "base", Enabled: true},
			{Name: "example-mod", Enabled: true},
			{Name: "disabled-mod", Enabled: false},
			{Name: "another-mod", Enabled: true},
			{Name: "quality", Enabled: true},
		},
	}

	d := NewDownloader("testuser", "testtoken")
	enabledMods := d.GetEnabledMods(modList)

	expected := []string{"example-mod", "another-mod"}
	if len(enabledMods) != len(expected) {
		t.Errorf("expected %d enabled mods, got %d", len(expected), len(enabledMods))
	}

	for i, mod := range enabledMods {
		if mod != expected[i] {
			t.Errorf("expected mod %s at index %d, got %s", expected[i], i, mod)
		}
	}
}

func TestGetEnabledModNames(t *testing.T) {
	modList := &ModList{
		Mods: []Mod{
			{Name: "base", Enabled: true},
			{Name: "example-mod", Enabled: true},
			{Name: "disabled-mod", Enabled: false},
		},
	}

	d := NewDownloader("testuser", "testtoken")
	enabledNames := d.GetEnabledModNames(modList)

	if !enabledNames["base"] {
		t.Error("expected 'base' to be enabled")
	}
	if !enabledNames["example-mod"] {
		t.Error("expected 'example-mod' to be enabled")
	}
	if enabledNames["disabled-mod"] {
		t.Error("expected 'disabled-mod' to be disabled")
	}
}

func TestGetLatestRelease(t *testing.T) {
	tests := []struct {
		name     string
		modInfo  *ModInfo
		wantVer  string
		wantNil  bool
	}{
		{
			name: "multiple releases",
			modInfo: &ModInfo{
				Name: "test-mod",
				Releases: []Release{
					{Version: "1.0.0"},
					{Version: "1.1.0"},
					{Version: "1.2.0"},
				},
			},
			wantVer: "1.2.0",
			wantNil: false,
		},
		{
			name: "single release",
			modInfo: &ModInfo{
				Name: "test-mod",
				Releases: []Release{
					{Version: "1.0.0"},
				},
			},
			wantVer: "1.0.0",
			wantNil: false,
		},
		{
			name:     "no releases",
			modInfo:  &ModInfo{Name: "test-mod", Releases: []Release{}},
			wantVer:  "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDownloader("testuser", "testtoken")
			latest := d.GetLatestRelease(tt.modInfo)

			if tt.wantNil {
				if latest != nil {
					t.Errorf("expected nil, got version %s", latest.Version)
				}
			} else {
				if latest == nil {
					t.Error("expected non-nil release, got nil")
				} else if latest.Version != tt.wantVer {
					t.Errorf("expected version %s, got %s", tt.wantVer, latest.Version)
				}
			}
		})
	}
}

func TestCleanupOldMods(t *testing.T) {
	tmpDir := t.TempDir()
	oldModsDir := ModsDir
	ModsDir = tmpDir
	defer func() { ModsDir = oldModsDir }()

	modFiles := []string{
		"enabled-mod_1.0.0.zip",
		"enabled-mod_0.9.0.zip",
		"disabled-mod_1.0.0.zip",
		"another-enabled-mod_2.0.0.zip",
	}

	for _, file := range modFiles {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	enabledNames := map[string]bool{
		"enabled-mod":        true,
		"another-enabled-mod": true,
	}
	latestVersions := map[string]string{
		"enabled-mod":        "1.0.0",
		"another-enabled-mod": "2.0.0",
	}

	d := NewDownloader("testuser", "testtoken")
	deleted, kept := d.CleanupOldMods(enabledNames, latestVersions)

	if deleted != 2 {
		t.Errorf("expected 2 deleted files, got %d", deleted)
	}
	if kept != 2 {
		t.Errorf("expected 2 kept files, got %d", kept)
	}

	files, _ := filepath.Glob(filepath.Join(tmpDir, "*.zip"))
	if len(files) != 2 {
		t.Errorf("expected 2 files remaining, got %d", len(files))
	}
}

func TestWriteFilteredModList(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "mod-list.json")

	modList := &ModList{
		Mods: []Mod{
			{Name: "base", Enabled: true},
			{Name: "enabled-mod", Enabled: true},
			{Name: "disabled-mod", Enabled: false},
		},
	}

	d := NewDownloader("testuser", "testtoken")
	err := d.WriteFilteredModList(modList, outputFile)
	if err != nil {
		t.Fatalf("WriteFilteredModList() error = %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result ModList
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if len(result.Mods) != 2 {
		t.Errorf("expected 2 enabled mods in output, got %d", len(result.Mods))
	}

	for _, mod := range result.Mods {
		if !mod.Enabled {
			t.Errorf("expected all mods in filtered list to be enabled, found %s disabled", mod.Name)
		}
	}
}

func TestBaseGameMods(t *testing.T) {
	modList := &ModList{
		Mods: []Mod{
			{Name: "base", Enabled: true},
			{Name: "elevated-rails", Enabled: true},
			{Name: "quality", Enabled: true},
			{Name: "space-age", Enabled: true},
			{Name: "other-mod", Enabled: true},
		},
	}

	d := NewDownloader("testuser", "testtoken")
	enabledMods := d.GetEnabledMods(modList)

	for _, mod := range enabledMods {
		if baseGameMods[mod] {
			t.Errorf("base game mod %s should not be in enabled mods", mod)
		}
	}

	if len(enabledMods) != 1 {
		t.Errorf("expected 1 non-base-game mod, got %d", len(enabledMods))
	}
	if enabledMods[0] != "other-mod" {
		t.Errorf("expected 'other-mod', got %s", enabledMods[0])
	}
}
