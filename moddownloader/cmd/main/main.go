package main

import (
	"fmt"
	"os"

	"github.com/draxxris/factorio-moddownloader/internal/modloader"
)

func main() {
	username := os.Getenv("FACTORIO_USERNAME")
	token := os.Getenv("FACTORIO_TOKEN")

	if len(os.Args) >= 3 {
		username = os.Args[1]
		token = os.Args[2]
	}

	if username == "" || token == "" {
		fmt.Println("Error: Factorio username and token required.")
		fmt.Println("Usage: moddownloader <username> <token>")
		fmt.Println("Or set FACTORIO_USERNAME and FACTORIO_TOKEN environment variables.")
		os.Exit(1)
	}

	if _, err := os.Stat(modloader.ModListFile); os.IsNotExist(err) {
		fmt.Printf("Error: %s not found in current directory.\n", modloader.ModListFile)
		os.Exit(1)
	}

	if err := os.MkdirAll(modloader.ModsDir, 0755); err != nil {
		fmt.Printf("Error: Failed to create mods directory: %v\n", err)
		os.Exit(1)
	}

	d := modloader.NewDownloader(username, token)

	fmt.Printf("Loading %s...\n", modloader.ModListFile)
	modList, err := d.LoadModList(modloader.ModListFile)
	if err != nil {
		fmt.Printf("Error: Failed to load mod list: %v\n", err)
		os.Exit(1)
	}

	enabledMods := d.GetEnabledMods(modList)
	fmt.Printf("Found %d enabled mods to download:\n\n", len(enabledMods))

	success, failed, skipped, latestVersions, err := d.ProcessMods(modList)
	if err != nil {
		fmt.Printf("Error: Failed to process mods: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Cleaning up old mod files...")
	enabledModNames := d.GetEnabledModNames(modList)
	deleted, kept := d.CleanupOldMods(enabledModNames, latestVersions)
	fmt.Printf("  Kept %d current mod files, deleted %d outdated/disabled files\n\n", kept, deleted)

	fmt.Println("Writing filtered mod-list.json...")
	outputModList := fmt.Sprintf("%s/mod-list.json", modloader.ModsDir)
	if err := d.WriteFilteredModList(modList, outputModList); err != nil {
		fmt.Printf("Error: Failed to write filtered mod list: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Written %d enabled mods to %s\n\n", len(enabledModNames), outputModList)

	fmt.Println("==================================================")
	fmt.Println("Download complete!")
	fmt.Printf("  Successfully downloaded: %d\n", success)
	fmt.Printf("  Already exists (skipped): %d\n", skipped)
	fmt.Printf("  Failed: %d\n", failed)

	if failed > 0 {
		os.Exit(1)
	}
}
