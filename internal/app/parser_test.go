package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	db "sandstorm-tracker/internal/db"
	gen "sandstorm-tracker/internal/db/generated"
)

func TestParseAndWriteLogToDB_HCLog(t *testing.T) {
	dbPath := "test_parse_and_write_hclog_v2.sqlite"
	dbService, err := db.NewDatabaseService(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer func() {
		dbService.Close()
		_ = os.Remove(dbPath)
	}()
	queries := dbService.GetQueries()
	ctx := context.Background()

	// Insert a server row so serverID is valid for foreign key constraints
	serverPath := "test/path"
	serverParams := gen.CreateServerParams{
		ExternalID: "test-server-1",
		Name:       "Test Server",
		Path:       &serverPath,
	}
	server, err := queries.CreateServer(ctx, serverParams)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}
	serverID := server.ID

	// Open hc.log
	file, err := os.Open("hc.log")
	if err != nil {
		t.Fatalf("Error opening hc.log: %v", err)
	}
	defer file.Close()

	// Create parser with queries
	parser := NewLogParser(queries)
	scanner := bufio.NewScanner(file)

	// Parse and write all events
	linesProcessed := 0
	errorCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		linesProcessed++

		// Process the line (parse and write to DB)
		if err := parser.ParseAndProcess(ctx, line, serverID, "test.log"); err != nil {
			errorCount++
			t.Logf("Warning: error processing line %d: %v", linesProcessed, err)
		}
	}

	t.Logf("Processed %d lines with %d errors", linesProcessed, errorCount)
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning hc.log: %v", err)
	}

	// Now verify the database contains the expected data by querying
	players, err := queries.ListPlayers(ctx)
	if err != nil {
		t.Fatalf("ListPlayers failed: %v", err)
	}

	// Count kills from database for each tracked player
	rabbitKills := int64(0)
	originKills := int64(0)
	armoredBearKills := int64(0)
	blueKills := int64(0)
	foundPlayers := map[string]bool{}

	for _, player := range players {
		playerName := player.Name
		if strings.Contains(playerName, "Rabbit") {
			foundPlayers["Rabbit"] = true
			rabbitKills = calculateTotalKillsFromMatchStats(ctx, queries, player.ID)
		} else if strings.Contains(playerName, "0rigin") {
			foundPlayers["0rigin"] = true
			originKills = calculateTotalKillsFromMatchStats(ctx, queries, player.ID)
		} else if strings.Contains(playerName, "ArmoredBear") {
			foundPlayers["ArmoredBear"] = true
			armoredBearKills = calculateTotalKillsFromMatchStats(ctx, queries, player.ID)
		} else if strings.Contains(playerName, "Blue") {
			foundPlayers["Blue"] = true
			blueKills = calculateTotalKillsFromMatchStats(ctx, queries, player.ID)
		}
	}

	// Run the same assertions as the original test
	t.Run("Rabbit kills from DB", func(t *testing.T) {
		if rabbitKills != 51 {
			t.Errorf("Expected 51 Rabbit kills from DB, got %d", rabbitKills)
		}
	})
	t.Run("0rigin kills from DB", func(t *testing.T) {
		if originKills != 25 {
			t.Errorf("Expected 25 0rigin kills from DB, got %d", originKills)
		}
	})
	t.Run("ArmoredBear kills from DB", func(t *testing.T) {
		if armoredBearKills != 18 {
			t.Errorf("Expected 18 ArmoredBear kills from DB, got %d", armoredBearKills)
		}
	})
	t.Run("Blue kills from DB", func(t *testing.T) {
		if blueKills != 8 {
			t.Errorf("Expected 8 Blue kills from DB, got %d", blueKills)
		}
	})
	t.Run("total parsed kills from DB", func(t *testing.T) {
		total := rabbitKills + originKills + armoredBearKills + blueKills
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
func calculateTotalKillsFromMatchStats(ctx context.Context, queries *gen.Queries, playerID int64) int64 {
	// Get all match player stats for this player
	matches, err := queries.GetPlayerMatchHistory(ctx, gen.GetPlayerMatchHistoryParams{
		PlayerID: playerID,
		Limit:    1000, // Get all matches
	})
	if err != nil {
		return 0
	}

	var totalKills int64
	for _, match := range matches {
		if match.Kills != nil {
			totalKills += *match.Kills
		}
	}

	return totalKills
}
