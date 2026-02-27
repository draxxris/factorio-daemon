#!/usr/bin/env python3
"""
Script to download Factorio mods from mod-list.json.
Parses the mod list, queries the Factorio mods API, and downloads enabled mods.
"""

import json
import os
import sys
import urllib.request
import urllib.error
from pathlib import Path

# Configuration
MOD_LIST_FILE = "mod-list.json"
MODS_DIR = "mods"
API_BASE_URL = "https://mods.factorio.com/api/mods"
DOWNLOAD_BASE_URL = "https://mods.factorio.com"

# Base game mods that are provided with Factorio and should not be downloaded
BASE_GAME_MODS = {"base", "elevated-rails", "quality", "space-age"}

# Factorio credentials for downloading mods
# These can be set via environment variables or passed as arguments
USERNAME = os.environ.get("FACTORIO_USERNAME", "")
TOKEN = os.environ.get("FACTORIO_TOKEN", "")


def load_mod_list(filepath: str) -> dict:
    """Load and parse the mod-list.json file."""
    with open(filepath, 'r') as f:
        return json.load(f)


def get_enabled_mods(mod_list: dict) -> list[str]:
    """Extract enabled mods, excluding base game mods."""
    mods = []
    for mod in mod_list.get("mods", []):
        mod_name = mod.get("name", "")
        if mod.get("enabled", False) and mod_name not in BASE_GAME_MODS:
            mods.append(mod_name)
    return mods


def get_enabled_mod_names(mod_list: dict) -> set[str]:
    """Extract set of enabled mod names (including base game mods)."""
    return {
        mod.get("name", "")
        for mod in mod_list.get("mods", [])
        if mod.get("enabled", False)
    }


def parse_mod_filename(filename: str) -> tuple[str, str] | None:
    """
    Parse a mod zip filename to extract mod name and version.
    Format is typically: modname_version.zip
    Returns (mod_name, version) or None if not a valid mod zip.
    """
    if not filename.endswith('.zip'):
        return None
    
    # Remove .zip extension
    base = filename[:-4]
    
    # Find the last underscore (separates name from version)
    last_underscore = base.rfind('_')
    if last_underscore == -1:
        return None
    
    mod_name = base[:last_underscore]
    version = base[last_underscore + 1:]
    
    if not mod_name or not version:
        return None
    
    return (mod_name, version)


def cleanup_old_mods(enabled_mod_names: set[str], latest_versions: dict[str, str]) -> None:
    """
    Remove outdated and disabled mod zip files from mods directory.
    
    Args:
        enabled_mod_names: Set of mod names that are enabled
        latest_versions: Dict mapping mod names to their latest version string
    """
    mods_path = Path(MODS_DIR)
    if not mods_path.exists():
        return
    
    deleted_count = 0
    kept_count = 0
    
    for zip_file in mods_path.glob("*.zip"):
        parsed = parse_mod_filename(zip_file.name)
        if not parsed:
            continue
        
        mod_name, version = parsed
        
        # Delete if mod is not enabled
        if mod_name not in enabled_mod_names:
            print(f"  Deleting disabled mod: {zip_file.name}")
            zip_file.unlink()
            deleted_count += 1
            continue
        
        # Delete if outdated (version doesn't match latest)
        latest_version = latest_versions.get(mod_name)
        if latest_version and version != latest_version:
            print(f"  Deleting outdated mod: {zip_file.name} (latest: {latest_version})")
            zip_file.unlink()
            deleted_count += 1
            continue
        
        kept_count += 1
    
    print(f"  Kept {kept_count} current mod files, deleted {deleted_count} outdated/disabled files")


def write_filtered_mod_list(mod_list: dict, output_path: str) -> None:
    """Write a filtered mod-list.json containing only enabled mods."""
    # Filter to only enabled mods
    enabled_mods = [
        mod for mod in mod_list.get("mods", [])
        if mod.get("enabled", False)
    ]
    
    filtered_mod_list = {"mods": enabled_mods}
    
    # Ensure output directory exists
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, 'w') as f:
        json.dump(filtered_mod_list, f, indent=2)
    
    print(f"  Written {len(enabled_mods)} enabled mods to {output_path}")


def fetch_mod_info(mod_name: str) -> dict | None:
    """Fetch mod information from the Factorio mods API."""
    url = f"{API_BASE_URL}/{mod_name}/full"
    try:
        with urllib.request.urlopen(url, timeout=30) as response:
            return json.loads(response.read().decode('utf-8'))
    except urllib.error.HTTPError as e:
        print(f"  Error: HTTP {e.code} - {e.reason}")
        return None
    except urllib.error.URLError as e:
        print(f"  Error: Failed to connect - {e.reason}")
        return None
    except json.JSONDecodeError as e:
        print(f"  Error: Failed to parse JSON response - {e}")
        return None


def get_latest_release(mod_info: dict) -> dict | None:
    """Get the latest release from mod info."""
    releases = mod_info.get("releases", [])
    if not releases:
        return None
    
    # Sort by version and get the latest
    # Releases are typically ordered, but let's be safe
    latest = releases[-1]  # Usually the last one is the latest
    return latest


