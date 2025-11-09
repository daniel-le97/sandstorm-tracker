package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// testSetup creates a test app, server, and match for testing
func testSetup(t *testing.T) (core.App, context.Context, string, *Match) {
	t.Helper()

	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(testApp.Cleanup)

	ctx := context.Background()
	serverExternalID := "test-server"

	_, err = GetOrCreateServer(ctx, testApp, serverExternalID, "Test Server", "/path/to/server")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	mapName := "Map1"
	mode := "Push"
	startTime := time.Now()
	match, err := CreateMatch(ctx, testApp, serverExternalID, &mapName, &mode, &startTime)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	return testApp, ctx, serverExternalID, match
}

// createTestPlayer creates a player and optionally adds them to a match with stats
func createTestPlayer(t *testing.T, ctx context.Context, app core.App, steamID, name string, match *Match, joinTime *time.Time) *Player {
	t.Helper()

	player, err := CreatePlayer(ctx, app, steamID, name)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	if match != nil && joinTime != nil {
		err = UpsertMatchPlayerStats(ctx, app, match.ID, player.ID, nil, joinTime)
		if err != nil {
			t.Fatalf("Failed to create match player stats: %v", err)
		}
	}

	return player
}

// updatePlayerStats updates match player stats using raw SQL
func updatePlayerStats(t *testing.T, app core.App, matchID, playerID string, updates map[string]any) {
	t.Helper()

	query := "UPDATE match_player_stats SET "
	first := true
	for key := range updates {
		if !first {
			query += ", "
		}
		query += key + " = {:" + key + "}"
		first = false
	}
	query += " WHERE match = {:match} AND player = {:player}"

	updates["match"] = matchID
	updates["player"] = playerID

	_, err := app.DB().NewQuery(query).Bind(updates).Execute()
	if err != nil {
		t.Fatalf("Failed to update player stats: %v", err)
	}
}

func TestGetPlayerTotalKD(t *testing.T) {
	testApp, ctx, _, match := testSetup(t)

	// Create test player with stats
	startTime := time.Now()
	player := createTestPlayer(t, ctx, testApp, "76561198000000001", "TestPlayer", match, &startTime)

	// Set kills and deaths
	updatePlayerStats(t, testApp, match.ID, player.ID, map[string]any{
		"kills":  10,
		"deaths": 5,
	})

	// Test GetPlayerTotalKD
	kills, deaths, err := GetPlayerTotalKD(ctx, testApp, player.ID)
	if err != nil {
		t.Fatalf("GetPlayerTotalKD failed: %v", err)
	}

	if kills != 10 {
		t.Errorf("Expected 10 kills, got %d", kills)
	}
	if deaths != 5 {
		t.Errorf("Expected 5 deaths, got %d", deaths)
	}

	// Test with player that has no stats
	emptyPlayer := createTestPlayer(t, ctx, testApp, "76561198000000002", "EmptyPlayer", nil, nil)

	kills, deaths, err = GetPlayerTotalKD(ctx, testApp, emptyPlayer.ID)
	if err != nil {
		t.Fatalf("GetPlayerTotalKD failed for empty player: %v", err)
	}

	if kills != 0 {
		t.Errorf("Expected 0 kills for empty player, got %d", kills)
	}
	if deaths != 0 {
		t.Errorf("Expected 0 deaths for empty player, got %d", deaths)
	}
}

func TestGetPlayerStats(t *testing.T) {
	testApp, ctx, _, match := testSetup(t)

	// Create player stats with score and time
	joinTime := time.Now().Add(-30 * time.Minute)
	leftTime := time.Now()

	player := createTestPlayer(t, ctx, testApp, "76561198000000001", "TestPlayer", match, &joinTime)

	// Set score and times
	updatePlayerStats(t, testApp, match.ID, player.ID, map[string]any{
		"score":           1500,
		"first_joined_at": joinTime.Format("2006-01-02 15:04:05.000Z"),
		"last_left_at":    leftTime.Format("2006-01-02 15:04:05.000Z"),
	})

	// Test GetPlayerStats
	stats, err := GetPlayerStats(ctx, testApp, player.ID)
	if err != nil {
		t.Fatalf("GetPlayerStats failed: %v", err)
	}

	if stats.TotalScore != 1500 {
		t.Errorf("Expected score 1500, got %d", stats.TotalScore)
	}

	// Duration should be approximately 30 minutes (1800 seconds)
	// Allow some tolerance for time calculation differences
	if stats.TotalDurationSeconds < 1750 || stats.TotalDurationSeconds > 1850 {
		t.Errorf("Expected duration around 1800 seconds, got %d", stats.TotalDurationSeconds)
	}
}

