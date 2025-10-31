# App Package

This package consolidates the event parsing and file watching logic into a single, streamlined package.

## Design Philosophy

The `/app` package simplifies the previous architecture by:

1. **No intermediate structs** - Events are parsed and written directly to the database
2. **Single responsibility** - The parser identifies event types and processes them immediately
3. **Simplified flow** - `Log Line → Parse → Database` (no GameEvent structs in between)

## Components

### `parser.go`

- `LogParser` - Parses log lines and writes directly to database
- Pattern matching for different event types
- Direct database operations (no intermediate event structs)

Key methods:

- `ParseAndProcess(ctx, line, serverID)` - Main entry point, tries all event types
- `tryProcessKillEvent()` - Handles kill events (player kills, assists)
- `tryProcessPlayerJoin()` - Handles player join events
- `tryProcessPlayerDisconnect()` - Handles player disconnect events

### `watcher.go`

- `Watcher` - Monitors log files and processes new lines
- File offset tracking (resume from last position on restart)
- RCON client management

Key methods:

- `AddPath(path)` - Start watching a file or directory
- `Start()` - Begin watching for changes
- `Stop()` - Stop watching and save offsets
- `GetRconClient(serverID)` - Get or create RCON client for a server

## Usage Example

```go
import (
    "sandstorm-tracker/internal/app"
    "sandstorm-tracker/internal/db"
)

// Initialize database
dbService, err := db.NewDatabaseService("tracker.db")
if err != nil {
    log.Fatal(err)
}

// Load server configs
cfg, err := app.InitConfig()
if err != nil {
    log.Fatal(err)
}

// Create watcher
watcher, err := app.NewWatcher(dbService, cfg.Servers)
if err != nil {
    log.Fatal(err)
}

// Add paths to watch
for _, server := range cfg.Servers {
    if err := watcher.AddPath(server.LogPath); err != nil {
        log.Printf("Failed to watch %s: %v", server.LogPath, err)
    }
}

// Start watching
watcher.Start()

// ... run your application ...

// Cleanup
watcher.Stop()
dbService.Close()
```

## Event Processing Flow

```
Log File Change
    ↓
Watcher detects change
    ↓
Read new lines from offset
    ↓
For each line:
    ↓
ParseAndProcess(line, serverID)
    ↓
Try pattern matching:
  - Kill event? → Update player, weapon_stats, daily_stats
  - Join event? → Create/update player
  - Disconnect? → Log event
    ↓
Database updated
    ↓
Update file offset
```

## Database Operations

### Kill Events

For each killer in a multi-kill:

1. Upsert player (by SteamID)
2. Upsert player_stats (create if needed)
3. Update weapon_stats (first killer = kill, rest = assists)
4. Update daily_weapon_stats (same logic)
5. Update daily_player_stats (cumulative for the day)

### Player Join

1. Check if player exists (by SteamID)
2. Create player if new

## Benefits Over Previous Design

### Before (events + watcher packages):

```
Log Line → Parser → GameEvent struct → Handler → Database
```

- Extra allocations for GameEvent structs
- Data copied between layers
- More code to maintain

### After (app package):

```
Log Line → Parser → Database
```

- Direct processing, no intermediate structs
- Less memory allocation
- Simpler codebase
- Easier to understand and maintain

## Migration Notes

If you're migrating from the old `/events` and `/watcher` packages:

1. Replace imports:

   ```go
   // Old
   import "sandstorm-tracker/internal/events"
   import "sandstorm-tracker/internal/watcher"

   // New
   import "sandstorm-tracker/internal/app"
   ```

2. Replace watcher creation:

   ```go
   // Old
   fw, err := watcher.NewFileWatcher(dbService, serverConfigs)

   // New
   w, err := app.NewWatcher(dbService, serverConfigs)
   ```

3. The API is largely the same for the watcher component

## Testing

To test the parser:

```go
parser := app.NewLogParser(queries)
err := parser.ParseAndProcess(ctx, logLine, serverID)
```

The parser returns `nil` for successfully parsed lines (even if no database action was taken) and only returns errors for actual processing failures.
