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

   - Get the latest release from [Releases](https://github.com/daniel-le97/sandstorm-trackerv2/releases)
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

The tracker accepts configuration via YAML, TOML, or JSON files. See `sandstorm-tracker.example.yml` for a complete example with all available options.

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
   git clone https://github.com/daniel-le97/sandstorm-trackerv2.git
   cd sandstorm-trackerv2
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
  - [YAML file](./sandstorm-tracker.example.yml)
  - [TOML file](./sandstorm-tracker.example.toml)

## Usage

- Start your Insurgency: Sandstorm server(s) with logging enabled.
- Run the tracker as described above.
- Stats will be collected and stored in the configured database.

## License

MIT
