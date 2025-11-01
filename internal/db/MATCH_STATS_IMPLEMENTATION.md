# Per-Match Stats Implementation

## Overview

The tracker now uses **per-match stats** as the single source of truth. All player statistics are stored per-match and aggregated on-demand for leaderboards and time-based queries.

## Key Design Decision: Session Reuse

When a player disconnects and reconnects **within the same match**, we **reuse the same record** rather than creating separate sessions:

- ✅ Simple queries (one row per player per match)
- ✅ Easy to understand and maintain
- ✅ Stats continue accumulating across reconnects
- ✅ Tracks reconnect behavior with `session_count`

## Tables

### `match_player_stats`

Stores stats for each player in each match. **One row per player per match**, regardless of reconnects.

**Key fields:**

- `session_count`: How many times the player connected during this match (1 = never disconnected, 2+ = reconnected)
- `is_currently_connected`: Whether the player is currently in the match (1) or has left (0)
- `total_play_time`: Cumulative seconds across all sessions in this match
- `first_joined_at`: When they first joined this match
- `last_left_at`: When they last disconnected (NULL if still connected)

### `match_weapon_stats`

Weapon usage per player per match. Automatically aggregates across reconnects.

## Event Handling

### Player Joins Match

```go
// UpsertMatchPlayerStats handles both initial join and rejoin
db.UpsertMatchPlayerStats(ctx, UpsertParams{
    MatchID:        currentMatchID,
    PlayerID:       playerID,
    Team:           team,
    FirstJoinedAt:  timestamp,
})
// If record exists: increments session_count, sets is_currently_connected = 1
// If new: creates record with session_count = 1
```

### Player Leaves Match

```go
// Calculate session duration
duration := time.Since(playerJoinedAt).Seconds()

db.UpdateMatchPlayerDisconnect(ctx, DisconnectParams{
    LastLeftAt:      timestamp,
    DurationSeconds: duration,
    MatchID:         currentMatchID,
    PlayerID:        playerID,
})
// Sets is_currently_connected = 0
// Adds duration to total_play_time
```

### Player Gets Kill

```go
db.IncrementMatchPlayerKills(ctx, currentMatchID, killerID)
db.IncrementMatchPlayerDeaths(ctx, currentMatchID, victimID)

// Update weapon stats
db.UpsertMatchWeaponStats(ctx, UpsertWeaponParams{
    MatchID:    currentMatchID,
    PlayerID:   killerID,
    WeaponName: weapon,
    Kills:      1,
    Assists:    0,
})
```

## Query Examples

### Get Player's Last 10 Matches

```go
matches := db.GetPlayerMatchHistory(ctx, GetPlayerMatchHistoryParams{
    PlayerID: playerID,
    Limit:    10,
})

for _, match := range matches {
    fmt.Printf("Match on %s: %d kills, %d deaths (%.2f K/D)\n",
        match.Map,
        match.Kills,
        match.Deaths,
        float64(match.Kills)/float64(match.Deaths),
    )
}
```

### Get Leaderboard for Last 30 Days

```go
thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
topPlayers := db.GetTopPlayersByKillsForPeriod(ctx, GetTopPlayersParams{
    ServerID: serverID,
    Date:     thirtyDaysAgo,
    Limit:    10,
})

// Returns aggregated stats from all matches in period
```

### Check Who's Currently in a Match

```go
activePlayers := db.GetCurrentlyConnectedPlayers(ctx, currentMatchID)
// Only returns players where is_currently_connected = 1
```

### Get Player's Best Performance

```go
bestMatch := db.GetPlayerBestMatch(ctx, playerID)
fmt.Printf("Best match: %d kills on %s\n", bestMatch.Kills, bestMatch.Map)
```

## Reconnect Behavior Example

**Scenario:** Player joins, gets 5 kills, disconnects, rejoins, gets 3 more kills

**Database state:**

```sql
SELECT * FROM match_player_stats WHERE match_id = 123 AND player_id = 456;

-- Result:
match_id | player_id | kills | session_count | is_currently_connected | total_play_time
---------|-----------|-------|---------------|------------------------|----------------
123      | 456       | 8     | 2             | 1                      | 1847
```

**What happened:**

1. Player joins → Record created with `session_count = 1`
2. Player gets 5 kills → `kills = 5`
3. Player disconnects → `is_currently_connected = 0`, `total_play_time` updated
4. **Player rejoins → REUSES same record**, `session_count = 2`, `is_currently_connected = 1`
5. Player gets 3 more kills → `kills = 8` (5 + 3)

## Migration from Old Schema

The old schema had:

- `player_stats` (lifetime totals)
- `daily_player_stats` (daily aggregates)
- `weapon_stats` (lifetime weapon totals)
- `daily_weapon_stats` (daily weapon aggregates)

**Migration strategy:**

1. Keep old tables temporarily for reference
2. All new events write to `match_player_stats` only
3. Queries aggregate from matches on-demand
4. Optional: Add materialized views for hot queries (all-time leaderboards)
5. Remove old tables once validated

## Benefits of This Approach

1. **Single source of truth** - no sync issues
2. **Flexible aggregations** - any time period, any grouping
3. **Rich analytics** - match history, performance trends, map-specific stats
4. **Simple updates** - one INSERT per kill, not three UPDATEs
5. **Natural data lifecycle** - delete match = stats gone
6. **Audit trail** - every stat ties to a specific match
