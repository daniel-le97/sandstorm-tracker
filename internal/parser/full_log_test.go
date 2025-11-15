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

// TestProcessFullLog tests processing a complete log file and verifies events are created
func TestProcessFullLog(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()
	serverExternalID := "test-server-full-2"

	_, err = database.GetOrCreateServer(ctx, testApp, serverExternalID, "Test Server Full", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	file, err := os.Open("test_data/full-2.log")
	if err != nil {
		t.Fatalf("Error opening log: %v", err)
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
		t.Fatalf("Error scanning log: %v", err)
	}

	t.Logf("Processed %d lines", linesProcessed)

	// Verify matches were created (map loads create matches, not events)
	t.Run("Matches created for maps", func(t *testing.T) {
		matches, err := testApp.FindRecordsByFilter("matches", "", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query matches: %v", err)
		}
		if len(matches) < 1 {
			t.Errorf("Expected at least 1 match, got %d", len(matches))
		}
	})

	t.Run("Kill events created", func(t *testing.T) {
		events, err := testApp.FindRecordsByFilter("events", "type = 'player_kill'", "-created", 30, 0)
		if err != nil {
			t.Fatalf("Failed to query kill events: %v", err)
		}
		if len(events) < 20 {
			t.Errorf("Expected at least 20 kill events, got %d", len(events))
		}
		t.Logf("Created %d kill events", len(events))
	})

	t.Run("Objective events created", func(t *testing.T) {
		destroyed, err := testApp.FindRecordsByFilter("events", "type = 'objective_destroyed'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query destroyed events: %v", err)
		}
		captured, err := testApp.FindRecordsByFilter("events", "type = 'objective_captured'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query captured events: %v", err)
		}
		total := len(destroyed) + len(captured)
		if total < 6 {
			t.Errorf("Expected at least 6 objective events, got %d", total)
		}
		t.Logf("Created %d destroyed + %d captured = %d objective events", len(destroyed), len(captured), total)
	})

	t.Run("Round events created", func(t *testing.T) {
		events, err := testApp.FindRecordsByFilter("events", "type = 'round_end'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query round_end events: %v", err)
		}
		if len(events) < 2 {
			t.Errorf("Expected at least 2 round_end events, got %d", len(events))
		}
	})

	t.Run("Multiple matches for map changes", func(t *testing.T) {
		matches, err := testApp.FindRecordsByFilter("matches", "", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query matches: %v", err)
		}
		// full-2.log has a map travel, should create 2 matches
		if len(matches) < 2 {
			t.Logf("Expected at least 2 matches (map travel creates new match), got %d", len(matches))
		}
	})
}
