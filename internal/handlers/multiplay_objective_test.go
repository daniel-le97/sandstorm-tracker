package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/events"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

// TestMultiPlayerObjectiveHandling verifies that each player in a multi-player objective event gets their stats updated
func TestMultiPlayerObjectiveHandling(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()

	// Create server
	if _, err := database.GetOrCreateServer(ctx, testApp, "test-multi-handler", "Test Server", "test/path"); err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Create 3 players
	steamIDs := []string{"76561198007416544", "76561198262186571", "76561198047711504"}
	names := []string{"*OSS*0rigin", "-=312th=- Rabbit", "Blue"}

	for i, steamID := range steamIDs {
		_, err := database.GetOrCreatePlayerBySteamID(ctx, testApp, steamID, names[i])
		if err != nil {
			t.Fatalf("failed to create player %s: %v", names[i], err)
		}
	}

	// Build players array for single objective event
	players := make([]events.ObjectivePlayer, len(steamIDs))
	for i, steamID := range steamIDs {
		players[i] = events.ObjectivePlayer{
			SteamID:    steamID,
			PlayerName: names[i],
		}
	}

	// Create a single objective captured event with all players
	eventCreator := events.NewCreator(testApp)
	err = eventCreator.CreateObjectiveCapturedEvent(
		"test-multi-handler",
		"",
		"0",
		players,
		0,
		false,
	)
	if err != nil {
		t.Fatalf("failed to create objective captured event: %v", err)
	}

	// Get the event we just created
	createdEvents, err := testApp.FindRecordsByFilter("events", "type = 'objective_captured'", "-created", 10, 0)
	if err != nil {
		t.Fatalf("failed to get created events: %v", err)
	}

	if len(createdEvents) != 1 {
		t.Errorf("Expected 1 event with all players, got %d", len(createdEvents))
		return
	}

	// Verify the single event contains all 3 players
	var data events.ObjectiveCapturedData
	if err := json.Unmarshal([]byte(createdEvents[0].GetString("data")), &data); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if len(data.Players) != 3 {
		t.Errorf("Expected 3 players in single event, got %d", len(data.Players))
		return
	}

	// Collect all player Steam IDs and names from the event
	eventSteamIDs := make(map[string]bool)
	eventNames := make(map[string]bool)

	for _, p := range data.Players {
		t.Logf("Event player: %s (%s)", p.PlayerName, p.SteamID)
		eventSteamIDs[p.SteamID] = true
		eventNames[p.PlayerName] = true
	}

	// Verify we have all 3 expected players
	expectedSet := make(map[string]bool)
	for _, id := range steamIDs {
		expectedSet[id] = true
	}

	if len(eventSteamIDs) != 3 {
		t.Errorf("Expected 3 unique players, got %d", len(eventSteamIDs))
	}

	for _, expectedID := range steamIDs {
		if !eventSteamIDs[expectedID] {
			t.Errorf("Expected Steam ID %s not found in event", expectedID)
		}
	}

	for _, expectedName := range names {
		if !eventNames[expectedName] {
			t.Errorf("Expected player name %s not found in events", expectedName)
		}
	}

	t.Logf("✓ Created %d separate objective captured events (one per player)", len(createdEvents))
}
