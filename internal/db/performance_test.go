package db

import (
	"context"
	"os"
	"testing"
	"time"

	gen "sandstorm-tracker/internal/db/generated"
)

// TestQueryPerformanceWith10kMatches tests query performance on existing 10k match database
// This test expects test_performance_10k.sqlite to exist (created by data generation script)
func TestQueryPerformanceWith10kMatches(t *testing.T) {
	dbPath := "test_performance_10k.sqlite"

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Skip("Performance database not found. Generate it first with the data generation test.")
	}

	dbService, err := NewDatabaseService(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer dbService.Close()

	queries := dbService.GetQueries()
	ctx := context.Background()

	// Get database info
	dbInfo, _ := os.Stat(dbPath)
	t.Logf("\n=== Database Info ===")
	t.Logf("Database size: %.2f MB", float64(dbInfo.Size())/(1024*1024))

	// Get first server and player for testing
	servers, err := queries.ListServers(ctx)
	if err != nil || len(servers) == 0 {
		t.Fatalf("failed to get servers: %v", err)
	}
	server := servers[0]

	players, err := queries.ListPlayers(ctx)
	if err != nil || len(players) == 0 {
		t.Fatalf("failed to get players: %v", err)
	}
	player := players[0]

	// Run performance queries
	t.Logf("\n=== Running Performance Queries ===")

	// Query 1: Get active matches
	queryStart := time.Now()
	activeMatches, err := queries.GetActiveMatches(ctx)
	if err != nil {
		t.Fatalf("GetActiveMatches failed: %v", err)
	}
	t.Logf("GetActiveMatches: %v (found %d)", time.Since(queryStart), len(activeMatches))

	// Query 2: Get player stats for last 30 days
	queryStart = time.Now()
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	stats, err := queries.GetPlayerStatsForPeriod(ctx, gen.GetPlayerStatsForPeriodParams{
		PlayerID:  player.ID,
		ServerID:  server.ID,
		StartTime: &thirtyDaysAgo,
	})
	if err != nil {
		t.Fatalf("GetPlayerStatsForPeriod failed: %v", err)
	}
	queryTime := time.Since(queryStart)
	t.Logf("GetPlayerStatsForPeriod (30 days): %v", queryTime)

	kdr := float64(0)
	if stats.TotalDeaths != nil && *stats.TotalDeaths > 0 {
		kdr = *stats.TotalKills / *stats.TotalDeaths
	}
	t.Logf("  - Matches: %v, Kills: %.0f, Deaths: %.0f, KDR: %.2f",
		stats.TotalMatches, derefFloat64(stats.TotalKills), derefFloat64(stats.TotalDeaths), kdr)

	// Query 3: Get top players for last 7 days
	queryStart = time.Now()
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	topPlayers, err := queries.GetTopPlayersByKillsForPeriod(ctx, gen.GetTopPlayersByKillsForPeriodParams{
		ServerID:  server.ID,
		StartTime: &sevenDaysAgo,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("GetTopPlayersByKillsForPeriod failed: %v", err)
	}
	queryTime = time.Since(queryStart)
	t.Logf("GetTopPlayersByKillsForPeriod (7 days, top 10): %v", queryTime)
	for i, p := range topPlayers {
		pkdr := float64(0)
		if p.TotalDeaths != nil && *p.TotalDeaths > 0 {
			pkdr = float64(*p.TotalKills) / float64(*p.TotalDeaths)
		}
		t.Logf("  %d. %s - Kills: %.0f, Deaths: %.0f, KDR: %.2f, Matches: %v",
			i+1, p.Name, derefFloat64(p.TotalKills), derefFloat64(p.TotalDeaths), pkdr, p.MatchesPlayed)
	}

	// Query 4: Get player match history
	queryStart = time.Now()
	matchHistory, err := queries.GetPlayerMatchHistory(ctx, gen.GetPlayerMatchHistoryParams{
		PlayerID: player.ID,
		Limit:    100,
	})
	if err != nil {
		t.Fatalf("GetPlayerMatchHistory failed: %v", err)
	}
	t.Logf("GetPlayerMatchHistory (last 100 matches): %v (found %d)",
		time.Since(queryStart), len(matchHistory))

	// Query 5: Get player's best match
	queryStart = time.Now()
	bestMatch, err := queries.GetPlayerBestMatch(ctx, player.ID)
	if err != nil {
		t.Fatalf("GetPlayerBestMatch failed: %v", err)
	}
	t.Logf("GetPlayerBestMatch: %v", time.Since(queryStart))
	killsVal := int64(0)
	if bestMatch.Kills != nil {
		killsVal = *bestMatch.Kills
	}
	t.Logf("  - Map: %v, Mode: %v, Kills: %d",
		derefString(bestMatch.Map), bestMatch.Mode, killsVal)

	// Performance assertions - these should complete quickly with proper indexes
	if queryTime > 100*time.Millisecond {
		t.Logf("WARNING: Query took %v (expected < 100ms with proper indexes)", queryTime)
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
