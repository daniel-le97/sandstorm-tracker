package integration

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/events"
	"sandstorm-tracker/internal/handlers"
	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create string pointers for test values
func stringPtr(s string) *string { return &s }

// TestMapLoadEvent tests the complete flow of a map load event
func TestMapLoadEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create a map_load event
	eventsCollection, err := testApp.FindCollectionByNameOrId("events")
	require.NoError(t, err)

	serverRecord, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID})
	require.NoError(t, err)

	mapLoadEvent := core.NewRecord(eventsCollection)
	mapLoadEvent.Set("type", events.TypeMapLoad)
	mapLoadEvent.Set("server", serverRecord.Id)
	mapLoadEvent.Set("timestamp", time.Now())

	mapLoadData := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now(),
		PlayerTeam: nil,
		IsCatchup:  false,
	}
	dataJSON, _ := json.Marshal(mapLoadData)
	mapLoadEvent.Set("data", string(dataJSON))

	err = testApp.Save(mapLoadEvent)
	require.NoError(t, err)

	// Verify a match was created
	matches, err := testApp.FindRecordsByFilter("matches", "server = {:serverId}", "-created", 1, 0, map[string]any{"serverId": serverRecord.Id})
	require.NoError(t, err)
	require.Len(t, matches, 1)

	match := matches[0]
	assert.Equal(t, "Ministry", match.GetString("map"))
	assert.Equal(t, "Checkpoint", match.GetString("mode"))

	// Verify match_start event was emitted
	startEvents, err := testApp.FindRecordsByFilter("events", "type = 'match_start'", "-created", 1, 0)
	require.NoError(t, err)
	require.Len(t, startEvents, 1)

	startEvent := startEvents[0]
	var startData events.MatchStartData
	err = json.Unmarshal([]byte(startEvent.GetString("data")), &startData)
	require.NoError(t, err)
	assert.Equal(t, match.Id, startData.MatchID)
	assert.Equal(t, "Ministry", startData.Map)
}

// TestMapTravelEvent tests map travel creating a new match
func TestMapTravelEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create first match via map_load
	eventsCollection, err := testApp.FindCollectionByNameOrId("events")
	require.NoError(t, err)

	serverRecord, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID})
	require.NoError(t, err)

	mapLoadEvent := core.NewRecord(eventsCollection)
	mapLoadEvent.Set("type", events.TypeMapLoad)
	mapLoadEvent.Set("server", serverRecord.Id)
	mapLoadEvent.Set("timestamp", time.Now())

	mapLoadData := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now(),
		PlayerTeam: nil,
		IsCatchup:  false,
	}
	dataJSON, _ := json.Marshal(mapLoadData)
	mapLoadEvent.Set("data", string(dataJSON))
	err = testApp.Save(mapLoadEvent)
	require.NoError(t, err)

	// Get all matches after map load
	allMatches1, err := testApp.FindRecordsByFilter("matches", "server = {:serverId}", "-created", 100, 0, map[string]any{"serverId": serverRecord.Id})
	require.NoError(t, err)
	require.Len(t, allMatches1, 1)
	firstMatchID := allMatches1[0].Id
	firstMatchMap := allMatches1[0].GetString("map")
	assert.Equal(t, "Ministry", firstMatchMap)

	// Now create map_travel event to different map
	mapTravelEvent := core.NewRecord(eventsCollection)
	mapTravelEvent.Set("type", events.TypeMapTravel)
	mapTravelEvent.Set("server", serverRecord.Id)
	mapTravelEvent.Set("timestamp", time.Now().Add(5*time.Minute))

	mapTravelData := events.MapTravelData{
		Map:        "Oilfield",
		Scenario:   "Scenario_Refinery_Push_Insurgents",
		Timestamp:  time.Now().Add(5 * time.Minute),
		PlayerTeam: nil,
		IsCatchup:  false,
	}
	dataJSON, _ = json.Marshal(mapTravelData)
	mapTravelEvent.Set("data", string(dataJSON))
	err = testApp.Save(mapTravelEvent)
	require.NoError(t, err)

	// Get all matches after map travel
	allMatches2, err := testApp.FindRecordsByFilter("matches", "server = {:serverId}", "-created", 100, 0, map[string]any{"serverId": serverRecord.Id})
	require.NoError(t, err)
	// Should have at least 1 active match (the new one), first might be deleted if empty
	require.GreaterOrEqual(t, len(allMatches2), 1)

	// Find the new match (should be the most recent one with a different map)
	var newMatch *core.Record
	for _, m := range allMatches2 {
		if m.GetString("map") == "Oilfield" {
			newMatch = m
			break
		}
	}
	require.NotNil(t, newMatch, "Should have created a new match for Oilfield")
	assert.NotEqual(t, firstMatchID, newMatch.Id)
	assert.Equal(t, "Push", newMatch.GetString("mode"))

	// Verify match_start event for new match
	startEvents, err := testApp.FindRecordsByFilter("events", "type = 'match_start'", "-created", 1, 0)
	require.NoError(t, err)
	require.Len(t, startEvents, 1)
}

