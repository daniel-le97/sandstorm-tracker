# PocketBase Migration Progress

## Completed ✅

1. **PocketBase Integration in main.go**

   - Integrated watcher initialization into PocketBase's `OnServe` hook
   - Added `--paths` flag for log file watching
   - Added graceful shutdown in `OnTerminate` hook
   - Removed old `internal/db` imports

2. **Updated Watcher (internal/app/watcher.go)**

   - Changed from `*db.DatabaseService` to `*pocketbase.PocketBase`
   - Updated `NewWatcher` to accept PocketBase app
   - Removed sqlc query dependencies

3. **Updated LogParser (internal/app/parser.go)**

   - Changed from `*db.Queries` to `*pocketbase.PocketBase`
   - Updated `NewLogParser` to accept PocketBase app

4. **Created PocketBase Helpers (internal/app/db_helpers.go)**
   - `GetActiveMatch` - Find active match for server
   - `CreateMatch` - Create new match record
   - `GetPlayerByExternalID` - Find player by Steam ID
   - `CreatePlayer` - Create new player
   - `UpsertMatchPlayerStats` - Create/update player stats in match
   - `IncrementMatchPlayerKills` - Increment kill count
   - `IncrementMatchPlayerAssists` - Increment assist count
   - `UpsertMatchWeaponStats` - Create/update weapon stats
   - `GetOrCreateServer` - Get or create server record

## Next Steps 🚧

### 1. Update parser.go to use new helpers

Replace all `p.queries.*` calls with the new PocketBase helper functions:

- `p.queries.GetActiveMatch` → `GetActiveMatch(ctx, p.pbApp, serverID)`
- `p.queries.CreateMatch` → `CreateMatch(ctx, p.pbApp, ...)`
- `p.queries.GetPlayerByExternalID` → `GetPlayerByExternalID(ctx, p.pbApp, ...)`
- etc.

### 2. Update watcher.go server creation

Around line 264-269, replace the server creation logic with:

```go
serverIDStr, err := GetOrCreateServer(w.ctx, w.pbApp, serverID, logPath)
```

### 3. Test the migration

- Run the app with PocketBase
- Verify log parsing works
- Check database records in PocketBase admin UI

### 4. Clean up old code

Once testing is successful:

- Delete `internal/db/` directory
- Remove `modernc.org/sqlite` from go.mod (run `go mod tidy`)
- Delete `sandstorm-tracker.db` (old SQLite file)

## How to Run

```bash
# Start the app with log watching
go run . --paths="/path/to/server/logs"

# Or build and run
go build -o sandstorm-tracker
./sandstorm-tracker --paths="/path/to/server/logs"
```

## PocketBase Collections Used

- **servers** - Server instances
- **players** - Player profiles
- **matches** - Match records
- **match_player_stats** - Per-match player statistics
- **match_weapon_stats** - Per-match weapon statistics

All collections are defined in `pb_migrations/` directory.
