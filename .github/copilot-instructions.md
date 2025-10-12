# Copilot Instructions for sandstorm-trackerv2

## Project Architecture

- **Purpose:** Tracks player stats, weapon usage, and match history for Insurgency: Sandstorm servers by parsing server logs and storing data in a database.
- **Major Components:**
  - `main.go`: Entry point, config loading, main event loop
  - `db/`: Database schema, queries, and access logic (SQLite by default)
  - `internal/utils/`: Utility functions (e.g., database checks, file helpers)
  - `internal/watcher/`: File/directory watching and log ingestion
  - `internal/events/`: Event parsing and processing logic
  - `internal/config/`: Configuration loading and management using Viper
  - `cmd/`: CLI tools and commands
- **Data Flow:**
  1. Log files are watched and parsed for events.
  2. Events are processed and stats are updated in the database.
  3. Configurable via `sandstorm-tracker.json` (see below).

## Configuration & Conventions

- **Config:** All runtime config is loaded from `sandstorm-tracker.json` (or `example-config.json`). Use Viper for config access. Example structure:
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
      "path": "sandstorm_stats.db",
      "enableWAL": true,
      "cacheSize": 2000
    },
    "logging": { "level": "info", "enableServerLogs": true }
  }
  ```
- **Structs:** Use Go structs with `mapstructure` tags for config unmarshalling (see `main.go`).
- **Testing:** Place tests in `*_test.go` files next to the code they test. Run with `go test ./...`.
- **Build:** Use `go build -o sandstorm-tracker main.go`, or `task build` if available.
- **Ignore:** `.gitignore` excludes binaries, logs, local configs, and editor files.

## Developer Workflows

- **Add new tracked stats:** Update event parser, database schema, and config structs as needed.
- **Add a new server:** Edit `sandstorm-tracker.json` and restart the tracker.
- **Check DB:** Run with `-check` flag to print database stats.
- **Debugging:** Use log output and database inspection for troubleshooting.

## Patterns & Integration

- **Event Parsing:** Centralized in the main loop and watcher utilities. Extend by adding new event types and handlers.
- **Database:** All access via the `db/` package. Do not bypass this layer.
- **Utilities:** Shared helpers in `internal/utils/`.
- **External:** Relies on Insurgency: Sandstorm log format and SQLite (or compatible DB).

## Examples

- To add a new stat, update the event parser and extend the database schema in `db/schema.sql`.
- To support a new server, add its log path to the config and set `enabled: true`.

---

For more details, see `README.md` in the project root.
