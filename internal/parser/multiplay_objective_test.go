package parser

import (
	"context"
	"encoding/json"
	"testing"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/events"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

// TestMultiPlayerObjectiveEvents tests that multi-player objectives create separate events for each player
func TestMultiPlayerObjectiveEvents(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()
	serverExternalID := "test-server-multi-obj"

	_, err = database.GetOrCreateServer(ctx, testApp, serverExternalID, "Test Server Multi Obj", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Parse two objective captured lines with multiple players
	testLines := []string{
		// 2 players capturing
		"[OBJ_CAPTURED] [2025.10.04-21.31.01:003][404]LogGameplayEvents: Display: Objective 0 was captured for team 0 from team 1 by *OSS*0rigin[76561198007416544], -=312th=- Rabbit[76561198262186571].",
		// 3 players capturing
		"[OBJ_CAPTURED] [2025.10.04-21.35.09:295][ 74]LogGameplayEvents: Display: Objective 2 was captured for team 0 from team 1 by -=312th=- Rabbit[76561198262186571], Blue[76561198047711504], *OSS*0rigin[76561198007416544].",
		// 2 players destroying
		"[OBJ_DESTROYED] [2025.10.04-21.32.55:614][297]LogGameplayEvents: Display: Objective 1 owned by team 1 was destroyed for team 0 by -=312th=- Rabbit[76561198262186571], ArmoredBear[76561198995742987].",
	}

	parser := NewLogParser(testApp, testApp.Logger())

	for _, line := range testLines {
		parser.ParseAndProcess(ctx, line, serverExternalID, "test.log")
	}

	// Should have 3 events total: 1 from first capture (2 players) + 1 from second capture (3 players) + 1 destroy (2 players)
	allEvents, err := testApp.FindRecordsByFilter("events", "", "-created", 50, 0)
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	t.Logf("Total events created: %d", len(allEvents))
	if len(allEvents) < 3 {
		t.Errorf("Expected at least 3 events (1 captured + 1 captured + 1 destroyed), got %d", len(allEvents))
	}

	// Check captured events
	capturedEvents, err := testApp.FindRecordsByFilter("events", "type = 'objective_captured'", "-created", 50, 0)
	if err != nil {
		t.Fatalf("Failed to query captured events: %v", err)
	}

	t.Logf("Captured events created: %d", len(capturedEvents))
	if len(capturedEvents) != 2 {
		t.Errorf("Expected 2 captured events (one with 2 players, one with 3 players), got %d", len(capturedEvents))
	}

	// Verify each event has correct player data
	playerSteamIDs := make(map[string]bool)
	totalCapturedPlayers := 0
	for i, evt := range capturedEvents {
		dataStr := evt.GetString("data")
		var data events.ObjectiveCapturedData
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			t.Errorf("Failed to unmarshal event data: %v", err)
			continue
		}
		t.Logf("Captured event %d: %d players, objective=%s, team=%d",
			i+1, len(data.Players), data.Objective, data.CapturingTeam)
		for _, p := range data.Players {
			playerSteamIDs[p.SteamID] = true
			totalCapturedPlayers++
			t.Logf("  - player=%s (%s)", p.PlayerName, p.SteamID)
		}
	}

	// Should have seen 3 unique players across captures
	if len(playerSteamIDs) != 3 {
		t.Errorf("Expected 3 unique players in captured events, got %d: %v", len(playerSteamIDs), playerSteamIDs)
	}
	if totalCapturedPlayers != 5 {
		t.Errorf("Expected 5 total captured players (2+3), got %d", totalCapturedPlayers)
	}

	// Check destroyed events
	destroyedEvents, err := testApp.FindRecordsByFilter("events", "type = 'objective_destroyed'", "-created", 50, 0)
	if err != nil {
		t.Fatalf("Failed to query destroyed events: %v", err)
	}

	t.Logf("Destroyed events created: %d", len(destroyedEvents))
	if len(destroyedEvents) != 1 {
		t.Errorf("Expected 1 destroyed event (with 2 players), got %d", len(destroyedEvents))
	}

	for i, evt := range destroyedEvents {
		dataStr := evt.GetString("data")
		var data events.ObjectiveDestroyedData
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			t.Errorf("Failed to unmarshal event data: %v", err)
			continue
		}
		t.Logf("Destroyed event %d: %d players, objective=%s, team=%d",
			i+1, len(data.Players), data.Objective, data.DestroyingTeam)
		for _, p := range data.Players {
			t.Logf("  - player=%s (%s)", p.PlayerName, p.SteamID)
		}
	}
}
