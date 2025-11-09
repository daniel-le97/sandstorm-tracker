# Sandstorm Tracker Application Flow

## A2S Cron Job Activation Flow

The A2S (Source Server Query Protocol) cron jobs use a **lazy activation** pattern to optimize resource usage. They are not started immediately when the application starts, but rather when servers become active.

### 1. Application Startup (`app.onServe()`)

When the application starts:

- ✅ PocketBase server initializes
- ✅ Database migrations run
- ✅ Configuration is loaded and validated
- ✅ RCON pool is initialized with configured servers
- ✅ A2S pool is initialized with configured servers
- ✅ File watcher is created with server log paths
- ✅ Web routes are registered
- ✅ Callbacks are registered for server lifecycle events:
  - `OnServerActive` → Registers A2S cron job for that server
  - `OnServerInactive` → Unregisters A2S cron job for that server
- ✅ File watcher starts monitoring log files
- ❌ **A2S cron jobs are NOT started yet**

### 2. Log File Monitoring

The watcher continuously monitors configured log files:

```go
// Watches for file changes in real-time
for _, serverCfg := range app.Config.Servers {
    if serverCfg.Enabled {
        app.Watcher.AddPath(serverCfg.LogPath)
    }
}
```

### 3. Server Becomes "Active"

A server is marked as "active" when:

1. **New log content is detected** in the log file
2. **Log lines are successfully processed** by the parser
3. **First log rotation after startup** is detected

When this happens:

```go
// In watcher.go - after processing lines
if linesProcessed > 0 {
    // Mark server as active and trigger callback (only fires once per server)
    w.markServerActive(serverID)
}
```

The `markServerActive()` function:

- Checks if server is already marked as active (prevents duplicate registrations)
- Sets `activeServers[serverID] = true`
- Logs: `Server {serverID} became active (log rotation detected)`
- Triggers the `onServerActive` callback

### 4. A2S Cron Job Registration

When the `onServerActive` callback fires:

```go
// In app.go
app.Watcher.OnServerActive(func(serverID string) {
    app.Logger().Info("Server became active, registering A2S cron job", "serverID", serverID)
    jobs.RegisterA2SForServer(app, app.Config, serverID)
})
```

This creates a unique cron job:

- **Job Name**: `a2s_player_scores_{serverID}`
- **Schedule**: Every minute (`* * * * *`)
- **Action**:
  - Query server via A2S protocol
  - Retrieve current player list and scores
  - Update `match_player_stats` table with current scores
  - Create/update player records as needed

### 5. A2S Job Execution

Once registered, the A2S cron job runs every minute:

```go
// Query server status via A2S protocol
status, err := pool.QueryServer(ctx, queryAddr)

// For each player on the server:
for _, player := range status.Players {
    // Find or create player by name
    playerRecord, _ := findOrCreatePlayerByName(app, player.Name)

    // Update their score in the active match
    updatePlayerMatchScore(app, activeMatch.Id, playerRecord.Id, player.Score)
}
```

### 6. Server Becomes "Inactive"

A server is marked as "inactive" when:

- **No log activity for 10 seconds**
- Detected by the inactivity monitor goroutine

When this happens:

```go
// In app.go
app.Watcher.OnServerInactive(func(serverID string) {
    app.Logger().Info("Server became inactive, unregistering A2S cron job", "serverID", serverID)
    jobs.UnregisterA2SForServer(app, serverID)
})
```

The A2S cron job is removed:

- Job `a2s_player_scores_{serverID}` is unregistered
- Stops querying inactive servers
- Saves system resources

### 7. Reactivation

If a server becomes active again (new logs detected):

- The `OnServerActive` callback fires again
- A new A2S cron job is registered
- Monitoring resumes

## Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Application Starts                                           │
│    - Initialize components                                      │
│    - Register lifecycle callbacks                               │
│    - Start file watcher                                         │
│    ❌ No A2S cron jobs yet                                      │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Watcher Monitors Log Files                                   │
│    - Detects file changes                                       │
│    - Waits for new content                                      │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼ (Log rotation detected)
┌─────────────────────────────────────────────────────────────────┐
│ 3. Server Becomes Active                                        │
│    - New log lines processed                                    │
│    - markServerActive(serverID) called                          │
│    - onServerActive callback fires (once per server)            │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. A2S Cron Job Registered                                      │
│    ✅ Job: a2s_player_scores_{serverID}                         │
│    ✅ Schedule: Every minute                                    │
│    ✅ Action: Query server & update scores                      │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼ (Every minute)
┌─────────────────────────────────────────────────────────────────┐
│ 5. A2S Job Executes                                             │
│    - Query server via A2S protocol                              │
│    - Get player list and scores                                 │
│    - Update match_player_stats table                            │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼ (No logs for 10s)
┌─────────────────────────────────────────────────────────────────┐
│ 6. Server Becomes Inactive                                      │
│    - onServerInactive callback fires                            │
│    - A2S cron job unregistered                                  │
│    - Stops querying server                                      │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼ (New logs detected)
                    ↻ Back to Step 3 (Reactivation)
```

## Key Benefits of Lazy Activation

1. **Resource Efficiency**: Only query servers that are actively running
2. **Automatic Discovery**: No manual intervention needed to start monitoring
3. **Automatic Cleanup**: Stops querying when servers go offline
4. **Per-Server Isolation**: Each server has its own cron job
5. **Resilient**: Automatically recovers when servers restart

## Data Flow: Score Updates

```
A2S Query → Server Response → Find/Create Player → Find Active Match → Update match_player_stats
```

**Tables Involved**:

- `servers`: Server records with external_id and log paths
- `matches`: Active match records (where `end_time = ''`)
- `players`: Player records (created by name from A2S or Steam ID from logs)
- `match_player_stats`: Per-match player statistics including **score**

**Score Field**: Stored in `match_player_stats.score` (updated every minute for active players)
