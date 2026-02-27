#!/bin/bash

# Factorio Headless Server Launcher
# Reads configuration from an env file to download and run Factorio headless server

set -e

# Check for env file argument
if [ -z "$1" ]; then
    echo "Usage: $0 <env_file> [additional factorio arguments]"
    echo "  <env_file> - Path to environment file containing NAME, VERSION, and optional TITLE variables"
    exit 1
fi

ENV_FILE="$1"
shift  # Remove env file from arguments, remaining args will be passed to factorio

# Check if env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: Env file '$ENV_FILE' not found"
    exit 1
fi

CREDS_FILE=.env
# Check if creds file exists
if [ ! -f "$CREDS_FILE" ]; then
    echo "Error: credential file '$CREDS_FILE' not found"
    exit 1
fi

# Source the env files
source "$ENV_FILE"
source $CREDS_FILE

# Validate required variables
if [ -z "$NAME" ]; then
    echo "Error: NAME variable not set in env file"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo "Error: VERSION variable not set in env file"
    exit 1
fi

# TITLE is optional, default to "Factorio Server"
if [ -z "$TITLE" ]; then
    TITLE="Factorio Server"
fi

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set up the working directory (NAME is a subdirectory relative to script location)
WORK_DIR="$SCRIPT_DIR/$NAME"

