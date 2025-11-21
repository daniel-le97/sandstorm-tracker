# Sandstorm Tracker

Sandstorm Tracker is a Go project that tracks kills, playtime, alive time, weapon stats, and match history for Insurgency: Sandstorm servers. It ingests server logs, parses events, and stores statistics in a database for analysis and visualization.

## Features

- Tracks player kills, deaths, and assists
- Records playtime and alive time per player
- Collects weapon usage and stats
- Maintains match history and session data
- Supports multiple servers
- Configurable via YAML/TOML config files

## Requirements

- Insurgency: Sandstorm server(s) with log access
- Go 1.25+ (only if building from source)
- [Task](https://taskfile.dev/#/installation) (optional, for using the taskfile)

## Quick Start

Choose one of the two methods below:

### Option 1: Download Pre-built Binary (Recommended)

1. **Get the binary:**

   - Download from [GitHub Releases](https://github.com/daniel-le97/sandstorm-tracker/releases)
   - Choose the appropriate file for your OS:
     - Windows: `sandstorm-tracker_windows_amd64.zip`
     - Linux: `sandstorm-tracker_linux_amd64.tar.gz`
     - macOS: `sandstorm-tracker_darwin_amd64.tar.gz`

2. **Extract and setup:**

   ```sh
   # Extract the archive
   unzip sandstorm-tracker_windows_amd64.zip
   # or
   tar -xzf sandstorm-tracker_linux_amd64.tar.gz

   cd sandstorm-tracker
   ```

3. **Configure the application** (see [Configuration](#configuration) section below)

4. **Install as a service** (see [Installation as a Service](#installation-as-a-service) section below)

5. **Verify and test** (see [Verification & Testing](#verification--testing) section below)

### Option 2: Build from Source

1. **Clone the repository:**

   ```sh
   git clone https://github.com/daniel-le97/sandstorm-tracker.git
   cd sandstorm-tracker
   ```

2. **Install dependencies:**

   ```sh
   go mod download
   ```

3. **Build the binary:**

   ```sh
   go build -o sandstorm-tracker main.go
   ```

   Or use Task (if installed):

   ```sh
   task build
   ```

4. **Configure the application** (see [Configuration](#configuration) section below)

5. **Install as a service** (see [Installation as a Service](#installation-as-a-service) section below)

6. **Verify and test** (see [Verification & Testing](#verification--testing) section below)

## Configuration

The tracker uses a **3-tier configuration system** (in order of priority):

1. **Environment Variables** (highest priority) - Override config file values
2. **Config File** (YAML/TOML) - Main configuration
3. **Built-in Defaults** (lowest priority) - Sensible fallbacks

### Configuration File

Create `sandstorm-tracker.yml` or `sandstorm-tracker.toml` in the app directory. Example configs are included:

```sh
# Copy the example config
cp sandstorm-tracker.example.yml sandstorm-tracker.yml

# Edit with your settings
nano sandstorm-tracker.yml  # Linux/macOS
# or open in your editor on Windows
```

Example config files:

- `sandstorm-tracker.example.yml`
- `sandstorm-tracker-saw.example.yml`

### SAW Mode (Recommended)

For [sandstorm-admin-wrapper](https://github.com/Joe-Klauza/sandstorm-admin-wrapper) users, the tracker auto-discovers servers:

**Configuration:**

```yaml
sawPath: "/opt/sandstorm-admin-wrapper"

logging:
  level: "info"
  enableServerLogs: true
```

**Via Environment Variable:**

```bash
export SAW_PATH="/opt/sandstorm-admin-wrapper"
# Config file value will be overridden
```

The tracker will:

- Read `{SAW_PATH}/admin-interface/config/server-configs.json`
- Construct log paths: `{SAW_PATH}/sandstorm-server/Insurgency/Saved/Logs/{SERVER_UUID}.log`
- Extract RCON addresses and passwords
- Configure query addresses

### Manual Mode

For standalone servers:

```yaml
servers:
  - name: "Main Server"
    logPath: "/opt/sandstorm/Insurgency/Saved/Logs"
    rconAddress: "127.0.0.1:27015"
    rconPassword: "your_rcon_password"
    queryAddress: "127.0.0.1:27016"
    enabled: true

logging:
  level: "info"
  enableServerLogs: true
```

### Environment Variable Overrides

Use environment variables to override config file values (useful for Docker/cloud deployments):

```bash
# Set SAW path via environment
export SAW_PATH="/path/to/saw"

# Set individual RCON passwords
export RCON_PASSWORD_0="server0_password"
export RCON_PASSWORD_1="server1_password"
```

Or in a `.env` file:

```
SAW_PATH=/opt/sandstorm-admin-wrapper
RCON_PASSWORD_0=server0_password
RCON_PASSWORD_1=server1_password
```

### Security Best Practices

- **Never commit actual config or .env files to version control**
- Use strong, unique RCON passwords
- Secure file permissions:
  - Linux/macOS: `chmod 600 sandstorm-tracker.yml .env`
  - Windows: Set restricted ACLs via `icacls`
- Store sensitive files outside web-accessible directories
- Rotate RCON passwords regularly

## Setup as a Service

After configuring the application, set it up to run automatically:

### Windows

```powershell
# Run as Administrator
.\scripts\setup-service.ps1
```

### Linux

```bash
sudo ./scripts/setup-service.sh
```

Both set up automatic startup and restart on failure (5-minute restart intervals).

## Running the Application

### Start the Service

**Windows:**

```powershell
Start-ScheduledTask -TaskName SandstormTracker
```

**Linux:**

```bash
sudo systemctl start sandstorm-tracker
```

### Manual Testing

To test without installing as a service:

```sh
# Check for updates
./sandstorm-tracker update

# Start the server
./sandstorm-tracker serve
```

### Create Superuser Account (First-Time Only)

Before accessing the PocketBase admin dashboard at `http://localhost:8090/_/`, create a superuser account:

```sh
# Windows
.\sandstorm-tracker.exe superuser create admin@example.com password123

# Linux/macOS
./sandstorm-tracker superuser create admin@example.com password123
```

Replace `admin@example.com` and `password123` with your desired credentials. Use a strong, unique password.

## Usage

- Start your Insurgency: Sandstorm server(s) with logging enabled.
- Run the tracker as described above.
- Stats will be collected and stored in the configured database.
- Access the PocketBase admin dashboard at `http://localhost:8090/_/` to view collected data

## Tools

This project includes several standalone command-line tools in the `tools/` directory:

### Server Manager (`tools/servermgr`)

A standalone tool for managing Insurgency: Sandstorm dedicated server instances using SAW configurations.

**Features:**

- Start/stop individual or all servers
- Check server status and detect stale processes
- Update SteamCMD and game server files
- Process management with PID tracking
- Apply server-specific configuration files

**Quick Start:**

```bash
# Build the tool
cd tools/servermgr
go build -o servermgr.exe

# Set SAW path (optional)
$env:SAW_PATH = "C:\path\to\sandstorm-admin-wrapper"

# Start a server
servermgr start server-1

# Check status
servermgr status

# Stop all servers
servermgr stop --all
```

See [tools/servermgr/README.md](tools/servermgr/README.md) for complete documentation and [tools/servermgr/QUICKREF.md](tools/servermgr/QUICKREF.md) for a quick reference guide.

### Other Tools

- **`tools/a2s-test-simple`**: Simple A2S query protocol testing
- **`tools/rcon-test`**: RCON connection testing
- **`tools/run-server`**: Development server runner

## Development

For developers contributing to or extending the project, see the [Development Guide](DEVELOPMENT.md) for detailed information on:

- Project architecture
- Building and testing
- Adding new features
- Code organization

## License

MIT
