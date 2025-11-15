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

		// Verify match was created with correct data (map loads create matches, not events)
		matches, err := testApp.FindRecordsByFilter("matches", "", "-created", 1, 0)
		if err != nil || len(matches) == 0 {
			t.Fatalf("failed to find match: %v", err)
		}

		match := matches[0]
		if match.GetString("map") != "Town" {
			t.Errorf("Expected map 'Town', got %s", match.GetString("map"))
		}
		if match.GetString("mode") != "Scenario_Hideout_Checkpoint_Security" {
			t.Errorf("Expected mode 'Scenario_Hideout_Checkpoint_Security', got %s", match.GetString("mode"))
		}
		if match.GetString("player_team") != "Security" {
			t.Errorf("Expected player_team 'Security', got %s", match.GetString("player_team"))
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

		matches, err := testApp.FindRecordsByFilter("matches", "", "-created", 1, 0)
		if err != nil || len(matches) == 0 {
			t.Fatalf("failed to find match")
		}

		match := matches[0]
		if match.GetString("map") != "Ministry" {
			t.Errorf("Expected map 'Ministry', got %s", match.GetString("map"))
		}
		if match.GetString("player_team") != "Security" {
			t.Errorf("Expected player_team 'Security', got %s", match.GetString("player_team"))
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

		matches, err := testApp.FindRecordsByFilter("matches", "", "-created", 1, 0)
		if err != nil || len(matches) == 0 {
			t.Fatalf("failed to find match")
		}

		match := matches[0]
		if match.GetString("map") != "Ministry" {
			t.Errorf("Expected map 'Ministry', got %s", match.GetString("map"))
		}
		if match.GetString("player_team") != "Insurgents" {
			t.Errorf("Expected player_team 'Insurgents', got %s", match.GetString("player_team"))
		}
	})
}