// TestGameOverEvent tests game over ending a match gracefully
func TestGameOverEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server and match
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create map_load event first
	eventsCollection, err := testApp.FindCollectionByNameOrId("events")
	require.NoError(t, err)

	serverRecord, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID})
	require.NoError(t, err)

	mapLoadEvent := core.NewRecord(eventsCollection)
	mapLoadEvent.Set("type", events.TypeMapLoad)
	mapLoadEvent.Set("server", serverRecord.Id)
	mapLoadEvent.Set("timestamp", time.Now())

	mapLoadData := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now(),
		PlayerTeam: nil,
		IsCatchup:  false,
	}
	dataJSON, _ := json.Marshal(mapLoadData)
	mapLoadEvent.Set("data", string(dataJSON))
	err = testApp.Save(mapLoadEvent)
	require.NoError(t, err)

	// Add a player to the match
	player, err := database.GetOrCreatePlayerBySteamID(ctx, testApp, "76561198995742987", "TestPlayer")
	require.NoError(t, err)

	activeMatch, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)

	team := int64(0)
	err = database.UpsertMatchPlayerStats(ctx, testApp, activeMatch.ID, player.ID, &team, nil)
	require.NoError(t, err)

	// Verify player is in the match
	playerStats, err := testApp.FindRecordsByFilter("match_player_stats", "match = {:matchId} && player = {:playerId}", "", 1, 0, map[string]any{"matchId": activeMatch.ID, "playerId": player.ID})
	require.NoError(t, err)
	require.Len(t, playerStats, 1)
	assert.True(t, playerStats[0].GetBool("is_currently_connected"))

	// Create game_over event
	gameOverEvent := core.NewRecord(eventsCollection)
	gameOverEvent.Set("type", events.TypeGameOver)
	gameOverEvent.Set("server", serverRecord.Id)
	gameOverEvent.Set("timestamp", time.Now().Add(10*time.Minute))

	gameOverData := events.GameOverData{
		Timestamp: time.Now().Add(10 * time.Minute),
	}
	dataJSON, _ = json.Marshal(gameOverData)
	gameOverEvent.Set("data", string(dataJSON))
	err = testApp.Save(gameOverEvent)
	require.NoError(t, err)

	// Verify match is ended
	endedMatch, err := testApp.FindRecordById("matches", activeMatch.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, endedMatch.GetString("end_time"), "Match should be ended")
	assert.Equal(t, "finished", endedMatch.GetString("status"), "Match status should be set to 'finished'")

	// Verify player is disconnected
	playerStats, err = testApp.FindRecordsByFilter("match_player_stats", "match = {:matchId} && player = {:playerId}", "", 1, 0, map[string]any{"matchId": activeMatch.ID, "playerId": player.ID})
	require.NoError(t, err)
	require.Len(t, playerStats, 1)
	assert.False(t, playerStats[0].GetBool("is_currently_connected"), "Player should be marked as disconnected")

	// Verify match_end event was emitted
	endEvents, err := testApp.FindRecordsByFilter("events", "type = 'match_end'", "-created", 1, 0)
	require.NoError(t, err)
	require.Len(t, endEvents, 1)

	endEvent := endEvents[0]
	var endData events.MatchEndData
	err = json.Unmarshal([]byte(endEvent.GetString("data")), &endData)
	require.NoError(t, err)
	assert.Equal(t, activeMatch.ID, endData.MatchID)
}

