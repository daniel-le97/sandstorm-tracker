package parser

import (
	"context"
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
}
