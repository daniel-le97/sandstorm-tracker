package integration

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/handlers"
	"sandstorm-tracker/internal/parser"
	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullGameSession tests processing a complete game log file
func TestFullGameSession(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	p := parser.NewLogParser(testApp, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(testApp, nil)
	gameHandlers.RegisterHooks()

	// Process log file
	logPath := filepath.Join("..", "..", "internal", "parser", "test_data", "full-2.log")
	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	linesProcessed := 0
	for scanner.Scan() {
		line := scanner.Text()
		p.ParseAndProcess(ctx, line, serverID, "test.log")
		linesProcessed++
	}
	require.NoError(t, scanner.Err())

	// Wait for all events to process
	time.Sleep(200 * time.Millisecond)

	t.Logf("Processed %d lines", linesProcessed)

	// Verify player was created
	player, err := database.GetPlayerByExternalID(ctx, testApp, "76561198995742987")
	require.NoError(t, err, "Player ArmoredBear should exist")
	assert.Equal(t, "ArmoredBear", player.Name)

	// Verify matches were created (initial + map travel)
	matches, err := testApp.FindRecordsByFilter("matches", "", "-created", 100, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(matches), 1, "At least one match should exist")

	// Verify kills were recorded
	stats, err := testApp.FindRecordsByFilter(
		"match_player_stats",
		"player = {:player}",
		"-created",
		100,
		0,
		map[string]any{"player": player.ID},
	)
	require.NoError(t, err)

	totalKills := 0
	for _, stat := range stats {
		totalKills += stat.GetInt("kills")
	}
	assert.Greater(t, totalKills, 0, "Player should have kills recorded")

	// Verify weapon stats exist
	weaponStats, err := testApp.FindRecordsByFilter(
		"match_weapon_stats",
		"player = {:player}",
		"-created",
		100,
		0,
		map[string]any{"player": player.ID},
	)
	require.NoError(t, err)
	assert.Greater(t, len(weaponStats), 0, "Weapon stats should be recorded")
}

// TestMultiPlayerGame tests a game with multiple players
func TestMultiPlayerGame(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	p := parser.NewLogParser(testApp, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(testApp, nil)
	gameHandlers.RegisterHooks()

	// Process HC log (has multiple players)
	logPath := filepath.Join("..", "..", "internal", "parser", "test_data", "hc.log")
	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p.ParseAndProcess(ctx, scanner.Text(), serverID, "test.log")
	}
	require.NoError(t, scanner.Err())

	time.Sleep(300 * time.Millisecond)

	// Verify multiple players exist
	players, err := testApp.FindRecordsByFilter("players", "", "-created", 100, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(players), 3, "Multiple players should be created")

	// Verify total kills across all players
	allStats, err := testApp.FindRecordsByFilter("match_player_stats", "", "-created", 1000, 0)
	require.NoError(t, err)

	totalKills := 0
	for _, stat := range allStats {
		totalKills += stat.GetInt("kills")
	}
	assert.Greater(t, totalKills, 50, "Multiple kills should be recorded")
}

// TestWeaponStatsAggregation tests that weapon variants are aggregated correctly
func TestWeaponStatsAggregation(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	p := parser.NewLogParser(testApp, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(testApp, nil)
	gameHandlers.RegisterHooks()

	// Create match
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse kills with same weapon but different IDs
	killLine1 := `[2025.11.08-14.00.01:000][  1]LogGameplayEvents: Display: TestPlayer[76561198995742987, team 0] killed Bot1[INVALID, team 1] with BP_Firearm_AKM_C_2147480339`
	require.NoError(t, p.ParseAndProcess(ctx, killLine1, serverID, "test.log"))

	killLine2 := `[2025.11.08-14.00.02:000][  2]LogGameplayEvents: Display: TestPlayer[76561198995742987, team 0] killed Bot2[INVALID, team 1] with BP_Firearm_AKM_C_9999999999`
	require.NoError(t, p.ParseAndProcess(ctx, killLine2, serverID, "test.log"))

	time.Sleep(100 * time.Millisecond)

	// Verify weapon stats are aggregated under "AKM"
	player, err := database.GetPlayerByExternalID(ctx, testApp, "76561198995742987")
	require.NoError(t, err)

	match, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)

	weaponStats, err := testApp.FindRecordsByFilter(
		"match_weapon_stats",
		"match = {:match} && player = {:player}",
		"-created",
		100,
		0,
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	require.Len(t, weaponStats, 1, "Both kills should aggregate to one weapon record")
	assert.NotEmpty(t, weaponStats[0].GetString("weapon_name")) // Should be "AKM" if parser sets it
	assert.Equal(t, 2, weaponStats[0].GetInt("kills"))
}
