package integration

import (
	"context"
	"encoding/json"
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

// TestKillEventFlow tests the complete flow from parsing a kill to database updates
func TestKillEventFlow(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server and match
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	p := parser.NewLogParser(testApp, testApp.Logger())
	// Note: We pass nil for score debouncer in these unit tests as we're testing parser+handler logic only
	gameHandlers := handlers.NewGameEventHandlers(testApp, nil)
	gameHandlers.RegisterHooks()

	// Parse map load to create match
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse player login
	loginLine := `[2025.11.08-14.00.01:000][  1]LogNet: Login request: ?Name=TestPlayer userId: SteamNWI:76561198995742987 platform: SteamNWI`
	require.NoError(t, p.ParseAndProcess(ctx, loginLine, serverID, "test.log"))

	// Parse player join
	joinLine := `[2025.11.08-14.00.02:000][  2]LogNet: Join succeeded: TestPlayer`
	require.NoError(t, p.ParseAndProcess(ctx, joinLine, serverID, "test.log"))

	// Parse kill event
	killLine := `[2025.11.08-14.00.03:000][  3]LogGameplayEvents: Display: TestPlayer[76561198995742987, team 0] killed Bot[INVALID, team 1] with BP_Firearm_AKM_C_2147480339`
	require.NoError(t, p.ParseAndProcess(ctx, killLine, serverID, "test.log"))

	// Wait for event processing
	time.Sleep(50 * time.Millisecond)

	// Verify player was created
	player, err := database.GetPlayerByExternalID(ctx, testApp, "76561198995742987")
	require.NoError(t, err)
	assert.Equal(t, "TestPlayer", player.Name)

	// Verify match exists and has stats
	match, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)

	// Verify kill was recorded
	stats, err := testApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.GetInt("kills"))

	// Verify weapon stats
	weaponStats, err := testApp.FindFirstRecordByFilter(
		"match_weapon_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, weaponStats.GetString("weapon_name")) // Should be "AKM" if parser sets it
	assert.Equal(t, 1, weaponStats.GetInt("kills"))
}

// TestObjectiveEventFlow tests objective capture/destroy events
func TestObjectiveEventFlow(t *testing.T) {
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

	// Create match
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse objective captured
	objLine := `[2025.11.08-14.00.01:000][  1]LogGameplayEvents: Display: Objective 1 was captured for team 0 from team 1 by TestPlayer[76561198995742987].`
	require.NoError(t, p.ParseAndProcess(ctx, objLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Verify player created and objective stat recorded
	player, err := database.GetPlayerByExternalID(ctx, testApp, "76561198995742987")
	require.NoError(t, err)

	match, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)

	stats, err := testApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.GetInt("objectives_captured"))
}

// TestRoundEndFlow tests round end triggers scoring
func TestRoundEndFlow(t *testing.T) {
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

	// Parse round end
	roundLine := `[2025.11.08-14.05.00:000][100]LogGameplayEvents: Display: Round 1 Over: Team 0 won (win reason: Elimination)`
	require.NoError(t, p.ParseAndProcess(ctx, roundLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Verify round_end event was created
	events, err := testApp.FindRecordsByFilter("events", "type = 'round_end'", "-created", 10, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)

	var data map[string]interface{}
	json.Unmarshal([]byte(events[0].GetString("data")), &data)
	assert.Equal(t, float64(0), data["winning_team"])
}

// TestPlayerLeaveFlow tests disconnect tracking
func TestPlayerLeaveFlow(t *testing.T) {
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

	// Create match and add player
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	loginLine := `[2025.11.08-14.00.01:000][  1]LogNet: Login request: ?Name=TestPlayer userId: SteamNWI:76561198995742987 platform: SteamNWI`
	require.NoError(t, p.ParseAndProcess(ctx, loginLine, serverID, "test.log"))

	joinLine := `[2025.11.08-14.00.02:000][  2]LogNet: Join succeeded: TestPlayer`
	require.NoError(t, p.ParseAndProcess(ctx, joinLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Parse disconnect
	dcLine := `[2025.11.08-14.10.00:000][500]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198995742987), Result: (EOS_Success)`
	require.NoError(t, p.ParseAndProcess(ctx, dcLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Verify player was disconnected from match
	match, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)

	player, err := database.GetPlayerByExternalID(ctx, testApp, "76561198995742987")
	require.NoError(t, err)

	stats, err := testApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.NotNil(t, stats.GetDateTime("disconnected_at"))
}