// TestGameOverWithMultiplePlayers tests game over disconnects all players
func TestGameOverWithMultiplePlayers(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server and match
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create match
	eventsCollection, err := testApp.FindCollectionByNameOrId("events")
	require.NoError(t, err)

	serverRecord, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID})
	require.NoError(t, err)

	mapLoadEvent := core.NewRecord(eventsCollection)
	mapLoadEvent.Set("type", events.TypeMapLoad)
	mapLoadEvent.Set("server", serverRecord.Id)
	mapLoadEvent.Set("timestamp", time.Now())

	mapLoadData := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now(),
		PlayerTeam: nil,
		IsCatchup:  false,
	}
	dataJSON, _ := json.Marshal(mapLoadData)
	mapLoadEvent.Set("data", string(dataJSON))
	err = testApp.Save(mapLoadEvent)
	require.NoError(t, err)

	activeMatch, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)

	// Add 3 players to the match
	team := int64(0)
	for i := 1; i <= 3; i++ {
		steamID := "7656119899574298" + string(rune('0'+i))
		player, err := database.GetOrCreatePlayerBySteamID(ctx, testApp, steamID, "Player"+string(rune('A'+i-1)))
		require.NoError(t, err)

		err = database.UpsertMatchPlayerStats(ctx, testApp, activeMatch.ID, player.ID, &team, nil)
		require.NoError(t, err)
	}

	// Verify all 3 players are connected
	allStats, err := testApp.FindRecordsByFilter("match_player_stats", "match = {:matchId}", "", 100, 0, map[string]any{"matchId": activeMatch.ID})
	require.NoError(t, err)
	require.Len(t, allStats, 3)
	for _, stat := range allStats {
		assert.True(t, stat.GetBool("is_currently_connected"))
	}

	// Game over
	gameOverEvent := core.NewRecord(eventsCollection)
	gameOverEvent.Set("type", events.TypeGameOver)
	gameOverEvent.Set("server", serverRecord.Id)
	gameOverEvent.Set("timestamp", time.Now().Add(10*time.Minute))

	gameOverData := events.GameOverData{
		Timestamp: time.Now().Add(10 * time.Minute),
	}
	dataJSON, _ = json.Marshal(gameOverData)
	gameOverEvent.Set("data", string(dataJSON))
	err = testApp.Save(gameOverEvent)
	require.NoError(t, err)

	// Verify all players are disconnected
	allStats, err = testApp.FindRecordsByFilter("match_player_stats", "match = {:matchId}", "", 100, 0, map[string]any{"matchId": activeMatch.ID})
	require.NoError(t, err)
	require.Len(t, allStats, 3)
	for _, stat := range allStats {
		assert.False(t, stat.GetBool("is_currently_connected"), "All players should be disconnected after game over")
	}

	// Verify match status is set to finished
	finishedMatch, err := testApp.FindRecordById("matches", activeMatch.ID)
	require.NoError(t, err)
	assert.Equal(t, "finished", finishedMatch.GetString("status"), "Match status should be set to 'finished'")
}

