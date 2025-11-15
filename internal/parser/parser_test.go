package parser

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sandstorm-tracker/internal/database"
	"strings"
	"testing"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

const testDataDir = "./test_data"

// setupTestCollections verifies that migrations created the expected collections
func setupTestCollections(t *testing.T, testApp *tests.TestApp) {
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
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	setupTestCollections(t, testApp)
	ctx := context.Background()
	serverExternalID := "test-server-1"

	_, err = database.GetOrCreateServer(ctx, testApp, serverExternalID, "Main Server", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	file, err := os.Open("test_data/hc.log")
	if err != nil {
		t.Fatalf("Error opening hc.log: %v", err)
	}
	defer file.Close()

	parser := NewLogParser(testApp, testApp.Logger())
	scanner := bufio.NewScanner(file)
	linesProcessed := 0

	for scanner.Scan() {
		parser.ParseAndProcess(ctx, scanner.Text(), serverExternalID, "test.log")
		linesProcessed++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning hc.log: %v", err)
	}

	t.Logf("Processed %d lines", linesProcessed)

	// Verify kill events were created
	killEvents, err := testApp.FindRecordsByFilter("events", "type = 'player_kill'", "-created", 200, 0)
	if err != nil {
		t.Fatalf("failed to query kill events: %v", err)
	}

	// Verify that kill events were created (hc.log has 102 kills total)
	// We're testing the parser creates events, not the exact distribution
	if len(killEvents) < 100 {
		t.Errorf("Expected at least 100 kill events, got %d", len(killEvents))
	}

	// Verify events have required fields in data JSON
	firstEvent := killEvents[0]
	data := firstEvent.GetString("data")
	requiredFields := []string{"killer_name", "killer_steam_id", "victim_name", "weapon"}
	for _, field := range requiredFields {
		if !strings.Contains(data, fmt.Sprintf(`"%s"`, field)) {
			t.Errorf("Kill event missing required field: %s", field)
		}
	}
}

// TestExtractAllEvents reads log files and extracts all recognized events
// to output files for analysis purposes (filters out RCON commands)
func TestExtractAllEvents(t *testing.T) {
	testCases := []struct {
		name       string
		inputFile  string
		outputFile string
	}{
		{
			name:       "HC Log",
			inputFile:  "test_data/hc.log",
			outputFile: "test_data/hc.extracted.log",
		},
		{
			name:       "Normal Log",
			inputFile:  "test_data/normal.log",
			outputFile: "test_data/normal.extracted.log",
		},
		{
			name:       "UUID Log",
			inputFile:  "test_data/1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde.log",
			outputFile: "test_data/1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde.extracted.log",
		},
		{
			name:       "Full Log",
			inputFile:  "test_data/full.log",
			outputFile: "test_data/full.extracted.log",
		},
		{
			name:       "Full Log 2",
			inputFile:  "test_data/full-2.log",
			outputFile: "test_data/full-2.extracted.log",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Open input log file
			file, err := os.Open(tc.inputFile)
			if err != nil {
				t.Fatalf("Error opening %s: %v", tc.inputFile, err)
			}
			defer file.Close()

			// Create output file for extracted events
			outFile, err := os.Create(tc.outputFile)
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
				} else if parser.patterns.PlayerRegister.MatchString(line) {
					eventType = "REGISTER"
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
				} else if parser.patterns.MapTravel.MatchString(line) {
					eventType = "MAP_TRAVEL"
					matched = true
					// } else if parser.patterns.DifficultyChange.MatchString(line) {
					// 	eventType = "DIFFICULTY"
					// 	matched = true
				} else if parser.patterns.MapVote.MatchString(line) {
					eventType = "MAP_VOTE"
					matched = true
				} else if parser.patterns.ChatCommand.MatchString(line) {
					eventType = "CHAT_CMD"
					matched = true
				} else if parser.patterns.ObjectiveDestroyed.MatchString(line) {
					eventType = "OBJ_DESTROYED"
					matched = true
				} else if parser.patterns.ObjectiveCaptured.MatchString(line) {
					eventType = "OBJ_CAPTURED"
					matched = true
				}

				if matched {
					eventsExtracted++
					eventCounts[eventType]++
					fmt.Fprintf(outFile, "[%s] %s\n", eventType, line)
				}
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("Error scanning %s: %v", tc.inputFile, err)
			}

			// Log statistics
			t.Logf("Total lines processed: %d", totalLines)
			t.Logf("Events extracted: %d", eventsExtracted)
			t.Logf("\nEvent breakdown:")
			for eventType, count := range eventCounts {
				t.Logf("  %s: %d", eventType, count)
			}
			t.Logf("\nExtracted events written to: %s", tc.outputFile)
		})
	}
}
