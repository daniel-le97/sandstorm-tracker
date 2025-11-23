package parser

import (
	"context"
	"encoding/json"
	"sandstorm-tracker/internal/database"
	"testing"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

func TestMapLoadEvents(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()

	t.Run("Hideout_Hardcore_Checkpoint_Security", func(t *testing.T) {
		serverExternalID := "test-server-hideout"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Hideout Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.10.04-21.18.15:445][  0]LogLoad: LoadMap: /Game/Maps/Town/Town?Name=Player?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=10?Game=CheckpointHardcore?Lighting=Day`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify map_load event was created with correct lighting
		events, err := testApp.FindRecordsByFilter("events", "type = 'map_load'", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find map_load event: %v", err)
		}

		event := events[0]
		eventData := event.GetString("data")
		if eventData == "" {
			t.Error("Event data is empty")
		}

		// Parse event data and verify lighting
		var mapLoadData map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &mapLoadData); err != nil {
			t.Fatalf("failed to parse event data: %v", err)
		}

		if lighting, ok := mapLoadData["lighting"].(string); !ok || lighting != "Day" {
			t.Errorf("Expected lighting 'Day', got %v", mapLoadData["lighting"])
		}
	})

	t.Run("Hideout_Hardcore_Checkpoint_Security_Legacy", func(t *testing.T) {
		serverExternalID := "test-server-hideout-legacy"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Hideout Server Legacy", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.10.04-21.18.15:445][  0]LogLoad: LoadMap: /Game/Maps/Town/Town?Name=Player?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=10?Game=CheckpointHardcore`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify MAP_LOAD event was created (matches are now created by handlers in response to events)
		events, err := testApp.FindRecordsByFilter("events", "", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find event: %v", err)
		}

		event := events[0]
		if event.GetString("type") != "map_load" {
			t.Errorf("Expected event type 'map_load', got %s", event.GetString("type"))
		}

		// Verify event data contains the correct map info
		eventData := event.GetString("data")
		if eventData == "" {
			t.Error("Event data is empty")
		}
	})

	t.Run("Ministry_Checkpoint_Security", func(t *testing.T) {
		serverExternalID := "test-server-ministry"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Ministry Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")

		events, err := testApp.FindRecordsByFilter("events", "", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find event")
		}

		event := events[0]
		if event.GetString("type") != "map_load" {
			t.Errorf("Expected event type 'map_load', got %s", event.GetString("type"))
		}
	})

	t.Run("Insurgents_Team", func(t *testing.T) {
		serverExternalID := "test-server-insurgents"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Insurgents Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Insurgents?MaxPlayers=8?Lighting=Day`
		parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")

		events, err := testApp.FindRecordsByFilter("events", "", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find event")
		}

		event := events[0]
		if event.GetString("type") != "map_load" {
			t.Errorf("Expected event type 'map_load', got %s", event.GetString("type"))
		}
	})

	t.Run("Night_Lighting", func(t *testing.T) {
		serverExternalID := "test-server-night"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Night Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Town/Town?Name=Player?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=10?Game=CheckpointHardcore?Lighting=Night`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify map_load event was created with correct lighting
		events, err := testApp.FindRecordsByFilter("events", "type = 'map_load'", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find map_load event: %v", err)
		}

		event := events[0]
		eventData := event.GetString("data")
		if eventData == "" {
			t.Error("Event data is empty")
		}

		// Parse event data and verify lighting is Night
		var mapLoadData map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &mapLoadData); err != nil {
			t.Fatalf("failed to parse event data: %v", err)
		}

		if lighting, ok := mapLoadData["lighting"].(string); !ok || lighting != "Night" {
			t.Errorf("Expected lighting 'Night', got %v", mapLoadData["lighting"])
		}
	})

	t.Run("MapTravel_With_Lighting", func(t *testing.T) {
		serverExternalID := "test-server-map-travel"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Map Travel Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-14.00.00:000][  0]LogGameMode: ProcessServerTravel: Ministry?Scenario=Scenario_Ministry_Checkpoint_Security?Game=CheckpointHardcore?Lighting=Day`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify map_travel event was created with correct lighting
		events, err := testApp.FindRecordsByFilter("events", "type = 'map_travel'", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find map_travel event: %v", err)
		}

		event := events[0]
		eventData := event.GetString("data")
		if eventData == "" {
			t.Error("Event data is empty")
		}

		// Parse event data and verify lighting
		var mapTravelData map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &mapTravelData); err != nil {
			t.Fatalf("failed to parse event data: %v", err)
		}

		if lighting, ok := mapTravelData["lighting"].(string); !ok || lighting != "Day" {
			t.Errorf("Expected lighting 'Day', got %v", mapTravelData["lighting"])
		}
	})

	t.Run("MapTravel_Without_Lighting", func(t *testing.T) {
		serverExternalID := "test-server-map-travel-no-light"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Map Travel No Light Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-14.00.00:000][  0]LogGameMode: ProcessServerTravel: Ministry?Scenario=Scenario_Ministry_Checkpoint_Security?Game=CheckpointHardcore`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify map_travel event was created without lighting
		events, err := testApp.FindRecordsByFilter("events", "type = 'map_travel'", "-created", 1, 0)
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find map_travel event: %v", err)
		}

		event := events[0]
		eventData := event.GetString("data")
		if eventData == "" {
			t.Error("Event data is empty")
		}

		// Parse event data and verify lighting is nil
		var mapTravelData map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &mapTravelData); err != nil {
			t.Fatalf("failed to parse event data: %v", err)
		}

		if lighting := mapTravelData["lighting"]; lighting != nil {
			t.Errorf("Expected lighting to be nil for legacy log format, got %v", lighting)
		}
	})

	t.Run("MapLoad_Without_Lighting_No_Game", func(t *testing.T) {
		serverExternalID := "test-server-no-game-no-light"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "No Game No Light Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Get server record for filtering
		serverRec, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverExternalID})
		if err != nil {
			t.Fatalf("failed to find server record: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify map_load event was created without lighting or game - filter by server to isolate this test
		events, err := testApp.FindRecordsByFilter("events", "type = 'map_load' && server = {:serverId}", "-created", 1, 0, map[string]any{"serverId": serverRec.Id})
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find map_load event: %v", err)
		}

		event := events[0]
		if event.GetString("type") != "map_load" {
			t.Errorf("Expected event type 'map_load', got %s", event.GetString("type"))
		}

		// Verify event data
		eventData := event.GetString("data")
		if eventData == "" {
			t.Error("Event data is empty")
		}

		var mapLoadData map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &mapLoadData); err != nil {
			t.Fatalf("failed to parse event data: %v", err)
		}

		// Lighting should be nil when not provided
		if lighting := mapLoadData["lighting"]; lighting != nil {
			t.Errorf("Expected lighting to be nil when not provided, got %v", lighting)
		}
	})

	t.Run("MapLoad_Multiple_Servers_Different_Lighting", func(t *testing.T) {
		// Test that multiple servers with different lighting work correctly
		serverID1 := "test-server-day-multi"
		serverID2 := "test-server-night-multi"
		serverID3 := "test-server-legacy-multi"

		_, err := database.GetOrCreateServer(ctx, testApp, serverID1, "Day Server", "test/path1")
		if err != nil {
			t.Fatalf("failed to create server 1: %v", err)
		}
		_, err = database.GetOrCreateServer(ctx, testApp, serverID2, "Night Server", "test/path2")
		if err != nil {
			t.Fatalf("failed to create server 2: %v", err)
		}
		_, err = database.GetOrCreateServer(ctx, testApp, serverID3, "Legacy Server", "test/path3")
		if err != nil {
			t.Fatalf("failed to create server 3: %v", err)
		}

		// Get server records for filtering
		server1, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID1})
		if err != nil {
			t.Fatalf("failed to find server 1 record: %v", err)
		}
		server2, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID2})
		if err != nil {
			t.Fatalf("failed to find server 2 record: %v", err)
		}
		server3, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverID3})
		if err != nil {
			t.Fatalf("failed to find server 3 record: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())

		// Process Day log
		logLine1 := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Town/Town?Name=Player?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=10?Game=CheckpointHardcore?Lighting=Day`
		err = parser.ParseAndProcess(ctx, logLine1, serverID1, "test.log")
		if err != nil {
			t.Fatalf("failed to process day log: %v", err)
		}

		// Process Night log
		logLine2 := `[2025.11.08-15.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Game=Checkpoint?Lighting=Night`
		err = parser.ParseAndProcess(ctx, logLine2, serverID2, "test.log")
		if err != nil {
			t.Fatalf("failed to process night log: %v", err)
		}

		// Process Legacy log (no lighting)
		logLine3 := `[2025.11.08-16.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Refinery/Refinery?Name=Player?Scenario=Scenario_Refinery_Push_Insurgents?MaxPlayers=8?Game=Push`
		err = parser.ParseAndProcess(ctx, logLine3, serverID3, "test.log")
		if err != nil {
			t.Fatalf("failed to process legacy log: %v", err)
		}

		// Find events for each server separately to isolate results
		dayEvents, err := testApp.FindRecordsByFilter("events", "type = 'map_load' && server = {:serverId}", "-created", 1, 0, map[string]any{"serverId": server1.Id})
		if err != nil || len(dayEvents) == 0 {
			t.Fatalf("failed to find day event: %v", err)
		}

		nightEvents, err := testApp.FindRecordsByFilter("events", "type = 'map_load' && server = {:serverId}", "-created", 1, 0, map[string]any{"serverId": server2.Id})
		if err != nil || len(nightEvents) == 0 {
			t.Fatalf("failed to find night event: %v", err)
		}

		legacyEvents, err := testApp.FindRecordsByFilter("events", "type = 'map_load' && server = {:serverId}", "-created", 1, 0, map[string]any{"serverId": server3.Id})
		if err != nil || len(legacyEvents) == 0 {
			t.Fatalf("failed to find legacy event: %v", err)
		}

		// Check day event
		var dayData map[string]interface{}
		json.Unmarshal([]byte(dayEvents[0].GetString("data")), &dayData)
		if lighting, ok := dayData["lighting"].(string); !ok || lighting != "Day" {
			t.Errorf("Expected Day lighting, got %v", dayData["lighting"])
		}

		// Check night event
		var nightData map[string]interface{}
		json.Unmarshal([]byte(nightEvents[0].GetString("data")), &nightData)
		if lighting, ok := nightData["lighting"].(string); !ok || lighting != "Night" {
			t.Errorf("Expected Night lighting, got %v", nightData["lighting"])
		}

		// Check legacy event (no lighting)
		var legacyData map[string]interface{}
		json.Unmarshal([]byte(legacyEvents[0].GetString("data")), &legacyData)
		if lighting := legacyData["lighting"]; lighting != nil {
			t.Errorf("Expected nil lighting for legacy log, got %v", lighting)
		}
	})

	t.Run("MapLoad_Without_Lighting_With_Game", func(t *testing.T) {
		serverExternalID := "test-server-with-game-no-light"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "With Game No Light Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Get server record for filtering
		serverRec, err := testApp.FindFirstRecordByFilter("servers", "external_id = {:id}", map[string]any{"id": serverExternalID})
		if err != nil {
			t.Fatalf("failed to find server record: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		logLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Game=Checkpoint`

		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Verify map_load event was created - filter by server to isolate this test
		events, err := testApp.FindRecordsByFilter("events", "type = 'map_load' && server = {:serverId}", "-created", 1, 0, map[string]any{"serverId": serverRec.Id})
		if err != nil || len(events) == 0 {
			t.Fatalf("failed to find map_load event: %v", err)
		}

		event := events[0]
		eventData := event.GetString("data")

		var mapLoadData map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &mapLoadData); err != nil {
			t.Fatalf("failed to parse event data: %v", err)
		}

		// Lighting should be nil
		if lighting := mapLoadData["lighting"]; lighting != nil {
			t.Errorf("Expected lighting to be nil, got %v", lighting)
		}

		// But game should be present
		if game := mapLoadData["game"]; game != "Checkpoint" {
			t.Errorf("Expected game 'Checkpoint', got %v", game)
		}
	})
}
