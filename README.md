# Sandstorm Tracker

Sandstorm Tracker is a Go project that tracks kills, playtime, alive time, weapon stats, and match history for Insurgency: Sandstorm servers. It ingests server logs, parses events, and stores statistics in a database for analysis and visualization.

## Features

- Tracks player kills, deaths, and assists
- Records playtime and alive time per player
- Collects weapon usage and stats
- Maintains match history and session data
- Supports multiple servers
- Configurable via JSON config file

## Requirements

- Go 1.20+
- Insurgency: Sandstorm server(s) with log access
- SQLite (default) or compatible database
- (Optional) GNU Make or Taskfile for automation

## Setup

1. **Clone the repository:**
   ```sh
   git clone https://github.com/daniel-le97/sandstorm-trackerv2.git
   cd sandstorm-trackerv2
   ```
2. **Configure your servers:**

   - Copy `example-config.json` or `sandstorm-tracker.json` to your project root and edit paths, server names, and database settings as needed.

3. **Build the project:**

   ```sh
   go build -o sandstorm-tracker main.go
   ```

   Or use Taskfile/Makefile if available:

   ```sh
   task build
   # or
   make
   ```

4. **Run the tracker:**
   ```sh
   ./sandstorm-tracker -paths="/path/to/your/Insurgency/Saved/Logs"
   ```
   - You can specify multiple log paths in the config file.
   - Use `-db` to set a custom database path.
   - Use `-check` to print database stats and exit.

## Configuration

- Edit `sandstorm-tracker.json` to add your servers, log paths, and database options.
- Example config:
  ```json
  {
    "servers": [
      {
        "name": "Main Server",
        "logPath": "/opt/sandstorm/Insurgency/Saved/Logs",
        "enabled": true
      }
    ],
    "database": {
      "path": "sandstorm-tracker.db",
      "enableWAL": true,
      "cacheSize": 2000
    },
    "logging": { "level": "info", "enableServerLogs": true }
  }
  ```

## Usage

- Start your Insurgency: Sandstorm server(s) with logging enabled.
- Run the tracker as described above.
- Stats will be collected and stored in the configured database.

## Development

- Standard Go project layout.
- Main logic in `main.go`, utilities in `internal/utils/`, database code in `db/`.
- Tests can be run with:
  ```sh
  go test ./...
  ```

## License

MIT
