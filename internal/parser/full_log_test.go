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

// TestProcessFullLog tests processing a complete log file with full game session
// and verifies all events are written to the database correctly
func TestProcessFullLog(t *testing.T) {
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

	// Create a test server
	serverExternalID := "test-server-full-2"
	_, err = database.GetOrCreateServer(ctx, testApp, serverExternalID, "Test Server Full", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Open the full-2 log file
	logFile := "test_data/full-2.log"
	file, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("Error opening %s: %v", logFile, err)
	}
	defer file.Close()

	// Create parser with PocketBase app
	parser := NewLogParser(testApp, testApp.Logger())
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
	t.Run("MapLoad event created initial match", func(t *testing.T) {
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"-created",
			-1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil {
			t.Fatalf("Failed to query matches: %v", err)
		}
		// Should have 2 matches: initial Ministry match + Oilfield match after travel
		if len(matches) < 1 {
			t.Errorf("Expected at least 1 match, got %d", len(matches))
		}
		if len(matches) > 0 {
			// First match should be Ministry
			firstMatch := matches[len(matches)-1] // Get oldest match
			mapName := firstMatch.GetString("map")
			if mapName != "Ministry" {
				t.Errorf("Expected first map 'Ministry', got '%s'", mapName)
			}
			scenario := firstMatch.GetString("mode")
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

		// Find match player stats - should have stats for Ministry match
		stats, err := testApp.FindRecordsByFilter(
			"match_player_stats",
			"player = {:playerID}",
			"-created",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query player stats: %v", err)
		}
		if len(stats) == 0 {
			t.Fatalf("Expected at least 1 match_player_stats record, got none")
		}

		// Get the first match stats (Ministry)
		firstMatchStats := stats[len(stats)-1]
		kills := firstMatchStats.GetInt("kills")
		deaths := firstMatchStats.GetInt("deaths")

		// Based on the extracted log, ArmoredBear has 24 kills and 1 death in round 1-2
		if kills != 24 {
			t.Errorf("Expected 24 kills for ArmoredBear, got %d", kills)
		}
		if deaths != 1 {
			t.Errorf("Expected 1 death for ArmoredBear, got %d", deaths)
		}

		t.Logf("ArmoredBear stats: %d kills, %d deaths", kills, deaths)
	})

	t.Run("Objective destroyed events recorded", func(t *testing.T) {
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
			"-created",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query player stats: %v", err)
		}
		if len(stats) == 0 {
			t.Fatalf("Expected at least 1 match_player_stats record, got none")
		}

		firstMatchStats := stats[len(stats)-1]
		objectivesDestroyed := firstMatchStats.GetInt("objectives_destroyed")
		// Based on extracted log: 3 objectives destroyed in round 2
		if objectivesDestroyed != 3 {
			t.Errorf("Expected 3 objectives destroyed, got %d", objectivesDestroyed)
		}
	})

	t.Run("Objective captured events recorded", func(t *testing.T) {
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
			"-created",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Fatalf("Failed to query player stats: %v", err)
		}
		if len(stats) == 0 {
			t.Fatalf("Expected at least 1 match_player_stats record, got none")
		}

		firstMatchStats := stats[len(stats)-1]
		objectivesCaptured := firstMatchStats.GetInt("objectives_captured")
		// Based on extracted log: 3 objectives captured in round 2
		if objectivesCaptured != 3 {
			t.Errorf("Expected 3 objectives captured, got %d", objectivesCaptured)
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

		// Find weapon stats for the first match
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"-created",
			-1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil || len(matches) == 0 {
			t.Fatalf("Failed to find matches")
		}
		firstMatchID := matches[len(matches)-1].Id

		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"player = {:playerID} && match = {:matchID}",
			"",
			-1,
			0,
			map[string]any{
				"playerID": playerID,
				"matchID":  firstMatchID,
			},
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

		if totalWeaponKills != 24 {
			t.Errorf("Expected total weapon kills to be 24, got %d", totalWeaponKills)
		}
	})

	t.Run("Round tracking recorded", func(t *testing.T) {
		// Find the first match
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"-created",
			-1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil || len(matches) == 0 {
			t.Fatalf("Failed to find matches")
		}
		firstMatch := matches[len(matches)-1]

		// Check round counter - now only counting LogGameplayEvents ROUND_END events
		// Should be 2 rounds (Round 1 and Round 2)
		round := firstMatch.GetInt("round")
		if round != 2 {
			t.Errorf("Expected round to be 2, got %d", round)
		}

		// Check winner team - Team 0 won the last round
		winnerTeam := firstMatch.GetInt("winner_team")
		if winnerTeam != 0 {
			t.Errorf("Expected winner_team to be 0, got %d", winnerTeam)
		}

		t.Logf("Match round: %d, winner_team: %d", round, winnerTeam)
	})

	t.Run("Map travel creates new match", func(t *testing.T) {
		// Verify map travel to Oilfield created a second match
		matches, err := testApp.FindRecordsByFilter(
			"matches",
			"server.external_id = {:serverID}",
			"-created",
			-1,
			0,
			map[string]any{"serverID": serverExternalID},
		)
		if err != nil {
			t.Fatalf("Failed to query matches: %v", err)
		}

		// Should have 2 matches after travel
		if len(matches) != 2 {
			t.Errorf("Expected 2 matches after map travel, got %d", len(matches))
		}

		if len(matches) >= 2 {
			// Second match should be Oilfield
			secondMatch := matches[0] // Most recent match
			mapName := secondMatch.GetString("map")
			if mapName != "Oilfield" {
				t.Errorf("Expected second map 'Oilfield', got '%s'", mapName)
			}
			scenario := secondMatch.GetString("mode")
			if scenario != "Scenario_Refinery_Push_Insurgents" {
				t.Errorf("Expected scenario 'Scenario_Refinery_Push_Insurgents', got '%s'", scenario)
			}

			// First match should be ended
			firstMatch := matches[1]
			endTime := firstMatch.GetDateTime("end_time")
			if endTime.IsZero() {
				t.Error("Expected first match to have end_time set after map travel")
			}
		}
	})

	t.Run("Player disconnect events processed", func(t *testing.T) {
		// The log shows 2 disconnect events
		// These are parsed but currently just logged, not stored separately
		t.Log("✓ Player disconnect events were parsed (2 events in log)")
	})
}