if [ "$VERSION" = "latest" ]; then
    # For latest, we need to fetch the actual version number
    echo "Fetching latest version number..."
    LATEST_VERSION=$(curl -s https://www.factorio.com/api/latest-releases | jq -r '.stable.headless' || echo "")
    if [ -n "$LATEST_VERSION" ]; then
        COMPARE_VERSION="$LATEST_VERSION"
        VERSION="$LATEST_VERSION"  # Always update VERSION to actual version
    fi
fi

echo "=== Factorio Headless Server Launcher ==="
echo "Name: $NAME"
echo "Version: $VERSION"
echo "Working directory: $WORK_DIR"
echo "=========================================="

# Function to get current installed version
get_current_version() {
    if [ -f "$WORK_DIR/bin/x64/factorio" ]; then
        "$WORK_DIR/bin/x64/factorio" --version 2>/dev/null | head -n 1 | awk '{print $2}' || echo ""
    else
        echo ""
    fi
}

# Check if NAME directory already exists
if [ -d "$WORK_DIR" ]; then
    echo "Found existing installation at: $WORK_DIR"
    
    CURRENT_VERSION=$(get_current_version)
    
    if [ -n "$CURRENT_VERSION" ]; then
        echo "Current installed version: $CURRENT_VERSION"
        echo "Requested version: $VERSION"
        
        # Determine the version string to compare
        COMPARE_VERSION="$VERSION"

        # Compare versions
        if [ "$CURRENT_VERSION" = "$COMPARE_VERSION" ]; then
            echo "Version matches - no download needed"
        else
            echo "Version differs - will re-download"
            rm -rf "$WORK_DIR/bin" "$WORK_DIR/data" 2>/dev/null || true
        fi
    else
        echo "No valid Factorio installation found, will download"
    fi
else
    echo "Creating new installation directory: $WORK_DIR"
    mkdir -p "$WORK_DIR"
fi

# Download Factorio if needed
if [ ! -f "$WORK_DIR/bin/x64/factorio" ]; then
    echo "Downloading Factorio headless server..."
    
    # Create saves directory
    mkdir -p "$WORK_DIR/saves"
    
    DOWNLOAD_URL="https://www.factorio.com/get-download/$VERSION/headless/linux64"
    echo "Downloading version $VERSION..."
    
    # Download the archive
    # Note: This requires the user to have a Factorio account and proper authentication
    # For authenticated downloads, you may need to use curl with session cookies or wget with auth
    TEMP_ARCHIVE="$WORK_DIR/factorio_download.tar.xz"
    
    # Try direct download (works if IP is whitelisted on factorio.com)
    if curl -L -o "$TEMP_ARCHIVE" "$DOWNLOAD_URL" 2>&1 | tail -n 5; then
        echo "Download complete"
    else
        echo "Error: Failed to download Factorio"
        echo "Note: You may need to:"
        echo "  1. Download manually from https://www.factorio.com/download"
        echo "  2. Place the headless archive in $WORK_DIR/"
        echo "  3. Name it factorio_download.tar.xz"
        rm -f "$TEMP_ARCHIVE"
        exit 1
    fi
    
    # Extract the archive to a temporary location
    echo "Extracting..."
    TEMP_EXTRACT="$WORK_DIR/_factorio_extract"
    mkdir -p "$TEMP_EXTRACT"
    tar -xJf "$TEMP_ARCHIVE" -C "$TEMP_EXTRACT"
    rm -f "$TEMP_ARCHIVE"
    
    # The archive includes a factorio/ directory at the root - strip it out
    # Move all contents from the factorio/ subdirectory directly to WORK_DIR
    if [ -d "$TEMP_EXTRACT/factorio" ]; then
        # Move everything inside factorio/ to WORK_DIR
        mv "$TEMP_EXTRACT/factorio/"* "$WORK_DIR/" 2>/dev/null || true
        mv "$TEMP_EXTRACT/factorio/."* "$WORK_DIR/" 2>/dev/null || true
        rmdir "$TEMP_EXTRACT/factorio" 2>/dev/null || true
    else
        # Fallback: if no factorio/ subdirectory, move everything from temp to work_dir
        mv "$TEMP_EXTRACT/"* "$WORK_DIR/" 2>/dev/null || true
    fi
    rm -rf "$TEMP_EXTRACT"
    
    # Verify installation
    if [ -f "$WORK_DIR/bin/x64/factorio" ]; then
        INSTALLED_VERSION=$(get_current_version)
        echo "Successfully installed Factorio $INSTALLED_VERSION"
    else
        echo "Error: Installation verification failed"
        exit 1
    fi
fi

# Copy and configure server-settings.json
echo "Configuring server settings..."

# Find the server-settings.json in templates/ directory
if [ -f "$SCRIPT_DIR/templates/server-settings.json" ]; then
    SETTINGS_SOURCE="$SCRIPT_DIR/templates/server-settings.json"
else
    echo "Warning: templates/server-settings.json not found, using Factorio defaults"
    SETTINGS_SOURCE=""
fi

if [ -n "$SETTINGS_SOURCE" ] && [ -d "$WORK_DIR/data" ]; then
    # Use awk with -v for safe variable handling (handles special characters)
    awk -v title="$TITLE" \
        -v username="$FACTORIO_USERNAME" \
        -v token="$FACTORIO_TOKEN" \
        -v non_blocking="$NON_BLOCKING_SAVE" \
        -v description="$DESCRIPTION" '
        {
            gsub(/\r/, "")
            gsub(/TITLE_REPLACE/, title)
            gsub(/USERNAME_REPLACE/, username)
            gsub(/TOKEN_REPLACE/, token)
            gsub(/NON_BLOCKING_SAVE_REPLACE/, non_blocking)
            gsub(/DESCRIPTION_REPLACE/, description)
            print
        }
        ' "$SETTINGS_SOURCE" > "$WORK_DIR/data/server-settings.json"
    echo "Server settings copied and configured"
elif [ -n "$SETTINGS_SOURCE" ]; then
    echo "Warning: Factorio data directory not found, skipping server settings"
fi

# Save the version for future reference
INSTALLED_VERSION=$(get_current_version)
echo "$INSTALLED_VERSION" > "$WORK_DIR/.version"

EXE=bin/x64/factorio

echo ""
echo "=== Starting Factorio Server in $WORK_DIR ==="
echo "Binary: $EXE"
echo "Version: $INSTALLED_VERSION"
echo "================================="

cd $WORK_DIR

mkdir -p config
echo "Setting PORT=$PORT"
sed -e "s/PORT_REPLACE/$PORT/" \
    ../templates/config.ini > config/config.ini

if [ -f mod-list.json ]; then
    echo "mod-list.json found, executing python mod downloader script"
    python3.12 ../download_mods.py
fi

if [ -f mod-settings.dat ]; then
    cp mod-settings.dat mods/
fi

SAVE_FILE="saves/main.zip"
if [ -f main.zip ]; then
    mv main.zip saves/
elif [ ! -f "$SAVE_FILE" ]; then
    echo "Save file not found, creating new world..."
    $EXE --create $SAVE_FILE
fi

# Launch Factorio headless server
# Additional arguments can be passed via command line or environment
exec "$WORK_DIR/bin/x64/factorio" \
    --start-server "$SAVE_FILE" \
    --server-settings data/server-settings.json \
    "$@"