// TestLogFileCreatedEvent tests log file creation (server restart recovery)
func TestLogFileCreatedEvent(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server"

	// Setup: Create server and match
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create a match via map_load
	eventsCollection, err := testApp.FindCollectionByNameOrId("events")
	require.NoError(t, err)

	serverRecord, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID})
	require.NoError(t, err)

	mapLoadEvent := core.NewRecord(eventsCollection)
	mapLoadEvent.Set("type", events.TypeMapLoad)
	mapLoadEvent.Set("server", serverRecord.Id)
	mapLoadEvent.Set("timestamp", time.Now())

	mapLoadData := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now(),
		PlayerTeam: nil,
		IsCatchup:  false,
	}
	dataJSON, _ := json.Marshal(mapLoadData)
	mapLoadEvent.Set("data", string(dataJSON))
	err = testApp.Save(mapLoadEvent)
	require.NoError(t, err)

	// Verify match was created
	activeMatch, err := database.GetActiveMatch(ctx, testApp, serverID)
	require.NoError(t, err)
	require.NotNil(t, activeMatch)
	matchID := activeMatch.ID

	// Create log file created event (simulating server restart)
	logFileTime := time.Now().Add(5 * time.Minute)
	logFileEvent := core.NewRecord(eventsCollection)
	logFileEvent.Set("type", events.TypeLogFileCreated)
	logFileEvent.Set("server", serverRecord.Id)
	logFileEvent.Set("timestamp", logFileTime)

	logFileData := events.LogFileCreatedData{
		Timestamp: logFileTime,
	}
	dataJSON, _ = json.Marshal(logFileData)
	logFileEvent.Set("data", string(dataJSON))
	err = testApp.Save(logFileEvent)
	require.NoError(t, err)

	// Verify file_creation_time was updated on server
	updatedServer, err := testApp.FindRecordById("servers", serverRecord.Id)
	require.NoError(t, err)
	fileCreationTime := updatedServer.GetString("log_file_creation_time")
	assert.NotEmpty(t, fileCreationTime, "Server file_creation_time should be updated")

	// Verify the stale match was ended with crashed status
	endedMatch, err := testApp.FindRecordById("matches", matchID)
	require.NoError(t, err)
	assert.NotEmpty(t, endedMatch.GetString("end_time"), "Stale match should be ended after log file creation")
	assert.Equal(t, "crashed", endedMatch.GetString("status"), "Match status should be set to 'crashed'")

	log.Printf("[TEST] Log file created event successfully ended stale match %s with crashed status", matchID)
}

