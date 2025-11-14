package parser

import (
	"context"
	"sandstorm-tracker/internal/database"
	"testing"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

func TestMapLoadEvents(t *testing.T) {
	// Create test PocketBase app with temp directory
	tempDir := t.TempDir()
	testApp, err := tests.NewTestApp(tempDir)
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create collections
	setupTestCollections(t, testApp)

	ctx := context.Background()

	// Test case 1: Hideout Hardcore Checkpoint Security
	t.Run("Hideout_Hardcore_Checkpoint_Security", func(t *testing.T) {
		serverExternalID := "test-server-hideout"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Hideout Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp, testApp.Logger())

		// Map load log line
		logLine := `[2025.10.04-21.18.15:445][  0]LogLoad: LoadMap: /Game/Maps/Town/Town?Name=Player?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=10?Game=CheckpointHardcore?Lighting=Day`

		// Process the line
		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Query the match from database
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"",
			1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil || len(matches) == 0 {
			t.Fatalf("failed to find match: %v", err)
		}

		match := matches[0]

		// Verify map name
		mapName := match.GetString("map")
		if mapName != "Town" {
			t.Errorf("Expected map 'Town', got '%s'", mapName)
		}

		// Verify mode (scenario)
		mode := match.GetString("mode")
		if mode != "Scenario_Hideout_Checkpoint_Security" {
			t.Errorf("Expected mode 'Scenario_Hideout_Checkpoint_Security', got '%s'", mode)
		}

		// Verify player team
		playerTeam := match.GetString("player_team")
		if playerTeam != "Security" {
			t.Errorf("Expected player_team 'Security', got '%s'", playerTeam)
		}
	})

	// Test case 2: Ministry Checkpoint Security
	t.Run("Ministry_Checkpoint_Security", func(t *testing.T) {
		serverExternalID := "test-server-ministry"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Ministry Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp, testApp.Logger())

		// Map load log line
		logLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`

		// Process the line
		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Query the match from database
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"",
			1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil || len(matches) == 0 {
			t.Fatalf("failed to find match: %v", err)
		}

		match := matches[0]

		// Verify map name
		mapName := match.GetString("map")
		if mapName != "Ministry" {
			t.Errorf("Expected map 'Ministry', got '%s'", mapName)
		}

		// Verify mode (scenario)
		mode := match.GetString("mode")
		if mode != "Scenario_Ministry_Checkpoint_Security" {
			t.Errorf("Expected mode 'Scenario_Ministry_Checkpoint_Security', got '%s'", mode)
		}

		// Verify player team
		playerTeam := match.GetString("player_team")
		if playerTeam != "Security" {
			t.Errorf("Expected player_team 'Security', got '%s'", playerTeam)
		}
	})

	// Test case 3: Insurgents team
	t.Run("Insurgents_Team", func(t *testing.T) {
		serverExternalID := "test-server-insurgents"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Insurgents Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp, testApp.Logger())

		// Map load log line with Insurgents scenario
		logLine := `[2025.11.08-14.00.00:000][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Insurgents?MaxPlayers=8?Lighting=Day`

		// Process the line
		err = parser.ParseAndProcess(ctx, logLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process log line: %v", err)
		}

		// Query the match from database
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"",
			1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil || len(matches) == 0 {
			t.Fatalf("failed to find match: %v", err)
		}

		match := matches[0]

		// Verify player team
		playerTeam := match.GetString("player_team")
		if playerTeam != "Insurgents" {
			t.Errorf("Expected player_team 'Insurgents', got '%s'", playerTeam)
		}
	})
}
