# Server Crash Recovery Strategy

## Problem

When a game server crashes, you may have:

1. **Unfinished matches** - `end_time` is NULL
2. **Ghost connections** - Players marked as `is_currently_connected = 1` who are actually gone
3. **No match end event** - No winner_team recorded

## Detection

### On Startup / Periodic Check

Check for stale matches that are likely from a crash:

```go
// Find matches running longer than 2 hours (Insurgency matches are usually 20-45 min)
twoHoursAgo := time.Now().Add(-2 * time.Hour)

// Option 1: Get all active matches and check manually
activeMatches, err := queries.GetActiveMatches(ctx)
for _, match := range activeMatches {
    if match.StartTime.Before(twoHoursAgo) {
        // This is probably a crashed server
        handleStaleMatch(match)
    }
}

// Option 2: Use the auto-cleanup query
err := queries.ForceEndStaleMatches(ctx, twoHoursAgo)
```

## Recovery Strategies

### Strategy 1: Force-End on Startup (Recommended)

When the tracker starts, automatically clean up any orphaned matches:

```go
func (app *Application) RecoverFromCrashes(ctx context.Context) error {
    log.Info("Checking for crashed/unfinished matches...")

    // Force-end matches older than 2 hours
    twoHoursAgo := time.Now().Add(-2 * time.Hour)
    err := app.db.ForceEndStaleMatches(ctx, twoHoursAgo)
    if err != nil {
        return fmt.Errorf("failed to end stale matches: %w", err)
    }

    // Find and disconnect ghost connections
    staleConnections, err := app.db.GetStaleConnections(ctx)
    if err != nil {
        return fmt.Errorf("failed to get stale connections: %w", err)
    }

    for _, conn := range staleConnections {
        log.Warnf("Disconnecting ghost connection: player %d in ended match %d",
            conn.PlayerID, conn.MatchID)
        err = app.db.UpdateMatchPlayerDisconnect(ctx, db.DisconnectParams{
            MatchID:         conn.MatchID,
            PlayerID:        conn.PlayerID,
            LastLeftAt:      conn.UpdatedAt, // Use last update as disconnect time
            DurationSeconds: 0, // Can't calculate accurate duration
        })
    }

    log.Info("Crash recovery complete")
    return nil
}
```

### Strategy 2: Detect on New Match Start

When a new match starts on a server, check if there's still an active match:

```go
func (app *Application) HandleMatchStart(serverID int64, timestamp time.Time) error {
    // Check for existing active match on this server
    activeMatch, err := app.db.GetActiveMatch(ctx, serverID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return err
    }

    if activeMatch != nil {
        log.Warnf("Found unfinished match %d on server %d, force-ending it",
            activeMatch.ID, serverID)

        // Force-end the old match
        err = app.db.EndMatch(ctx, db.EndMatchParams{
            ID:         activeMatch.ID,
            EndTime:    timestamp,
            WinnerTeam: sql.NullInt64{Valid: false}, // Unknown winner
        })

        // Disconnect all players from old match
        err = app.db.DisconnectAllPlayersInMatch(ctx, db.DisconnectParams{
            MatchID:    activeMatch.ID,
            Timestamp:  timestamp,
        })
    }

    // Now create the new match
    return app.createNewMatch(serverID, timestamp)
}
```

### Strategy 3: Periodic Cleanup Task

Run a background task every 30 minutes to clean up stale matches:

```go
func (app *Application) StartCleanupWorker(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Clean up matches running longer than 2 hours
            twoHoursAgo := time.Now().Add(-2 * time.Hour)
            err := app.db.ForceEndStaleMatches(ctx, twoHoursAgo)
            if err != nil {
                log.Errorf("Cleanup worker failed: %v", err)
            }
        }
    }
}
```

## What the Queries Do

### `ForceEndStaleMatches`

```sql
UPDATE matches
SET
    end_time = COALESCE(
        (SELECT MAX(updated_at) FROM match_player_stats WHERE match_id = matches.id),
        start_time
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE end_time IS NULL AND start_time < ?;
```

**Logic:**

- Finds matches with `end_time IS NULL` older than threshold
- Sets `end_time` to the last player activity (or start_time if no activity)
- This preserves accurate match duration when possible

### `DisconnectAllPlayersInMatch`

Disconnects all players still marked as connected in a specific match.

### `GetStaleConnections`

Finds inconsistencies where players are marked as connected but the match has ended.

## Recommended Approach

**Combine all three strategies:**

1. **On startup**: Clean up any stale matches from previous crashes
2. **On match start**: Check for active match on same server and force-end it
3. **Background task**: Periodic cleanup as a safety net

```go
func main() {
    // ... initialization ...

    // 1. Startup recovery
    if err := app.RecoverFromCrashes(ctx); err != nil {
        log.Fatalf("Crash recovery failed: %v", err)
    }

    // 2. Start background cleanup worker
    go app.StartCleanupWorker(ctx)

    // 3. Handle match events with active match detection
    // ... event loop ...
}
```

## Configuration

Add to `sandstorm-tracker.yml`:

```yaml
database:
  cleanup:
    enabled: true
    stale_match_threshold: 2h # Force-end matches older than this
    cleanup_interval: 30m # How often to run periodic cleanup
```

## Benefits

✅ **Automatic recovery** from server crashes
✅ **Data integrity** - no ghost connections or stuck matches
✅ **Accurate stats** - uses last known player activity for end_time
✅ **Multiple safety nets** - startup + event-based + periodic
✅ **Low overhead** - only runs on startup and periodically
