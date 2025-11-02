package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/tests"
)

const testDataDir = "./test_pb_data"

// setupTestCollections runs migrations to create collections
func setupTestCollections(t *testing.T, testApp *tests.TestApp) {
	// Register jsvm plugin to enable JavaScript migrations
	migrationsDir := filepath.Join(testDataDir, "pb_migrations")
	jsvm.MustRegister(testApp, jsvm.Config{
		MigrationsDir: migrationsDir,
	})

	// Run migrations to create the collections
	if err := testApp.RunAllMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Log("✓ Migrations completed")

	// Verify collections exist
	collections := []string{"servers", "players", "matches", "match_player_stats", "match_weapon_stats"}
	for _, name := range collections {
		if _, err := testApp.FindCollectionByNameOrId(name); err != nil {
			t.Fatalf("Collection %s not found after migration: %v", name, err)
		}
	}
	t.Log("✓ All test collections verified")
}

func TestParseAndWriteLogToDB_HCLog(t *testing.T) {
	// Create test PocketBase app
	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create collections
	setupTestCollections(t, testApp)

	ctx := context.Background()

	// Create a test server using PocketBase Record API
	serverExternalID := "test-server-1"
	_, err = GetOrCreateServer(ctx, testApp, serverExternalID, "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Open hc.log
	file, err := os.Open("hc.log")
	if err != nil {
		t.Fatalf("Error opening hc.log: %v", err)
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
		// Note: ParseAndProcess expects external_id as serverID parameter
		if err := parser.ParseAndProcess(ctx, line, serverExternalID, "test.log"); err != nil {
			errorCount++
			t.Logf("Warning: error processing line %d: %v", linesProcessed, err)
		}
	}

	t.Logf("Processed %d lines with %d errors", linesProcessed, errorCount)
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning hc.log: %v", err)
	}

	// Verify kill counts were recorded correctly
	// Expected kills (only counting first killer in multi-killer events):
	// These are the correct values when only the first killer gets credit
	expectedKills := map[string]int{
		"Rabbit":      51,
		"0rigin":      25,
		"ArmoredBear": 18,
		"Blue":        8,
	}

	actualKills := make(map[string]int)
	foundPlayers := make(map[string]bool)

	// Query each player's total kills from match_weapon_stats
	for playerName := range expectedKills {
		// Find player by name (contains the name)
		players, err := testApp.FindRecordsByFilter(
			"players",
			"name ~ {:name}",
			"-created",
			100,
			0,
			map[string]any{"name": playerName},
		)
		if err != nil {
			t.Logf("Warning: Error finding player %s: %v", playerName, err)
			continue
		}

		if len(players) == 0 {
			t.Logf("Warning: Player %s not found in database", playerName)
			continue
		}

		// Mark player as found
		foundPlayers[playerName] = true

		// Get the player record
		player := players[0]
		playerID := player.Id

		// Sum kills from match_weapon_stats
		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"player = {:playerID}",
			"",
			-1,
			0,
			map[string]any{"playerID": playerID},
		)
		if err != nil {
			t.Logf("Warning: Error getting weapon stats for %s: %v", playerName, err)
			continue
		}

		totalKills := 0
		for _, stat := range weaponStats {
			kills := stat.GetInt("kills")
			totalKills += kills
		}

		actualKills[playerName] = totalKills
	}

	// Run individual tests for each player
	t.Run("Rabbit kills from DB", func(t *testing.T) {
		if actualKills["Rabbit"] != 51 {
			t.Errorf("Expected 51 Rabbit kills from DB, got %d", actualKills["Rabbit"])
		}
	})

	t.Run("0rigin kills from DB", func(t *testing.T) {
		if actualKills["0rigin"] != 25 {
			t.Errorf("Expected 25 0rigin kills from DB, got %d", actualKills["0rigin"])
		}
	})

	t.Run("ArmoredBear kills from DB", func(t *testing.T) {
		if actualKills["ArmoredBear"] != 18 {
			t.Errorf("Expected 18 ArmoredBear kills from DB, got %d", actualKills["ArmoredBear"])
		}
	})

	t.Run("Blue kills from DB", func(t *testing.T) {
		if actualKills["Blue"] != 8 {
			t.Errorf("Expected 8 Blue kills from DB, got %d", actualKills["Blue"])
		}
	})

	t.Run("total parsed kills from DB", func(t *testing.T) {
		total := actualKills["Rabbit"] + actualKills["0rigin"] + actualKills["ArmoredBear"] + actualKills["Blue"]
		if total != 102 {
			t.Errorf("Expected 102 total parsed kills from DB (excluding assists), got %d", total)
		}
	})

	t.Run("all expected players found", func(t *testing.T) {
		expectedPlayers := []string{"Rabbit", "0rigin", "ArmoredBear", "Blue"}
		for _, name := range expectedPlayers {
			if !foundPlayers[name] {
				t.Errorf("Expected player %s in DB, but not found", name)
			}
		}
	})
}

