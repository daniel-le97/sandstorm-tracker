# Copilot Instructions for sandstorm-tracker

## Project Architecture

- **Purpose:** Tracks player stats, weapon usage, and match history for Insurgency: Sandstorm servers by parsing server logs and storing data in PocketBase.
- **Backend:** PocketBase v0.31.0 - embedded database with real-time subscriptions and admin dashboard
- **Data Flow:**
  1. PocketBase starts and runs migrations to create collections
  2. Watcher is initialized in OnServe hook with server paths from config
  3. Log files are watched and parsed for events
  4. Events are processed and stored using PocketBase Record API
  5. Stats are queryable via PocketBase API or admin dashboard

## Configuration & Conventions

- **Config:** Runtime config loaded from YAML/TOML files using Viper:

  ```yaml
  servers:
    - name: "Main Server"
      logPath: "/opt/sandstorm/Insurgency/Saved/Logs"
      enabled: true

  logging:
    level: "info"
    enableServerLogs: true
  ```

- **PocketBase Collections:**
  - `servers`: Server records (external_id, path)
  - `players`: Player records (external_id/Steam ID, name)
  - `matches`: Match records (server, map, mode, start_time, end_time)
  - `match_player_stats`: Per-match player stats (kills, deaths, assists)
  - `match_weapon_stats`: Per-match weapon stats (weapon, kills)
- **Testing:**
  - Use `tests.NewTestApp()` for PocketBase tests
  - Register `jsvm` plugin to run JavaScript migrations in tests
  - Place tests in `*_test.go` files, run with `go test ./...`
- **Build:** Standard Go build or use PocketBase CLI
- **Migrations:** Create new migrations using PocketBase admin dashboard

## Developer Workflows

- **Add new tracked stats:**
  1. Create migration to add fields to collections
  2. Update parser to extract new stats from logs
  3. Update db_helpers to store new stats
- **Add a new server:** Edit config YAML and restart (or use PocketBase admin UI)
- **View data:** Access PocketBase admin dashboard at `http://localhost:8090/_/`
- **Debugging:** Use log output and PocketBase admin dashboard for data inspection

## Patterns & Integration

- **Event Parsing:** Parser extracts events from logs and uses db_helpers to store in PocketBase
- **Database Access:** Always use PocketBase Record API through db_helpers functions:
  - `GetOrCreateServer()` - Find or create server record
  - `GetActiveMatch()` - Get current active match for a server
  - `CreateMatch()` - Start new match
  - `GetPlayerByExternalID()` - Find player by Steam ID
  - `CreatePlayer()` - Register new player
  - `UpsertMatchPlayerStats()` - Update player stats in match
  - `IncrementMatchPlayerKills()` - Increment kill count
  - `UpsertMatchWeaponStats()` - Update weapon usage stats
- **PocketBase Hooks:** Use OnServe, OnTerminate for lifecycle management

## Examples
- To query stats: Use PocketBase Record API filters or admin dashboard
- To test: Use `tests.NewTestApp("./test_pb_data")` and jsvm plugin for migrations

---

For more details, see `README.md` in the project root.
