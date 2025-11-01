# Rolling Stats with Match-Based System

## Overview

With the match-based stats system, you can easily calculate rolling time windows (last 7 days, 30 days, etc.) by aggregating match stats within that period.

**No daily aggregation tables needed!** - Just query matches within the time window.

## Example Queries

### Last 30 Days K/D Ratio for a Player

```go
package main

import (
    "context"
    "fmt"
    "time"
    db "sandstorm-tracker/internal/db/generated"
)

func GetPlayerLast30DaysKD(ctx context.Context, queries *db.Queries, playerID, serverID int64) (float64, error) {
    // Calculate 30 days ago
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

    // Get aggregated stats for this period
    stats, err := queries.GetPlayerStatsForPeriod(ctx, db.GetPlayerStatsForPeriodParams{
        PlayerID: playerID,
        ServerID: serverID,
        Date:     thirtyDaysAgo,
    })
    if err != nil {
        return 0, err
    }

    // Calculate K/D ratio
    if stats.TotalDeaths == 0 {
        return float64(stats.TotalKills), nil
    }

    kd := float64(stats.TotalKills) / float64(stats.TotalDeaths)
    return kd, nil
}
```

### Last 7 Days Leaderboard

```go
func GetLast7DaysLeaderboard(ctx context.Context, queries *db.Queries, serverID int64) ([]db.GetTopPlayersByKillsForPeriodRow, error) {
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)

    topPlayers, err := queries.GetTopPlayersByKillsForPeriod(ctx, db.GetTopPlayersByKillsForPeriodParams{
        ServerID: serverID,
        Date:     sevenDaysAgo,
        Limit:    10,
    })
    if err != nil {
        return nil, err
    }

    return topPlayers, nil
}
```

### Player Stats for Any Custom Period

```go
func GetPlayerStatsForCustomPeriod(ctx context.Context, queries *db.Queries, playerID, serverID int64, days int) (*PlayerStats, error) {
    startDate := time.Now().AddDate(0, 0, -days)

    stats, err := queries.GetPlayerStatsForPeriod(ctx, db.GetPlayerStatsForPeriodParams{
        PlayerID: playerID,
        ServerID: serverID,
        Date:     startDate,
    })
    if err != nil {
        return nil, err
    }

    kd := 0.0
    if stats.TotalDeaths > 0 {
        kd = float64(stats.TotalKills) / float64(stats.TotalDeaths)
    }

    return &PlayerStats{
        MatchesPlayed: stats.TotalMatches,
        Kills:         stats.TotalKills,
        Deaths:        stats.TotalDeaths,
        Assists:       stats.TotalAssists,
        KDRatio:       kd,
        TotalScore:    stats.TotalScore,
        PlayTime:      time.Duration(stats.TotalPlayTime) * time.Second,
        Reconnects:    stats.TotalSessions - stats.TotalMatches, // Extra sessions = reconnects
    }, nil
}

type PlayerStats struct {
    MatchesPlayed int64
    Kills         int64
    Deaths        int64
    Assists       int64
    KDRatio       float64
    TotalScore    int64
    PlayTime      time.Duration
    Reconnects    int64
}
```

### Top Weapons for Last 30 Days

```go
func GetPlayerTopWeaponsLast30Days(ctx context.Context, queries *db.Queries, playerID int64) ([]db.GetPlayerTopWeaponsForPeriodRow, error) {
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

    weapons, err := queries.GetPlayerTopWeaponsForPeriod(ctx, db.GetPlayerTopWeaponsForPeriodParams{
        PlayerID: playerID,
        Date:     thirtyDaysAgo,
        Limit:    5, // Top 5 weapons
    })
    if err != nil {
        return nil, err
    }

    return weapons, nil
}
```

## RCON Chat Command Examples

### !stats Command (Last 30 Days)

```go
func HandleStatsCommand(ctx context.Context, queries *db.Queries, playerID, serverID int64, playerName string) string {
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

    stats, err := queries.GetPlayerStatsForPeriod(ctx, db.GetPlayerStatsForPeriodParams{
        PlayerID: playerID,
        ServerID: serverID,
        Date:     thirtyDaysAgo,
    })
    if err != nil {
        return fmt.Sprintf("Error fetching stats for %s", playerName)
    }

    kd := 0.0
    if stats.TotalDeaths > 0 {
        kd = float64(stats.TotalKills) / float64(stats.TotalDeaths)
    }

    return fmt.Sprintf(
        "%s (30d): %d kills | %d deaths | %.2f K/D | %d matches",
        playerName,
        stats.TotalKills,
        stats.TotalDeaths,
        kd,
        stats.TotalMatches,
    )
}
```