func TestGetTopPlayersByScorePerMin(t *testing.T) {
	testApp, ctx, _, match := testSetup(t)

	// Create 3 players with different scores and play times
	players := []struct {
		name  string
		score int
		mins  int // minutes played
	}{
		{"TopPlayer", 3000, 10},   // 300 score/min
		{"MidPlayer", 2000, 20},   // 100 score/min
		{"LowPlayer", 1000, 20},   // 50 score/min
		{"NoTimePlayer", 5000, 0}, // Should be excluded (< 1 min)
	}

	for _, p := range players {
		joinTime := time.Now().Add(-time.Duration(p.mins) * time.Minute)
		leftTime := time.Now()

		player := createTestPlayer(t, ctx, testApp, "steam_"+p.name, p.name, match, &joinTime)

		// Set score and times
		updatePlayerStats(t, testApp, match.ID, player.ID, map[string]any{
			"score":           p.score,
			"first_joined_at": joinTime.Format("2006-01-02 15:04:05.000Z"),
			"last_left_at":    leftTime.Format("2006-01-02 15:04:05.000Z"),
		})
	}

	// Test GetTopPlayersByScorePerMin
	topPlayers, err := GetTopPlayersByScorePerMin(ctx, testApp, 10)
	if err != nil {
		t.Fatalf("GetTopPlayersByScorePerMin failed: %v", err)
	}

	// Should return 3 players (NoTimePlayer excluded for < 1 min play time)
	if len(topPlayers) != 3 {
		t.Fatalf("Expected 3 players, got %d", len(topPlayers))
	}

	// Check order (highest score/min first)
	if topPlayers[0].Name != "TopPlayer" {
		t.Errorf("Expected TopPlayer first, got %s", topPlayers[0].Name)
	}
	if topPlayers[1].Name != "MidPlayer" {
		t.Errorf("Expected MidPlayer second, got %s", topPlayers[1].Name)
	}
	if topPlayers[2].Name != "LowPlayer" {
		t.Errorf("Expected LowPlayer third, got %s", topPlayers[2].Name)
	}

	// Check score per minute calculations (with tolerance)
	if topPlayers[0].ScorePerMin < 290 || topPlayers[0].ScorePerMin > 310 {
		t.Errorf("TopPlayer score/min expected ~300, got %.2f", topPlayers[0].ScorePerMin)
	}
}

func TestGetPlayerRank(t *testing.T) {
	testApp, ctx, _, match := testSetup(t)

	// Create 5 players with different scores
	playerIDs := make([]string, 5)
	joinTime := time.Now().Add(-10 * time.Minute)
	leftTime := time.Now()

	for i := 0; i < 5; i++ {
		steamID := fmt.Sprintf("steam_%d", i)
		playerName := fmt.Sprintf("Player%d", i)
		player := createTestPlayer(t, ctx, testApp, steamID, playerName, match, &joinTime)
		playerIDs[i] = player.ID

		// Scores: 5000, 4000, 3000, 2000, 1000
		score := (5 - i) * 1000

		updatePlayerStats(t, testApp, match.ID, player.ID, map[string]any{
			"score":           score,
			"first_joined_at": joinTime.Format("2006-01-02 15:04:05.000Z"),
			"last_left_at":    leftTime.Format("2006-01-02 15:04:05.000Z"),
		})
	}

	// Test rank for the middle player (should be rank 3 out of 5)
	rank, totalPlayers, err := GetPlayerRank(ctx, testApp, playerIDs[2])
	if err != nil {
		t.Fatalf("GetPlayerRank failed: %v", err)
	}

	if rank != 3 {
		t.Errorf("Expected rank 3, got %d", rank)
	}
	if totalPlayers != 5 {
		t.Errorf("Expected 5 total players, got %d", totalPlayers)
	}

	// Test rank for top player
	rank, totalPlayers, err = GetPlayerRank(ctx, testApp, playerIDs[0])
	if err != nil {
		t.Fatalf("GetPlayerRank failed for top player: %v", err)
	}

	if rank != 1 {
		t.Errorf("Expected rank 1 for top player, got %d", rank)
	}
}

func TestGetTopWeapons(t *testing.T) {
	testApp, ctx, _, match := testSetup(t)

	// Create test player (no need to add to match for weapon stats)
	player := createTestPlayer(t, ctx, testApp, "76561198000000001", "TestPlayer", nil, nil)

	// Add weapon stats
	weapons := []struct {
		name  string
		kills int64
	}{
		{"M4A1", 50},
		{"AK-47", 30},
		{"AWP", 20},
		{"Pistol", 10},
	}

	for _, w := range weapons {
		err := UpsertMatchWeaponStats(ctx, testApp, match.ID, player.ID, w.name, &w.kills, nil)
		if err != nil {
			t.Fatalf("Failed to create weapon stats for %s: %v", w.name, err)
		}
	}

	// Test GetTopWeapons (limit 3)
	topWeapons, err := GetTopWeapons(ctx, testApp, player.ID, 3)
	if err != nil {
		t.Fatalf("GetTopWeapons failed: %v", err)
	}

	if len(topWeapons) != 3 {
		t.Fatalf("Expected 3 weapons, got %d", len(topWeapons))
	}

	// Check order and values
	if topWeapons[0].Name != "M4A1" || topWeapons[0].Kills != 50 {
		t.Errorf("Expected M4A1 with 50 kills, got %s with %d", topWeapons[0].Name, topWeapons[0].Kills)
	}
	if topWeapons[1].Name != "AK-47" || topWeapons[1].Kills != 30 {
		t.Errorf("Expected AK-47 with 30 kills, got %s with %d", topWeapons[1].Name, topWeapons[1].Kills)
	}
	if topWeapons[2].Name != "AWP" || topWeapons[2].Kills != 20 {
		t.Errorf("Expected AWP with 20 kills, got %s with %d", topWeapons[2].Name, topWeapons[2].Kills)
	}

	// Test with empty player
	emptyPlayer := createTestPlayer(t, ctx, testApp, "76561198000000002", "EmptyPlayer", nil, nil)

	topWeapons, err = GetTopWeapons(ctx, testApp, emptyPlayer.ID, 10)
	if err != nil {
		t.Fatalf("GetTopWeapons failed for empty player: %v", err)
	}

	if len(topWeapons) != 0 {
		t.Errorf("Expected 0 weapons for empty player, got %d", len(topWeapons))
	}
}
