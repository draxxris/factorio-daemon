# Factorio Daemon

A complete solution for running and managing multiple Factorio headless server instances on Linux with systemd. The project consists of two main components:

1. **Launcher** - A bash script and systemd service template that automatically downloads Factorio, installs mods from `mod-list.json`, and runs headless servers
2. **Webapp** - A Go-based web application for managing multiple Factorio server instances with a modern UI

## Features

### Launcher
- **Automatic Factorio Installation**: Downloads and installs the specified Factorio version automatically
- **Mod Auto-Download**: Parses `mod-list.json` and downloads enabled mods from the Factorio mod portal
- **Systemd Integration**: Template service file for easy instance management
- **Version Management**: Supports specific versions or "latest" to always run the newest stable release
- **Configuration Templates**: Pre-configured server settings and config files with variable substitution

### Webapp
- **Instance Dashboard**: View all Factorio instances with real-time status indicators
- **Service Control**: Start, stop, restart, enable/disable autostart for any instance
- **File Upload**: Drag-and-drop upload for mod-list.json, mod-settings.dat, and save files
- **Save Backup**: Manual backup and restore of save files
- **Log Viewer**: Real-time log streaming via Server-Sent Events
- **Instance Creation**: Create new instances with customizable settings

## Quick Start

### Prerequisites

- Linux system with systemd
- Factorio account credentials (for mod downloads and public server visibility)
- Python 3.12+ (for mod downloader)
- Go 1.21+ (for building the webapp)
- Root access for installation

### Launcher Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd factorio-daemon
   ```

2. **Install the launcher**:
   ```bash
   cd launcher
   sudo make install
   ```

3. **Configure credentials**:
   ```bash
   cd /opt/factorio
   # Create .env with your Factorio credentials
   echo 'FACTORIO_USERNAME=your_username' > .env
   echo 'FACTORIO_TOKEN=your_token' >> .env
   ```

4. **Create an instance**:
   ```bash
   # Copy the example environment file
   cp env-example env-myserver
   
   # Edit the instance configuration
   vi env-myserver
   ```

   Instance configuration (`env-myserver`):
   ```
   NAME=myserver
   VERSION=latest
   TITLE="[quality=rare] My Server"
   DESCRIPTION="Welcome to my server."
   PORT=34197
   NON_BLOCKING_SAVE=true
   ```

5. **Start your instance**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl start factorio@myserver
   sudo systemctl enable factorio@myserver
   ```

### Webapp Installation

```bash
cd webapp
make build
sudo make install
sudo systemctl daemon-reload
sudo systemctl enable --now factorio-webapp
```

Access the webapp at `http://your-server:8080`

## Usage

### Creating a Modpack

The easiest way to set up a modpack is to use the Factorio client:

1. Install and configure your mods in the Factorio client
2. Locate your `mod-list.json` and `mod-settings.dat` from the Factorio mods directory (`%APPDATA%\Factorio\mods`)
3. Upload these files via the webapp to your instance's staging area
4. Deploy the files - the launcher will automatically download all enabled mods

### Managing Instances via Webapp

1. Open http://localhost:8080 in your browser
2. Click on an instance card to manage it
3. Upload files via drag-and-drop or click
4. Stop the instance before deploying files
5. Click "Deploy Files" to move staged files to the instance directory
6. Start the instance - factorio.sh will move files to proper locations

The webapp provides a clean interface for:

- **Viewing Instances**: See all instances with their current status (running, stopped, etc.)
- **Controlling Services**: Start, stop, restart, enable, or disable instances
- **Uploading Files**: Drag and drop `mod-list.json`, `mod-settings.dat`, or save files
- **Deploying Files**: Move staged files to the instance directory
- **Viewing Logs**: Stream real-time logs from any instance
- **Backup/Restore**: Create backups of save files and restore them later

### File Deployment Flow

1. Upload files to staging area
2. Stop the instance (required)
3. Click "Deploy Files" - files are moved to `/opt/factorio/{instance}/`
4. Start the instance
5. `factorio.sh` moves files to proper locations:
   - `mod-list.json` → `mods/mod-list.json`
   - `mod-settings.dat` → `mods/mod-settings.dat`
   - `main.zip` → `saves/main.zip`

### Managing Instances via CLI

```bash
# Start an instance
sudo systemctl start factorio@myserver

# Stop an instance
sudo systemctl stop factorio@myserver

# Restart an instance
sudo systemctl restart factorio@myserver

# Enable autostart
sudo systemctl enable factorio@myserver

# Disable autostart
sudo systemctl disable factorio@myserver

# View logs
sudo journalctl -u factorio@myserver -f
```

## Configuration

### Launcher Environment Variables

Each instance has an `env-{name}` file with the following variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NAME` | Yes | - | Instance name (used for directory and systemd service) |
| `VERSION` | Yes | - | Factorio version or "latest" for newest stable |
| `TITLE` | No | "Factorio Server" | Server display name |
| `DESCRIPTION` | No | - | Server description |
| `PORT` | No | 34197 | Game port |
| `NON_BLOCKING_SAVE` | No | true | Enable non-blocking save |

### Webapp Environment Variables

These are typically pulled from the systemd unit file:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 8080 | HTTP server port |
| `SERVER_HOST` | 0.0.0.0 | HTTP server host |
| `FACTORIO_BASE_DIR` | /opt/factorio | Base directory for Factorio |
| `STAGING_DIR` | /opt/factorio/webapp/data/staging | Staging directory for uploads |
| `BACKUP_DIR` | /opt/factorio/webapp/data/backups | Backup directory for saves |
| `LOG_POLL_INTERVAL` | 2 | Log polling interval in seconds |
| `LOG_MAX_LINES` | 1000 | Maximum log lines to retrieve |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/instances` | List all instances with status |
| GET | `/api/instances/:name` | Get instance details |
| POST | `/api/instances` | Create new instance |
| DELETE | `/api/instances/:name` | Delete instance |
| POST | `/api/instances/:name/start` | Start service |
| POST | `/api/instances/:name/stop` | Stop service |
| POST | `/api/instances/:name/restart` | Restart service |
| POST | `/api/instances/:name/enable` | Enable autostart |
| POST | `/api/instances/:name/disable` | Disable autostart |
| GET | `/api/instances/:name/logs` | Get recent logs |
| GET | `/api/instances/:name/logs/stream` | Stream logs via SSE |
| POST | `/api/instances/:name/upload` | Upload files to staging |
| GET | `/api/instances/:name/staged` | List staged files |
| DELETE | `/api/instances/:name/staged` | Clear staged files |
| POST | `/api/instances/:name/deploy` | Deploy staged files |
| POST | `/api/instances/:name/backup` | Backup save file |
| GET | `/api/instances/:name/backups` | List backups |
| POST | `/api/instances/:name/backups/:filename/restore` | Restore backup |

## License

AGPL v3 License - See LICENSE.txt file for details