func TestExtractAllEventsFromHCLog(t *testing.T) {
	// Open hc.log
	file, err := os.Open("hc.log")
	if err != nil {
		t.Fatalf("Error opening hc.log: %v", err)
	}
	defer file.Close()

	// Create output file for extracted events
	outFile, err := os.Create("extracted_events.log")
	if err != nil {
		t.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	// Create parser (without queries since we're just extracting, not writing to DB)
	parser := &LogParser{
		patterns: newLogPatterns(),
	}

	scanner := bufio.NewScanner(file)

	eventCounts := make(map[string]int)
	totalLines := 0
	eventsExtracted := 0

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++

		// Check each event type and write to file if matched
		eventType := ""
		matched := false

		// Check kill events
		if parser.patterns.PlayerKill.MatchString(line) {
			eventType = "KILL"
			matched = true
		} else if parser.patterns.PlayerJoin.MatchString(line) {
			eventType = "JOIN"
			matched = true
		} else if parser.patterns.PlayerDisconnect.MatchString(line) {
			eventType = "DISCONNECT"
			matched = true
		} else if parser.patterns.RoundStart.MatchString(line) {
			eventType = "ROUND_START"
			matched = true
		} else if parser.patterns.RoundEnd.MatchString(line) {
			eventType = "ROUND_END"
			matched = true
		} else if parser.patterns.GameOver.MatchString(line) {
			eventType = "GAME_OVER"
			matched = true
		} else if parser.patterns.MapLoad.MatchString(line) {
			eventType = "MAP_LOAD"
			matched = true
		} else if parser.patterns.DifficultyChange.MatchString(line) {
			eventType = "DIFFICULTY"
			matched = true
		} else if parser.patterns.MapVote.MatchString(line) {
			eventType = "MAP_VOTE"
			matched = true
		} else if parser.patterns.ChatCommand.MatchString(line) {
			eventType = "CHAT_CMD"
			matched = true
		}
		// Skip RCON commands - too many for analysis
		// } else if parser.patterns.RconCommand.MatchString(line) {
		// 	eventType = "RCON_CMD"
		// 	matched = true
		// }

		if matched {
			eventCounts[eventType]++
			eventsExtracted++
			// Write event type and raw line to output file
			_, err := outFile.WriteString("[" + eventType + "] " + line + "\n")
			if err != nil {
				t.Fatalf("Error writing to output file: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning hc.log: %v", err)
	}

	// Log statistics
	t.Logf("Total lines processed: %d", totalLines)
	t.Logf("Events extracted: %d", eventsExtracted)
	t.Logf("\nEvent breakdown:")
	for eventType, count := range eventCounts {
		t.Logf("  %s: %d", eventType, count)
	}
	t.Logf("\nExtracted events written to: extracted_events.log")
}

// TestExtractAllEventsFromNormalLog reads normal.log and extracts all recognized events
// to extracted_events_normal.log for analysis purposes (filters out RCON commands)
func TestExtractAllEventsFromNormalLog(t *testing.T) {
	// Open normal.log
	file, err := os.Open("normal.log")
	if err != nil {
		t.Fatalf("Error opening normal.log: %v", err)
	}
	defer file.Close()

	// Create output file for extracted events
	outFile, err := os.Create("extracted_events_normal.log")
	if err != nil {
		t.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	// Create parser (without queries since we're just extracting, not writing to DB)
	parser := &LogParser{
		patterns: newLogPatterns(),
	}

	scanner := bufio.NewScanner(file)

	eventCounts := make(map[string]int)
	totalLines := 0
	eventsExtracted := 0

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++

		// Check each event type and write to file if matched
		eventType := ""
		matched := false

		if parser.patterns.PlayerKill.MatchString(line) {
			eventType = "KILL"
			matched = true
		} else if parser.patterns.PlayerJoin.MatchString(line) {
			eventType = "JOIN"
			matched = true
		} else if parser.patterns.PlayerDisconnect.MatchString(line) {
			eventType = "DISCONNECT"
			matched = true
		} else if parser.patterns.RoundStart.MatchString(line) {
			eventType = "ROUND_START"
			matched = true
		} else if parser.patterns.RoundEnd.MatchString(line) {
			eventType = "ROUND_END"
			matched = true
		} else if parser.patterns.GameOver.MatchString(line) {
			eventType = "GAME_OVER"
			matched = true
		} else if parser.patterns.MapLoad.MatchString(line) {
			eventType = "MAP_LOAD"
			matched = true
		} else if parser.patterns.DifficultyChange.MatchString(line) {
			eventType = "DIFFICULTY"
			matched = true
		} else if parser.patterns.MapVote.MatchString(line) {
			eventType = "MAP_VOTE"
			matched = true
		} else if parser.patterns.ChatCommand.MatchString(line) {
			eventType = "CHAT_CMD"
			matched = true
		}
		// Skip RCON commands - too many for analysis
		// } else if parser.patterns.RconCommand.MatchString(line) {
		// 	eventType = "RCON_CMD"
		// 	matched = true
		// }

		if matched {
			eventsExtracted++
			eventCounts[eventType]++
			fmt.Fprintf(outFile, "[%s] %s\n", eventType, line)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Log statistics
	t.Logf("Total lines processed: %d", totalLines)
	t.Logf("Events extracted: %d", eventsExtracted)
	t.Logf("\nEvent breakdown:")
	for eventType, count := range eventCounts {
		t.Logf("  %s: %d", eventType, count)
	}
	t.Logf("\nExtracted events written to: extracted_events_normal.log")
}

// Helper function to calculate total kills from match stats
func calculateTotalKillsFromMatchStats(ctx context.Context, app core.App, playerID string) int64 {
	// Get all match player stats for this player
	matchStats, err := app.FindRecordsByFilter(
		"match_player_stats",
		"player = {:playerID}",
		"-created",
		-1,
		0,
		map[string]any{"playerID": playerID},
	)
	if err != nil {
		return 0
	}

	var totalKills int64
	for _, stat := range matchStats {
		kills := stat.GetInt("kills")
		totalKills += int64(kills)
	}

	return totalKills
}
