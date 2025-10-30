# Daily Stats Usage Guide

## Overview

The tracker now maintains both **lifetime stats** and **daily stats** to enable rolling time window queries (last 7/30/60 days).

## Database Tables

### Lifetime Stats

- `player_stats` - All-time player totals
- `weapon_stats` - All-time weapon usage per player

### Daily Stats (New!)

- `daily_player_stats` - Player stats per day
- `daily_weapon_stats` - Weapon stats per player per day

## Common Queries

### 1. Last 30 Days Leaderboard

```go
// Get top 10 players by kills in last 30 days
thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
topPlayers, err := queries.GetTopPlayersByKillsForPeriod(ctx, db.GetTopPlayersByKillsForPeriodParams{
    ServerID: serverID,
    Date:     thirtyDaysAgo,
    Limit:    10,
})

// Chat output: "Top Players (30d): 1. Player123 (1,247) 2. PlayerXYZ (1,103) ..."
```

### 2. Player Stats for Last 30 Days

```go
// Get individual player stats for last 30 days
thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
stats, err := queries.GetPlayerStatsForPeriod(ctx, db.GetPlayerStatsForPeriodParams{
    PlayerID: playerID,
    ServerID: serverID,
    Date:     thirtyDaysAgo,
})

kd := float64(stats.TotalKills) / float64(stats.TotalDeaths)
// Chat: "Player123: 1,247 kills | 89 assists | 156 deaths | 8.0 K/D (Last 30 days)"
```

### 3. Top Weapons for Last 30 Days

```go
// Get player's top weapons for last 30 days
thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
weapons, err := queries.GetTopWeaponsForPlayerForPeriod(ctx, db.GetTopWeaponsForPlayerForPeriodParams{
    PlayerID: playerID,
    ServerID: serverID,
    Date:     thirtyDaysAgo,
    Limit:    5,
})

// Chat: "Top Weapons (30d): AK-74 (412) | M16A4 (287) | M9 (98)"
```

### 4. Daily Trend (Last 7 Days)

```go
// Get player's daily performance for last 7 days
sevenDaysAgo := time.Now().AddDate(0, 0, -7)
trend, err := queries.GetPlayerDailyTrend(ctx, db.GetPlayerDailyTrendParams{
    PlayerID: playerID,
    ServerID: serverID,
    Date:     sevenDaysAgo,
})

// Could show a simple trend indicator
// "Last 7 days: 287 kills (avg 41/day) ↑ improving"
```

### 5. All-Time Stats (Still Available!)

```go
// Get lifetime totals from weapon_stats
totalKills, err := queries.GetTotalKillsForPlayerStats(ctx, playerStatsID)

// Chat: "All-time: 15,234 kills | Member since: Jan 2025"
```

## Data Retention

Daily stats should be pruned periodically to keep database size manageable:

```go
// Delete daily stats older than 60 days (run daily)
sixtyDaysAgo := time.Now().AddDate(0, 0, -60)
err := queries.DeleteOldDailyPlayerStats(ctx, sixtyDaysAgo)
err = queries.DeleteOldDailyWeaponStats(ctx, sixtyDaysAgo)
```

## Storage Estimates

With 60 players across 6 servers:

- **Daily stats (60-day retention)**: ~20-30 MB
- **Lifetime stats**: ~5 MB
- **Total**: ~25-35 MB (very manageable!)

## Benefits

✅ True rolling windows (last 7/14/30 days)  
✅ Fair comparisons (same time period for all players)  
✅ Smooth stats progression (no hard resets)  
✅ Keep all-time totals (for player profiles)  
✅ Small storage footprint  
✅ Fast queries (with date indexes)