def download_mod(mod_name: str, download_url: str, filename: str, username: str, token: str) -> bool:
    """Download a mod file with authentication."""
    # Construct full download URL with authentication
    separator = "&" if "?" in download_url else "?"
    full_url = f"{DOWNLOAD_BASE_URL}{download_url}{separator}username={username}&token={token}"
    
    output_path = Path(MODS_DIR) / filename
    
    try:
        req = urllib.request.Request(full_url)
        req.add_header('User-Agent', 'FactorioModDownloader/1.0')
        
        with urllib.request.urlopen(req, timeout=120) as response:
            total_size = int(response.headers.get('Content-Length', 0))
            downloaded = 0
            chunk_size = 8192
            
            # Ensure mods directory exists
            output_path.parent.mkdir(parents=True, exist_ok=True)
            
            with open(output_path, 'wb') as f:
                while True:
                    chunk = response.read(chunk_size)
                    if not chunk:
                        break
                    f.write(chunk)
                    downloaded += len(chunk)
                    
                    # Progress indicator for larger files
                    if total_size > 0:
                        percent = (downloaded / total_size) * 100
                        if downloaded % (chunk_size * 10) == 0:  # Update every ~80KB
                            print(f"  Progress: {percent:.1f}%", end='\r')
        
        print(f"  Saved to: {output_path}")
        return True
        
    except urllib.error.HTTPError as e:
        print(f"  Error: HTTP {e.code} - {e.reason}")
        if e.code == 401:
            print("  Hint: Check your username and token")
        elif e.code == 403:
            print("  Hint: You may not have permission to download this mod")
        return False
    except urllib.error.URLError as e:
        print(f"  Error: Failed to download - {e.reason}")
        return False
    except Exception as e:
        print(f"  Error: {e}")
        return False


def main():
    """Main entry point."""
    # Check for credentials
    global USERNAME, TOKEN
    
    if len(sys.argv) >= 3:
        USERNAME = sys.argv[1]
        TOKEN = sys.argv[2]
    
    if not USERNAME or not TOKEN:
        print("Error: Factorio username and token required.")
        print("Usage: python download_mods.py <username> <token>")
        print("Or set FACTORIO_USERNAME and FACTORIO_TOKEN environment variables.")
        sys.exit(1)
    
    # Check if mod-list.json exists
    if not os.path.exists(MOD_LIST_FILE):
        print(f"Error: {MOD_LIST_FILE} not found in current directory.")
        sys.exit(1)
    
    # Create mods directory if it doesn't exist
    Path(MODS_DIR).mkdir(exist_ok=True)
    
    # Load mod list
    print(f"Loading {MOD_LIST_FILE}...")
    mod_list = load_mod_list(MOD_LIST_FILE)
    
    # Get enabled mods
    enabled_mods = get_enabled_mods(mod_list)
    print(f"Found {len(enabled_mods)} enabled mods to download:\n")
    
    # Process each mod and track latest versions
    success_count = 0
    failed_count = 0
    skipped_count = 0
    latest_versions: dict[str, str] = {}  # Track latest version for each mod
    
    for mod_name in enabled_mods:
        #print(f"Processing: {mod_name}")
        
        # Fetch mod info from API
        mod_info = fetch_mod_info(mod_name)
        if not mod_info:
            print(f"  Failed to fetch mod info for {mod_name}")
            failed_count += 1
            continue
        
        # Get latest release
        latest = get_latest_release(mod_info)
        if not latest:
            print(f"  No releases found for {mod_name}")
            skipped_count += 1
            continue
        
        version = latest.get("version", "unknown")
        download_url = latest.get("download_url", "")
        filename = latest.get("file_name", f"{mod_name}_{version}.zip")
        
        # Track the latest version for cleanup
        latest_versions[mod_name] = version
        
        #print(f"  Latest version: {version}")
        #print(f"  Download URL: {download_url}")
        
        # Check if already downloaded
        output_path = Path(MODS_DIR) / filename
        if output_path.exists():
            #print(f"  Already exists, skipping: {output_path}")
            skipped_count += 1
            continue
        
        # Download the mod
        if download_mod(mod_name, download_url, filename, USERNAME, TOKEN):
            success_count += 1
        else:
            failed_count += 1
        
        #print()  # Blank line between mods
    
    # Clean up outdated and disabled mod files
    print("Cleaning up old mod files...")
    enabled_mod_names = get_enabled_mod_names(mod_list)
    cleanup_old_mods(enabled_mod_names, latest_versions)
    print()
    
    # Write filtered mod-list.json to mods directory
    print("Writing filtered mod-list.json...")
    output_mod_list = Path(MODS_DIR) / "mod-list.json"
    write_filtered_mod_list(mod_list, str(output_mod_list))
    print()
    
    # Summary
    print("=" * 50)
    print(f"Download complete!")
    print(f"  Successfully downloaded: {success_count}")
    print(f"  Already exists (skipped): {skipped_count}")
    print(f"  Failed: {failed_count}")
    
    if failed_count > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()
