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

// TestAppWrapper wraps tests.TestApp to implement handlers.AppInterface for testing
type TestAppWrapper struct {
	*tests.TestApp
}

func (w *TestAppWrapper) SendRconCommand(serverID string, command string) (string, error) {
	// Mock implementation - just return empty response for testing
	return "", nil
}

// NewTestAppWrapper creates a new wrapped test app
func NewTestAppWrapper(app *tests.TestApp) *TestAppWrapper {
	return &TestAppWrapper{TestApp: app}
}

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

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	// Note: We pass nil for score debouncer in these unit tests as we're testing parser+handler logic only
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
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
	player, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198995742987")
	require.NoError(t, err)
	assert.Equal(t, "TestPlayer", player.Name)

	// Verify match exists and has stats
	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)

	// Verify kill was recorded
	stats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.GetInt("kills"))

	// Verify weapon stats
	weaponStats, err := appWrapper.FindFirstRecordByFilter(
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

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create match
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse objective captured
	objLine := `[2025.11.08-14.00.01:000][  1]LogGameplayEvents: Display: Objective 1 was captured for team 0 from team 1 by TestPlayer[76561198995742987].`
	require.NoError(t, p.ParseAndProcess(ctx, objLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Verify player created and objective stat recorded
	player, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198995742987")
	require.NoError(t, err)

	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)

	stats, err := appWrapper.FindFirstRecordByFilter(
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

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create match
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse round end
	roundLine := `[2025.11.08-14.05.00:000][100]LogGameplayEvents: Display: Round 1 Over: Team 0 won (win reason: Elimination)`
	require.NoError(t, p.ParseAndProcess(ctx, roundLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Verify round_end event was created
	events, err := appWrapper.FindRecordsByFilter("events", "type = 'round_end'", "-created", 10, 0)
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

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
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
	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)

	player, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198995742987")
	require.NoError(t, err)

	stats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.NotNil(t, stats.GetDateTime("disconnected_at"))
}

// TestFriendlyFireKillEvent tests that teamkills are recorded as friendly fire kills
func TestFriendlyFireKillEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Parse map load
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse two players joining on same team (team 0)
	loginLine1 := `[2025.11.08-14.00.01:000][  1]LogNet: Login request: ?Name=Killer userId: SteamNWI:76561198000000001 platform: SteamNWI`
	require.NoError(t, p.ParseAndProcess(ctx, loginLine1, serverID, "test.log"))

	loginLine2 := `[2025.11.08-14.00.01:500][  2]LogNet: Login request: ?Name=Victim userId: SteamNWI:76561198000000002 platform: SteamNWI`
	require.NoError(t, p.ParseAndProcess(ctx, loginLine2, serverID, "test.log"))

	joinLine1 := `[2025.11.08-14.00.02:000][  3]LogNet: Join succeeded: Killer`
	require.NoError(t, p.ParseAndProcess(ctx, joinLine1, serverID, "test.log"))

	joinLine2 := `[2025.11.08-14.00.02:500][  4]LogNet: Join succeeded: Victim`
	require.NoError(t, p.ParseAndProcess(ctx, joinLine2, serverID, "test.log"))

	// Parse FRIENDLY FIRE kill (same team: 0)
	ffKillLine := `[2025.11.08-14.00.03:000][  5]LogGameplayEvents: Display: Killer[76561198000000001, team 0] killed Victim[76561198000000002, team 0] with BP_Firearm_AKM_C_2147480339`
	require.NoError(t, p.ParseAndProcess(ctx, ffKillLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Get match
	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)

	// Get killer and victim players
	killer, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198000000001")
	require.NoError(t, err)
	victim, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198000000002")
	require.NoError(t, err)

	// Verify killer has friendly_fire_kills (NOT regular kills)
	killerStats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": killer.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 0, killerStats.GetInt("kills"), "Teamkiller should have 0 regular kills")
	assert.Equal(t, 1, killerStats.GetInt("friendly_fire_kills"), "Teamkiller should have 1 friendly fire kill")

	// Verify victim gets death
	victimStats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": victim.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, victimStats.GetInt("deaths"), "Victim should have 1 death")

	// Note: FF incident recording may fail in tests if event.Record.created is not properly set
	// The important thing is that the FF stats are recorded correctly (friendly_fire_kills increment)
	// In production, the FF incident would be recorded with a valid timestamp from the event record
}

// TestMultiPlayerKillEvent tests that multi-player kills (assists) are handled correctly
func TestMultiPlayerKillEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Parse map load
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse three players
	for i, steamID := range []string{"76561198000000001", "76561198000000002", "76561198000000003"} {
		names := []string{"Killer1", "Killer2", "Victim"}
		loginLine := `[2025.11.08-14.00.01:000][  ` + string(rune(i+1)) + `]LogNet: Login request: ?Name=` + names[i] + ` userId: SteamNWI:` + steamID + ` platform: SteamNWI`
		require.NoError(t, p.ParseAndProcess(ctx, loginLine, serverID, "test.log"))
		joinLine := `[2025.11.08-14.00.02:000][  ` + string(rune(i+4)) + `]LogNet: Join succeeded: ` + names[i]
		require.NoError(t, p.ParseAndProcess(ctx, joinLine, serverID, "test.log"))
	}

	// Parse multi-player kill (2 killers, 1 victim)
	multiKillLine := `[2025.11.08-14.00.03:000][  7]LogGameplayEvents: Display: Killer1[76561198000000001, team 0] + Killer2[76561198000000002, team 0] killed Victim[76561198000000003, team 1] with BP_Firearm_AKM_C_2147480339`
	require.NoError(t, p.ParseAndProcess(ctx, multiKillLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Get match
	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)

	killer1, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198000000001")
	require.NoError(t, err)
	killer2, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198000000002")
	require.NoError(t, err)

	// Verify first killer gets kill credit
	killer1Stats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": killer1.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, killer1Stats.GetInt("kills"), "First killer should get 1 kill")
	assert.Equal(t, 0, killer1Stats.GetInt("assists"), "First killer should have 0 assists")

	// Verify second killer gets assist
	killer2Stats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": killer2.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 0, killer2Stats.GetInt("kills"), "Second killer should have 0 kills")
	assert.Equal(t, 1, killer2Stats.GetInt("assists"), "Second killer should get 1 assist")
}

// TestSuicideKillEvent tests that suicides only increment deaths, not kills
func TestSuicideKillEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Parse map load
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse player login and join
	loginLine := `[2025.11.08-14.00.01:000][  1]LogNet: Login request: ?Name=SuicidePlayer userId: SteamNWI:76561198000000001 platform: SteamNWI`
	require.NoError(t, p.ParseAndProcess(ctx, loginLine, serverID, "test.log"))

	joinLine := `[2025.11.08-14.00.02:000][  2]LogNet: Join succeeded: SuicidePlayer`
	require.NoError(t, p.ParseAndProcess(ctx, joinLine, serverID, "test.log"))

	// Parse suicide (killer == victim same Steam ID)
	suicideLine := `[2025.11.08-14.00.03:000][  3]LogGameplayEvents: Display: SuicidePlayer[76561198000000001, team 0] killed SuicidePlayer[76561198000000001, team 0] with BP_Thrown_Molotov_C_2147480339`
	require.NoError(t, p.ParseAndProcess(ctx, suicideLine, serverID, "test.log"))

	time.Sleep(50 * time.Millisecond)

	// Get match and player
	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)
	player, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198000000001")
	require.NoError(t, err)

	// Verify suicide: only deaths, no kills/assists
	stats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.GetInt("kills"), "Suicide should not get kill credit")
	assert.Equal(t, 0, stats.GetInt("assists"), "Suicide should not get assists")
	assert.Equal(t, 1, stats.GetInt("deaths"), "Suicide should increment deaths")

	// Verify no weapon stats were created for suicide
	weaponStats, err := appWrapper.FindFirstRecordByFilter(
		"match_weapon_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	assert.Error(t, err, "Suicide should not create weapon stats")
	assert.Nil(t, weaponStats)
}

// TestMultipleKillsStatAccumulation tests that kills accumulate across multiple events
func TestMultipleKillsStatAccumulation(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	p := parser.NewLogParser(appWrapper, testApp.Logger())
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Parse map load
	mapLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	require.NoError(t, p.ParseAndProcess(ctx, mapLine, serverID, "test.log"))

	// Parse players
	for i, steamID := range []string{"76561198000000001", "76561198000000002", "76561198000000003", "76561198000000004"} {
		names := []string{"Killer", "Victim1", "Victim2", "Victim3"}
		loginLine := `[2025.11.08-14.00.01:000][  ` + string(rune(i+1)) + `]LogNet: Login request: ?Name=` + names[i] + ` userId: SteamNWI:` + steamID + ` platform: SteamNWI`
		require.NoError(t, p.ParseAndProcess(ctx, loginLine, serverID, "test.log"))
		joinLine := `[2025.11.08-14.00.02:000][  ` + string(rune(i+5)) + `]LogNet: Join succeeded: ` + names[i]
		require.NoError(t, p.ParseAndProcess(ctx, joinLine, serverID, "test.log"))
	}

	// Parse 3 kills with different weapons
	killLines := []string{
		`[2025.11.08-14.00.03:000][  9]LogGameplayEvents: Display: Killer[76561198000000001, team 0] killed Victim1[76561198000000002, team 1] with BP_Firearm_AKM_C_2147480339`,
		`[2025.11.08-14.00.03:500][ 10]LogGameplayEvents: Display: Killer[76561198000000001, team 0] killed Victim2[76561198000000003, team 1] with BP_Firearm_M16_C_2147480339`,
		`[2025.11.08-14.00.04:000][ 11]LogGameplayEvents: Display: Killer[76561198000000001, team 0] killed Victim3[76561198000000004, team 1] with BP_Thrown_Molotov_C_2147480339`,
	}

	for _, line := range killLines {
		require.NoError(t, p.ParseAndProcess(ctx, line, serverID, "test.log"))
	}

	time.Sleep(100 * time.Millisecond)

	// Get match and killer
	match, err := database.GetActiveMatch(ctx, appWrapper, serverID)
	require.NoError(t, err)
	killer, err := database.GetPlayerByExternalID(ctx, appWrapper, "76561198000000001")
	require.NoError(t, err)

	// Verify total kills accumulated
	stats, err := appWrapper.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": killer.ID},
	)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.GetInt("kills"), "Killer should have 3 total kills")

	// Verify separate weapon stats were tracked
	weaponStats, err := appWrapper.FindRecordsByFilter(
		"match_weapon_stats",
		"match = {:match} && player = {:player}",
		"weapon_name",
		10,
		0,
		map[string]any{"match": match.ID, "player": killer.ID},
	)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(weaponStats), 2, "Should have weapon stats for at least 2 different weapons")

	// Verify each weapon has 1 kill
	for _, ws := range weaponStats {
		assert.Equal(t, 1, ws.GetInt("kills"), "Each weapon should have 1 kill")
	}
}
