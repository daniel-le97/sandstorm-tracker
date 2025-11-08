package parser

import (
	"bufio"
	"context"
	"os"
	"sandstorm-tracker/internal/database"
	"testing"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

// TestProcessUUIDLog tests processing a complete log file with UUID filename
// and verifies all events are written to the database correctly
func TestProcessUUIDLog(t *testing.T) {
	// Create test PocketBase app with temp directory
	tempDir := t.TempDir()
	testApp, err := tests.NewTestApp(tempDir)
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Verify collections exist
	collections := []string{"servers", "players", "matches", "match_player_stats", "match_weapon_stats"}
	for _, name := range collections {
		if _, err := testApp.FindCollectionByNameOrId(name); err != nil {
			t.Fatalf("Collection %s not found after migration: %v", name, err)
		}
	}
	t.Log("✓ All test collections verified")

	ctx := context.Background()

	// Create a test server using the UUID from the filename
	serverExternalID := "1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde"
	_, err = database.GetOrCreateServer(ctx, testApp, serverExternalID, "Test Server", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Open the UUID log file
	logFile := "test_data/1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde.log"
	file, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("Error opening %s: %v", logFile, err)
	}
	defer file.Close()

	// Create parser with PocketBase app
	parser := NewLogParser(testApp)
	scanner := bufio.NewScanner(file)

	// Parse and write all events
	linesProcessed := 0
	errorCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		linesProcessed++

		// Process the line (parse and write to DB)
		if err := parser.ParseAndProcess(ctx, line, serverExternalID, logFile); err != nil {
			errorCount++
			t.Logf("Warning: error processing line %d: %v", linesProcessed, err)
		}
	}

	t.Logf("Processed %d lines with %d errors", linesProcessed, errorCount)
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning %s: %v", logFile, err)
	}

	// Verify events were processed correctly
	t.Run("MapLoad event created match", func(t *testing.T) {
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"",
			-1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil {
			t.Fatalf("Failed to query matches: %v", err)
		}
		if len(matches) != 1 {
			t.Errorf("Expected 1 match, got %d", len(matches))
		}
		if len(matches) > 0 {
			mapName := matches[0].GetString("map")
			if mapName != "Ministry" {
				t.Errorf("Expected map 'Ministry', got '%s'", mapName)
			}
			scenario := matches[0].GetString("mode")
			if scenario != "Scenario_Ministry_Checkpoint_Security" {
				t.Errorf("Expected scenario 'Scenario_Ministry_Checkpoint_Security', got '%s'", scenario)
			}
		}
	})

	t.Run("Player ArmoredBear created", func(t *testing.T) {
		players, err := testApp.FindRecordsByFilter(
			"players",
			"external_id = {:steamID}",
			"",
			-1,
			0,
			map[string]any{"steamID": "76561198995742987"},
		)
		if err != nil {
			t.Fatalf("Failed to query players: %v", err)
		}
		if len(players) != 1 {
			t.Errorf("Expected 1 player (ArmoredBear), got %d", len(players))
		}
		if len(players) > 0 {
			name := players[0].GetString("name")
			if name != "ArmoredBear" {
				t.Errorf("Expected player name 'ArmoredBear', got '%s'", name)
			}
		}
	})

	t.Run("Kill events recorded", func(t *testing.T) {
		// Find ArmoredBear player
		players, err := testApp.FindRecordsByFilter(
			"players",
			"external_id = {:steamID}",
			"",
			-1,
			0,
			map[string]any{"steamID": "76561198995742987"},
		)
		if err != nil || len(players) == 0 {
			t.Fatalf("Failed to find ArmoredBear player")
		}
		playerID := players[0].Id

		// Find match player stats
		stats, err := testApp.FindRecordsByFilter(
			"match_player_stats",
			"player = {:playerID}",
			"",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query player stats: %v", err)
		}
		if len(stats) != 1 {
			t.Errorf("Expected 1 match_player_stats record, got %d", len(stats))
		}
		if len(stats) > 0 {
			kills := stats[0].GetInt("kills")
			// Based on the extracted log, ArmoredBear has 13 kills
			if kills != 13 {
				t.Errorf("Expected 13 kills for ArmoredBear, got %d", kills)
			}
		}
	})

	t.Run("Objective destroyed event recorded", func(t *testing.T) {
		// Find ArmoredBear player
		players, err := testApp.FindRecordsByFilter(
			"players",
			"external_id = {:steamID}",
			"",
			-1,
			0,
			map[string]any{"steamID": "76561198995742987"},
		)
		if err != nil || len(players) == 0 {
			t.Fatalf("Failed to find ArmoredBear player")
		}
		playerID := players[0].Id

		// Find match player stats
		stats, err := testApp.FindRecordsByFilter(
			"match_player_stats",
			"player = {:playerID}",
			"",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query player stats: %v", err)
		}
		if len(stats) != 1 {
			t.Fatalf("Expected 1 match_player_stats record, got %d", len(stats))
		}

		objectivesDestroyed := stats[0].GetInt("objectives_destroyed")
		// Based on extracted log: 2 objectives destroyed
		if objectivesDestroyed != 2 {
			t.Errorf("Expected 2 objectives destroyed, got %d", objectivesDestroyed)
		}
	})

	t.Run("Objective captured event recorded", func(t *testing.T) {
		// Find ArmoredBear player
		players, err := testApp.FindRecordsByFilter(
			"players",
			"external_id = {:steamID}",
			"",
			-1,
			0,
			map[string]any{"steamID": "76561198995742987"},
		)
		if err != nil || len(players) == 0 {
			t.Fatalf("Failed to find ArmoredBear player")
		}
		playerID := players[0].Id

		// Find match player stats
		stats, err := testApp.FindRecordsByFilter(
			"match_player_stats",
			"player = {:playerID}",
			"",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query player stats: %v", err)
		}
		if len(stats) != 1 {
			t.Fatalf("Expected 1 match_player_stats record, got %d", len(stats))
		}

		objectivesCaptured := stats[0].GetInt("objectives_captured")
		// Based on extracted log: 1 objective captured
		if objectivesCaptured != 1 {
			t.Errorf("Expected 1 objective captured, got %d", objectivesCaptured)
		}
	})

	t.Run("Weapon stats recorded", func(t *testing.T) {
		// Find ArmoredBear player
		players, err := testApp.FindRecordsByFilter(
			"players",
			"external_id = {:steamID}",
			"",
			-1,
			0,
			map[string]any{"steamID": "76561198995742987"},
		)
		if err != nil || len(players) == 0 {
			t.Fatalf("Failed to find ArmoredBear player")
		}
		playerID := players[0].Id

		// Find weapon stats
		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"player = {:playerID}",
			"",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query weapon stats: %v", err)
		}

		// Should have multiple weapon records
		if len(weaponStats) == 0 {
			t.Error("Expected weapon stats records, got none")
		}

		// Verify total kills across all weapons equals match kills
		totalWeaponKills := 0
		for _, stat := range weaponStats {
			kills := stat.GetInt("kills")
			totalWeaponKills += kills
			weaponName := stat.GetString("weapon_name")
			t.Logf("Weapon: %s, Kills: %d", weaponName, kills)
		}

		if totalWeaponKills != 13 {
			t.Errorf("Expected total weapon kills to be 13, got %d", totalWeaponKills)
		}
	})

	t.Run("Round end recorded", func(t *testing.T) {
		// The log shows the round ended, but we don't currently track round end events
		// in match records. This is a placeholder for future round tracking.
		t.Log("Round end event detected in log (not yet tracked in DB)")
	})

	t.Run("Player disconnect recorded", func(t *testing.T) {
		// Disconnect event is parsed but not currently tracked in DB
		// This is a placeholder for future disconnect tracking
		t.Log("Player disconnect event detected in log (not yet tracked in DB)")
	})
}
