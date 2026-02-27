package modloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ModListFile     = "mod-list.json"
	apiBaseURL      = "https://mods.factorio.com/api/mods"
	downloadBaseURL = "https://mods.factorio.com"
)

var ModsDir = "mods"

var baseGameMods = map[string]bool{
	"base":          true,
	"elevated-rails": true,
	"quality":       true,
	"space-age":     true,
}

type ModList struct {
	Mods []Mod `json:"mods"`
}

type Mod struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type ModInfo struct {
	Name     string     `json:"name"`
	Releases []Release  `json:"releases"`
}

type Release struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
}

type Downloader struct {
	Username string
	Token    string
	Client   *http.Client
}

func NewDownloader(username, token string) *Downloader {
	return &Downloader{
		Username: username,
		Token:    token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (d *Downloader) LoadModList(path string) (*ModList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mod list: %w", err)
	}

	var modList ModList
	if err := json.Unmarshal(data, &modList); err != nil {
		return nil, fmt.Errorf("failed to parse mod list: %w", err)
	}

	return &modList, nil
}

func (d *Downloader) GetEnabledMods(modList *ModList) []string {
	var mods []string
	for _, mod := range modList.Mods {
		if mod.Enabled && !baseGameMods[mod.Name] {
			mods = append(mods, mod.Name)
		}
	}
	return mods
}

func (d *Downloader) GetEnabledModNames(modList *ModList) map[string]bool {
	enabled := make(map[string]bool)
	for _, mod := range modList.Mods {
		if mod.Enabled {
			enabled[mod.Name] = true
		}
	}
	return enabled
}

func ParseModFilename(filename string) (modName, version string, ok bool) {
	if !strings.HasSuffix(filename, ".zip") {
		return "", "", false
	}

	base := strings.TrimSuffix(filename, ".zip")

	lastUnderscore := strings.LastIndex(base, "_")
	if lastUnderscore == -1 {
		return "", "", false
	}

	modName = base[:lastUnderscore]
	version = base[lastUnderscore+1:]

	if modName == "" || version == "" {
		return "", "", false
	}

	return modName, version, true
}

func (d *Downloader) FetchModInfo(modName string) (*ModInfo, error) {
	url := fmt.Sprintf("%s/%s/full", apiBaseURL, modName)

	resp, err := d.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mod info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: failed to fetch mod info", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var modInfo ModInfo
	if err := json.Unmarshal(body, &modInfo); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &modInfo, nil
}

func (d *Downloader) GetLatestRelease(modInfo *ModInfo) *Release {
	if len(modInfo.Releases) == 0 {
		return nil
	}
	return &modInfo.Releases[len(modInfo.Releases)-1]
}

func (d *Downloader) DownloadMod(modName, downloadURL, filename string) error {
	separator := "&"
	if strings.Contains(downloadURL, "?") {
		separator = "&"
	} else {
		separator = "?"
	}

	fullURL := fmt.Sprintf("%s%s%susername=%s&token=%s",
		downloadBaseURL, downloadURL, separator, d.Username, d.Token)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "FactorioModDownloader/1.0")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: failed to download mod", resp.StatusCode)
	}

	if err := os.MkdirAll(ModsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mods directory: %w", err)
	}

	outputPath := filepath.Join(ModsDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (d *Downloader) CleanupOldMods(enabledModNames map[string]bool, latestVersions map[string]string) (deleted, kept int) {
	if _, err := os.Stat(ModsDir); os.IsNotExist(err) {
		return 0, 0
	}

	files, err := filepath.Glob(filepath.Join(ModsDir, "*.zip"))
	if err != nil {
		return 0, 0
	}

	for _, file := range files {
		filename := filepath.Base(file)
		modName, version, ok := ParseModFilename(filename)
		if !ok {
			continue
		}

		if !enabledModNames[modName] {
			os.Remove(file)
			deleted++
			continue
		}

		if latestVersion, exists := latestVersions[modName]; exists && version != latestVersion {
			os.Remove(file)
			deleted++
			continue
		}

		kept++
	}

	return deleted, kept
}

func (d *Downloader) WriteFilteredModList(modList *ModList, outputPath string) error {
	var enabledMods []Mod
	for _, mod := range modList.Mods {
		if mod.Enabled {
			enabledMods = append(enabledMods, mod)
		}
	}

	filteredList := ModList{Mods: enabledMods}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(filteredList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (d *Downloader) ProcessMods(modList *ModList) (success, failed, skipped int, latestVersions map[string]string, err error) {
	enabledMods := d.GetEnabledMods(modList)
	latestVersions = make(map[string]string)

	for _, modName := range enabledMods {
		modInfo, err := d.FetchModInfo(modName)
		if err != nil {
			failed++
			continue
		}

		latest := d.GetLatestRelease(modInfo)
		if latest == nil {
			skipped++
			continue
		}

		version := latest.Version
		downloadURL := latest.DownloadURL
		filename := latest.FileName
		if filename == "" {
			filename = fmt.Sprintf("%s_%s.zip", modName, version)
		}

		latestVersions[modName] = version

		outputPath := filepath.Join(ModsDir, filename)
		if _, err := os.Stat(outputPath); err == nil {
			skipped++
			continue
		}

		if err := d.DownloadMod(modName, downloadURL, filename); err != nil {
			failed++
			continue
		}

		success++
	}

	return success, failed, skipped, latestVersions, nil
}
