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

- Go 1.20+
- Insurgency: Sandstorm server(s) with log access
- ### optional
  - [Task](https://taskfile.dev/#/installation) (only if you want to use the taskfile)

## Setup

1. **Download the binary:**

   - Get the latest release from [Releases](https://github.com/daniel-le97/sandstorm-tracker/releases)
   - Or build from source (see Development section)

2. **Create your configuration file:**

   ```sh
   # Copy the example config
   cp sandstorm-tracker.example.yml my-config.yml

   # Edit with your server details
   nano my-config.yml
   ```

   Update the following in your config:

   - Server log paths
   - RCON addresses and passwords
   - Query addresses (optional, for A2S server monitoring)
   - Database path
   - Enable/disable servers as needed

3. **Secure your config file (Linux/Mac):**

   ```sh
   chmod 600 my-config.yml
   ```

   **Windows:**

   ```powershell
   icacls my-config.yml /inheritance:r /grant:r "${env:USERNAME}:(F)"
   ```

4. **Run the tracker:**

   ```sh
   ./sandstorm-tracker --config my-config.yml
   ```

   Or on Windows:

   ```powershell
   .\sandstorm-tracker.exe -c my-config.yml
   ```

## Configuration

The tracker accepts configuration via YAML, TOML, or JSON files. There are two configuration modes:

1. **SAW Mode** (Recommended for sandstorm-admin-wrapper users) - Auto-discovers servers
2. **Manual Mode** (For standalone servers) - Explicit configuration

### For sandstorm-admin-wrapper Users

If you're using [sandstorm-admin-wrapper](https://github.com/Joe-Klauza/sandstorm-admin-wrapper), configuration is simplified! The tracker automatically reads your SAW configuration and constructs log paths for all servers.

**Simple configuration - just provide the SAW path:**

```yaml
sawPath: "/opt/sandstorm-admin-wrapper" # Path to your SAW installation

# That's it! The tracker will:
# - Read server-configs.json automatically
# - Construct log paths for all servers
# - Extract RCON addresses and passwords
# - Configure query addresses (game port + 29)
```

**Example full configuration:**

```yaml
sawPath: "/opt/sandstorm-admin-wrapper"

logging:
  level: "info"
```

The tracker will automatically discover all servers configured in your SAW installation at:

- `{SAW_PATH}/admin-interface/config/server-configs.json`

And construct log paths like:

- `{SAW_PATH}/sandstorm-server/Insurgency/Saved/Logs/{SERVER_UUID}.log`

**Windows users:**

```yaml
sawPath: "C:\\SAW_1.0.4"
```

### For Standalone Server Users

If you're running a single Insurgency: Sandstorm server, you can use the directory path:

```yaml
servers:
  - name: "My Server"
    logPath: "/opt/sandstorm/Insurgency/Saved/Logs" # Directory path
    rconAddress: "127.0.0.1:27015"
    rconPassword: "your_rcon_password"
    queryAddress: "127.0.0.1:27016"
    enabled: true
```

The tracker will automatically monitor all `.log` files in that directory.

### Example Configuration

```yaml
servers:
  - name: "Main Server"
    logPath: "/opt/sandstorm/Insurgency/Saved/Logs"
    rconAddress: "127.0.0.1:27015"
    rconPassword: "your_rcon_password"
    queryAddress: "127.0.0.1:27016" # Optional: for server monitoring
    enabled: true

database:
  path: "sandstorm-tracker.db"
  enableWAL: true
  cacheSize: 2000

logging:
  level: "info"
  enableServerLogs: true
```

### Security Best Practices

- **Never commit your actual config file to version control**
- Use strong, unique passwords for each server
- Set restrictive file permissions (600 on Unix, restricted ACL on Windows)
- Consider using environment variables for passwords: `rconPassword: "${RCON_PASSWORD}"`
- Store config files outside web-accessible directories

## Development

### Building from Source

1. **Clone the repository:**

   ```sh
   git clone https://github.com/daniel-le97/sandstorm-tracker.git
   cd sandstorm-tracker
   ```

2. **Install dependencies:**

   ```sh
   go mod download
   ```

3. **Build the project:**

   ```sh
   go build -o sandstorm-tracker main.go
   ```

   Or use Task:

   ```sh
   task build
   ```

   - Use `-check` to print database stats and exit.

## Configuration

- Edit `sandstorm-tracker.yaml` to add your servers, log paths, and database options.
- Example config:
  - [YAML file](.\assets\configs\sandstorm-tracker.example.yml)
  - [TOML file](.\assets\configs\sandstorm-tracker.example.toml)

## Usage

- Start your Insurgency: Sandstorm server(s) with logging enabled.
- Run the tracker as described above.
- Stats will be collected and stored in the configured database.

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

## License

MIT
