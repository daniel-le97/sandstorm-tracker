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