// TestRoundEndWinnerTeam tests that winner_team is set in handleMatchEnd only when winning team matches player_team
func TestRoundEndWinnerTeam(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server-round-end"

	// Setup: Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Test Server", "/path")
	require.NoError(t, err)

	appWrapper := NewTestAppWrapper(testApp)
	gameHandlers := handlers.NewGameEventHandlers(appWrapper, nil)
	gameHandlers.RegisterHooks()

	// Create a map load event with Security as player_team (Scenario_Ministry_Checkpoint_Security)
	eventsCollection, err := testApp.FindCollectionByNameOrId("events")
	require.NoError(t, err)

	serverRecord, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID})
	require.NoError(t, err)

	mapLoadEvent := core.NewRecord(eventsCollection)
	mapLoadEvent.Set("type", events.TypeMapLoad)
	mapLoadEvent.Set("server", serverRecord.Id)
	mapLoadEvent.Set("timestamp", time.Now())

	mapLoadData := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now(),
		PlayerTeam: stringPtr("Security"),
		IsCatchup:  false,
	}
	dataJSON, _ := json.Marshal(mapLoadData)
	mapLoadEvent.Set("data", string(dataJSON))
	err = testApp.Save(mapLoadEvent)
	require.NoError(t, err)

	// Get the created match
	matches, err := testApp.FindRecordsByFilter("matches", "server = {:serverId} && end_time = ''", "-start_time", 1, 0, map[string]any{"serverId": serverRecord.Id})
	require.NoError(t, err)
	require.Len(t, matches, 1)

	match := matches[0]
	matchID := match.Id
	mode := match.GetString("mode")
	playerTeam := match.GetString("player_team")
	assert.Equal(t, "Checkpoint", mode, "Match mode should be Checkpoint")
	assert.Equal(t, "Security", playerTeam, "Match player_team should be Security")

	// Test 1: Team 0 (Security) wins - should set winner_team since it matches player_team
	roundEndEvent1 := core.NewRecord(eventsCollection)
	roundEndEvent1.Set("type", events.TypeRoundEnd)
	roundEndEvent1.Set("server", serverRecord.Id)
	roundEndEvent1.Set("timestamp", time.Now())

	roundEndData1 := events.RoundEndData{
		MatchID:     matchID,
		RoundNumber: 1,
		WinningTeam: 0, // Team 0 = Security (matches player_team)
		IsCatchup:   false,
	}
	dataJSON, _ = json.Marshal(roundEndData1)
	roundEndEvent1.Set("data", string(dataJSON))
	err = testApp.Save(roundEndEvent1)
	require.NoError(t, err)

	// Send MatchEnd event to trigger winner_team logic in handleMatchEnd
	matchEndEvent1 := core.NewRecord(eventsCollection)
	matchEndEvent1.Set("type", events.TypeMatchEnd)
	matchEndEvent1.Set("server", serverRecord.Id)
	matchEndEvent1.Set("timestamp", time.Now().Add(10*time.Second))

	matchEndData1 := events.MatchEndData{
		MatchID: matchID,
		EndTime: time.Now().Add(10 * time.Second),
	}
	dataJSON, _ = json.Marshal(matchEndData1)
	matchEndEvent1.Set("data", string(dataJSON))
	err = testApp.Save(matchEndEvent1)
	require.NoError(t, err)

	// Verify winner_team was set to 0 (Security) by handleMatchEnd
	updatedMatch1, err := testApp.FindRecordById("matches", matchID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), int64(updatedMatch1.GetInt("winner_team")), "Winner_team should be set to 0 when Security wins and player_team is Security")

	// Test 2: Start a new match with same server for the second test
	mapLoadEvent2 := core.NewRecord(eventsCollection)
	mapLoadEvent2.Set("type", events.TypeMapLoad)
	mapLoadEvent2.Set("server", serverRecord.Id)
	mapLoadEvent2.Set("timestamp", time.Now().Add(1*time.Minute))

	mapLoadData2 := events.MapLoadData{
		Map:        "Ministry",
		Scenario:   "Scenario_Ministry_Checkpoint_Security",
		Timestamp:  time.Now().Add(1 * time.Minute),
		PlayerTeam: stringPtr("Security"),
		IsCatchup:  false,
	}
	dataJSON, _ = json.Marshal(mapLoadData2)
	mapLoadEvent2.Set("data", string(dataJSON))
	err = testApp.Save(mapLoadEvent2)
	require.NoError(t, err)

	// Get the new match
	matches2, err := testApp.FindRecordsByFilter("matches", "server = {:serverId} && end_time = '' && start_time != {:oldTime}", "-start_time", 1, 0, map[string]any{"serverId": serverRecord.Id, "oldTime": match.GetString("start_time")})
	require.NoError(t, err)
	require.Len(t, matches2, 1)

	match2 := matches2[0]
	matchID2 := match2.Id

	// Test 2: Team 1 (Insurgents) wins - should NOT set winner_team since it doesn't match player_team
	roundEndEvent2 := core.NewRecord(eventsCollection)
	roundEndEvent2.Set("type", events.TypeRoundEnd)
	roundEndEvent2.Set("server", serverRecord.Id)
	roundEndEvent2.Set("timestamp", time.Now().Add(2*time.Minute))

	roundEndData2 := events.RoundEndData{
		MatchID:     matchID2,
		RoundNumber: 1,
		WinningTeam: 1, // Team 1 = Insurgents (does NOT match player_team)
		IsCatchup:   false,
	}
	dataJSON, _ = json.Marshal(roundEndData2)
	roundEndEvent2.Set("data", string(dataJSON))
	err = testApp.Save(roundEndEvent2)
	require.NoError(t, err)

	// Verify winner_team was NOT updated (still 0, not 1)
	updatedMatch2, err := testApp.FindRecordById("matches", matchID2)
	require.NoError(t, err)
	assert.Equal(t, int64(0), int64(updatedMatch2.GetInt("winner_team")), "Winner_team should remain 0 when Insurgents win since player_team is Security")

	log.Printf("[TEST] Round end events correctly set winner_team only when winning team matches player_team")
}
