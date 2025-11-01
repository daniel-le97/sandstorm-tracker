# Performance Testing

This directory contains performance tests for the match-based stats system.

## Test Database

The performance tests use a pre-generated database (`test_performance_10k.sqlite`) containing:

- **10,000 matches**
- **10 players** (reused across matches)
- **200-400 kills** per player per match
- **50-149 deaths** per player per match
- **3-7 weapons** used per player per match

**Database Stats:**

- Size: ~84 MB
- Total player-match records: 100,000
- Total weapon stat records: ~500,000
- Total kills tracked: ~30 million

## Running Performance Tests

```bash
# Run all performance tests
go test -v ./internal/db/... -run TestQueryPerformanceWith10kMatches

# Run with short flag to skip (if configured)
go test -v ./internal/db/... -short
```

## Query Performance Results

With proper indexes, queries complete in milliseconds even with 10k matches:

| Query                                  | Time  | Results                     |
| -------------------------------------- | ----- | --------------------------- |
| GetActiveMatches                       | < 1ms | All active matches          |
| GetPlayerStatsForPeriod (30 days)      | ~4ms  | Aggregated stats for player |
| GetTopPlayersByKillsForPeriod (7 days) | < 1ms | Top 10 leaderboard          |
| GetPlayerMatchHistory (100 matches)    | ~37ms | Last 100 matches            |
| GetPlayerBestMatch                     | ~14ms | Best performing match       |

## Database Indexes

The following indexes ensure fast query performance:

**Matches Table:**

- `idx_matches_server` - Server lookups
- `idx_matches_start_time` - Time-based filtering
- `idx_matches_end_time` - Active/completed matches
- `idx_matches_server_end_time` - Composite for active match queries
- `idx_matches_server_start_time` - Composite for time-based per-server queries

**Match Player Stats:**

- `idx_match_player_stats_match` - Match lookups
- `idx_match_player_stats_player` - Player lookups
- `idx_match_player_stats_connected` - Connection status
- `idx_match_player_stats_match_connected` - Composite for connected players

**Match Weapon Stats:**

- `idx_match_weapon_stats_match` - Match lookups
- `idx_match_weapon_stats_player` - Player lookups
- `idx_match_weapon_stats_weapon` - Weapon-specific queries
- `idx_match_weapon_stats_player_weapon` - Player weapon history

## Scaling Expectations

| Database Size     | Query Performance |
| ----------------- | ----------------- |
| 1,000 matches     | 1-5ms             |
| 10,000 matches    | 5-50ms            |
| 100,000 matches   | 10-100ms          |
| 1,000,000 matches | 50-500ms          |

## Notes

- The test database is kept in the repository for consistent performance testing
- All queries use proper SQLite indexes for optimal performance
- Tests validate that query times are within expected bounds
- Database uses WAL mode for better write performance