### !top Command (Last 30 Days)

```go
func HandleTopCommand(ctx context.Context, queries *db.Queries, serverID int64) string {
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

    topPlayers, err := queries.GetTopPlayersByKillsForPeriod(ctx, db.GetTopPlayersByKillsForPeriodParams{
        ServerID: serverID,
        Date:     thirtyDaysAgo,
        Limit:    3,
    })
    if err != nil {
        return "Error fetching leaderboard"
    }

    if len(topPlayers) == 0 {
        return "No stats available for the last 30 days"
    }

    msg := "Top Players (30d): "
    for i, player := range topPlayers {
        kd := 0.0
        if player.TotalDeaths > 0 {
            kd = float64(player.TotalKills) / float64(player.TotalDeaths)
        }
        msg += fmt.Sprintf("%d. %s (%d kills, %.2f K/D) ", i+1, player.Name, player.TotalKills, kd)
    }

    return msg
}
```

## Performance Considerations

### Query Performance

The queries use indexes on:

- `match_player_stats(player_id)` - Fast player lookups
- `match_player_stats(match_id)` - Fast match lookups
- `matches(server_id)` - Filter by server
- `matches(start_time)` - Time-based filtering

### Example Query Plan (30 days):

```sql
-- For a server with 1000 matches/day
-- 30 days = ~30,000 matches to scan
-- With proper indexes, this is FAST (< 100ms)

SELECT SUM(kills), SUM(deaths), COUNT(*)
FROM match_player_stats mps
JOIN matches m ON m.id = mps.match_id
WHERE mps.player_id = 123
  AND m.server_id = 1
  AND m.start_time >= '2025-10-01';
```

### Optimization Tips

1. **Add composite index** if needed:

   ```sql
   CREATE INDEX idx_matches_server_time ON matches(server_id, start_time);
   ```

2. **Cache hot queries** - Cache leaderboards for 5-10 minutes:

   ```go
   type CachedLeaderboard struct {
       Data      []db.GetTopPlayersByKillsForPeriodRow
       CachedAt  time.Time
   }

   var leaderboardCache *CachedLeaderboard

   if leaderboardCache == nil || time.Since(leaderboardCache.CachedAt) > 5*time.Minute {
       // Refresh cache
       leaderboardCache = &CachedLeaderboard{
           Data:     fetchLeaderboard(),
           CachedAt: time.Now(),
       }
   }
   ```

3. **Use materialized views** for all-time stats (optional):
   ```sql
   -- Create a view for all-time player stats
   CREATE VIEW player_lifetime_stats AS
   SELECT
       mps.player_id,
       SUM(mps.kills) as lifetime_kills,
       SUM(mps.deaths) as lifetime_deaths,
       COUNT(DISTINCT mps.match_id) as total_matches
   FROM match_player_stats mps
   GROUP BY mps.player_id;
   ```

## Benefits Over Daily Aggregation

✅ **Flexibility**: Any time period (7 days, 30 days, 90 days, custom ranges)
✅ **Accuracy**: No sync issues between aggregation tables
✅ **Simplicity**: One source of truth (match stats)
✅ **Match-level detail**: Can drill down to individual matches
✅ **No maintenance**: No daily cleanup jobs needed
✅ **Performance**: With proper indexes, queries are fast enough

## Common Time Periods

```go
// Helper functions for common periods
func Last7Days() time.Time   { return time.Now().AddDate(0, 0, -7) }
func Last30Days() time.Time  { return time.Now().AddDate(0, 0, -30) }
func Last90Days() time.Time  { return time.Now().AddDate(0, 0, -90) }
func ThisMonth() time.Time   {
    now := time.Now()
    return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}
func LastMonth() time.Time   {
    now := time.Now()
    return time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
}
```

## Summary

With the match-based system, rolling stats are **simple aggregations**:

- Query matches within time period using `m.start_time >= ?`
- SUM the kills/deaths/assists from `match_player_stats`
- Calculate K/D, average score, etc. in code

**No separate daily stats tables needed!** 🎉
