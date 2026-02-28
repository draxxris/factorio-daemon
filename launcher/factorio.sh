#!/bin/bash

# Factorio Headless Server Launcher
# Reads configuration from an env file to download and run Factorio headless server

set -euo pipefail

# Check for env file argument
if [ -z "$1" ]; then
    echo "Usage: $0 <env_file> [additional factorio arguments]"
    echo "  <env_file> - Path to environment file containing NAME, VERSION, and optional TITLE variables"
    exit 1
fi

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")

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
# shellcheck source=/dev/null
source "$ROOT_DIR/$ENV_FILE"
# shellcheck source=/dev/null
source "$ROOT_DIR/$CREDS_FILE"

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

# Set up the working directory (NAME is a subdirectory relative to script location)
WORK_DIR="$(pwd)/$NAME"
TEMPLATE_DIR="$(pwd)/launcher/templates"

# Check if TEMPLATE_DIR exists
if [ ! -d "$TEMPLATE_DIR" ]; then
    echo "Error: Template directory '$TEMPLATE_DIR' not found"
    exit 1
fi

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
            rm -rf "${WORK_DIR:?}/bin" "${WORK_DIR:?}/data" 2>/dev/null || true
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
    if ! curl -L -o "$TEMP_ARCHIVE" "$DOWNLOAD_URL"; then
        echo "Error: Failed to download Factorio"
        echo "Note: You may need to:"
        echo "  1. Download manually from https://www.factorio.com/download"
        echo "  2. Place the headless archive in $WORK_DIR/"
        echo "  3. Name it factorio_download.tar.xz"
        rm -f "$TEMP_ARCHIVE"
        exit 1
    fi
    echo "Download complete"
    
    # Extract the archive to a temporary location
    echo "Extracting..."
    TEMP_EXTRACT="$WORK_DIR/_factorio_extract"
    mkdir -p "$TEMP_EXTRACT"
    if ! tar -xJf "$TEMP_ARCHIVE" -C "$TEMP_EXTRACT"; then
        echo "Error: Failed to extract archive"
        rm -f "$TEMP_ARCHIVE"
        rm -rf "$TEMP_EXTRACT"
        exit 1
    fi
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
if [ -f "$TEMPLATE_DIR/server-settings.json" ]; then
    SETTINGS_SOURCE="$TEMPLATE_DIR/server-settings.json"
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

cd "$WORK_DIR"

# RCON Setup
# Generate or load RCON credentials
RCON_PORT_FILE="rcon-port"
RCON_PASSWD_FILE="rcon-passwd"

# Function to generate random alphanumeric string (16+ characters)
generate_random_password() {
    local length=${1:-16}
    tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$length"
}

# Function to generate random port between 20000-30000
generate_random_port() {
    echo $((20000 + RANDOM % 10001))
}

# Load or generate RCON credentials
if [ -f "$RCON_PORT_FILE" ] && [ -f "$RCON_PASSWD_FILE" ]; then
    echo "Loading existing RCON credentials..."
    RCON_PORT=$(cat "$RCON_PORT_FILE")
    RCON_PASSWORD=$(cat "$RCON_PASSWD_FILE")
else
    echo "Generating new RCON credentials..."
    RCON_PORT=$(generate_random_port)
    RCON_PASSWORD=$(generate_random_password 16)
    echo "$RCON_PORT" > "$RCON_PORT_FILE"
    echo "$RCON_PASSWORD" > "$RCON_PASSWD_FILE"
    chmod 600 "$RCON_PORT_FILE" "$RCON_PASSWD_FILE"
fi

echo "RCON Port: $RCON_PORT"
echo "RCON Password: (stored in $RCON_PASSWD_FILE)"

mkdir -p config

# Check if PORT is set
if [ -z "${PORT:-}" ]; then
    echo "Error: PORT variable not set in env file"
    exit 1
fi

echo "Setting PORT=$PORT"
awk -v port="$PORT" '
    {
        gsub(/\r/, "")
        gsub(/PORT_REPLACE/, port)
        print
    }
    ' "$TEMPLATE_DIR/config.ini" > config/config.ini

if [ ! -f config/config.ini ]; then
    echo "Error: Failed to create config/config.ini"
    exit 1
fi

if [ -f mod-list.json ]; then
    echo "mod-list.json found, executing python mod downloader script"
    if ! python3 ../download_mods.py; then
        echo "Error: Failed to download mods"
        exit 1
    fi
fi

if [ -f mod-settings.dat ]; then
    cp mod-settings.dat mods/
fi

SAVE_FILE="saves/main.zip"
if [ -f main.zip ]; then
    mv main.zip saves/
elif [ ! -f "$SAVE_FILE" ]; then
    echo "Save file not found, creating new world..."
    "$EXE" --create "$SAVE_FILE"
fi

# Launch Factorio headless server
# Additional arguments can be passed via command line or environment
exec "$WORK_DIR/bin/x64/factorio" \
    --start-server "$SAVE_FILE" \
    --server-settings data/server-settings.json \
    --rcon-bind "127.0.0.1:$RCON_PORT" \
    --rcon-password "$RCON_PASSWORD" \
    "$@"
