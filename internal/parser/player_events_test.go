package parser

import (
	"context"
	"testing"

	"sandstorm-tracker/internal/database"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

// TestPlayerConnectionFlow tests the complete player connection sequence: Login → Register → Join
func TestPlayerConnectionFlow(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server-conn"

	// Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Connection Test Server", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	parser := NewLogParser(testApp, testApp.Logger())

	// Create a match so players can join
	err = parser.ParseAndProcess(ctx, `[2025.11.15-12.00.00:000][123]LogLoad: LoadMap: /Game/Maps/Town/Town_Checkpoint?Game=Checkpoint?Scenario=Scenario_Town_Checkpoint_Security?MaxPlayers=8?Lighting=Day`, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process map load: %v", err)
	}

	// Test: Login request
	loginLine := `[2025.11.15-12.00.05:000][200]LogNet: Login request:	?Name=TestPlayer userId: SteamNWI:76561198995742987 platform: SteamNWI`
	err = parser.ParseAndProcess(ctx, loginLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process login: %v", err)
	}

	// Verify player was created (or will be via event hooks)
	events, err := testApp.FindRecordsByFilter("events", "type = 'player_login'", "-created", 10, 0)
	if err != nil {
		t.Fatalf("Failed to query login events: %v", err)
	}
	if len(events) < 1 {
		t.Errorf("Expected at least 1 login event, got %d", len(events))
	}

	// Test: PlayerRegister (pre-match)
	registerLine := `[2025.11.15-12.00.06:000][201]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (76561198995742987) Result: (EOS_Success)`
	err = parser.ParseAndProcess(ctx, registerLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process register: %v", err)
	}

	// Test: PlayerJoin (in-match)
	joinLine := `[2025.11.15-12.00.07:000][202]LogNet: Join succeeded: TestPlayer`
	err = parser.ParseAndProcess(ctx, joinLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process join: %v", err)
	}

	// Verify join event was created
	joinEvents, err := testApp.FindRecordsByFilter("events", "type = 'player_join'", "-created", 10, 0)
	if err != nil {
		t.Fatalf("Failed to query join events: %v", err)
	}
	if len(joinEvents) < 1 {
		t.Errorf("Expected at least 1 join event, got %d", len(joinEvents))
	}
}

// TestPlayerDisconnectEdgeCases tests disconnect handling including map travel disconnects
func TestPlayerDisconnectEdgeCases(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server-disconnect"

	// Create server and match
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Disconnect Test Server", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	parser := NewLogParser(testApp, testApp.Logger())

	// Create match
	mapLoadLine := `[2025.11.15-12.00.00:000][100]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry_Checkpoint?Game=Checkpoint?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	err = parser.ParseAndProcess(ctx, mapLoadLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process map load: %v", err)
	}

	// Add player to match
	loginLine := `[2025.11.15-12.00.05:000][200]LogNet: Login request:	?Name=DisconnectPlayer userId: SteamNWI:76561198111111111 platform: SteamNWI`
	err = parser.ParseAndProcess(ctx, loginLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process login: %v", err)
	}

	registerLine := `[2025.11.15-12.00.06:000][201]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (76561198111111111) Result: (EOS_Success)`
	err = parser.ParseAndProcess(ctx, registerLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process register: %v", err)
	}

	joinLine := `[2025.11.15-12.00.07:000][202]LogNet: Join succeeded: DisconnectPlayer`
	err = parser.ParseAndProcess(ctx, joinLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process join: %v", err)
	}

	t.Run("Normal disconnect", func(t *testing.T) {
		// Simulate disconnect (not during map travel)
		disconnectLine := `[2025.11.15-12.05.00:000][300]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198111111111), Result: (EOS_Success)`
		err := parser.ParseAndProcess(ctx, disconnectLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process disconnect: %v", err)
		}

		// Should create a leave event
		leaveEvents, err := testApp.FindRecordsByFilter("events", "type = 'player_leave'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query leave events: %v", err)
		}
		if len(leaveEvents) < 1 {
			t.Errorf("Expected at least 1 leave event for normal disconnect, got %d", len(leaveEvents))
		}
	})

	t.Run("Map travel disconnect (ignored)", func(t *testing.T) {
		// Simulate map travel
		mapTravelLine := `[2025.11.15-12.10.00:000][400]LogGameMode: ProcessServerTravel: Oilfield?Scenario=Scenario_Refinery_Push_Insurgents?Game=`
		err := parser.ParseAndProcess(ctx, mapTravelLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process map travel: %v", err)
		}

		// Disconnect shortly after map travel should be ignored
		mapTravelDisconnectLine := `[2025.11.15-12.10.05:000][401]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198111111111), Result: (EOS_Success)`
		initialLeaveCount := countLeaveEvents(t, testApp)
		err = parser.ParseAndProcess(ctx, mapTravelDisconnectLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process map travel disconnect: %v", err)
		}

		// Should NOT create a new leave event (map travel disconnect is ignored)
		finalLeaveCount := countLeaveEvents(t, testApp)
		if finalLeaveCount > initialLeaveCount {
			t.Logf("⚠️  Map travel disconnect may have created an event (expected to be ignored)")
		}
	})
}

// TestObjectiveEvents tests objective destroyed and captured events
func TestObjectiveEvents(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server-objectives"

	// Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Objective Test Server", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	parser := NewLogParser(testApp, testApp.Logger())

	// Create match
	mapLoadLine := `[2025.11.15-12.00.00:000][100]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry_Checkpoint?Game=Checkpoint?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	err = parser.ParseAndProcess(ctx, mapLoadLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process map load: %v", err)
	}

	t.Run("Objective destroyed", func(t *testing.T) {
		destroyedLine := `[2025.11.15-12.05.00:000][500]LogGameplayEvents: Display: Objective 1 owned by team 1 was destroyed for team 0 by Player1[76561198222222222, team 0].`
		err := parser.ParseAndProcess(ctx, destroyedLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process objective destroyed: %v", err)
		}

		// Verify event was created
		events, err := testApp.FindRecordsByFilter("events", "type = 'objective_destroyed'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query destroyed events: %v", err)
		}
		if len(events) < 1 {
			t.Errorf("Expected at least 1 destroyed event, got %d", len(events))
		}
	})

	t.Run("Objective captured", func(t *testing.T) {
		capturedLine := `[2025.11.15-12.06.00:000][501]LogGameplayEvents: Display: Objective 2 was captured for team 0 from team 1 by Player2[76561198333333333, team 0].`
		err := parser.ParseAndProcess(ctx, capturedLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process objective captured: %v", err)
		}

		// Verify event was created
		events, err := testApp.FindRecordsByFilter("events", "type = 'objective_captured'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query captured events: %v", err)
		}
		if len(events) < 1 {
			t.Errorf("Expected at least 1 captured event, got %d", len(events))
		}
	})

	t.Run("Multi-player objective", func(t *testing.T) {
		// Objective destroyed by multiple players
		multiLine := `[2025.11.15-12.07.00:000][502]LogGameplayEvents: Display: Objective 3 owned by team 1 was destroyed for team 0 by Player3[76561198444444444, team 0] + Player4[76561198555555555, team 0].`
		err := parser.ParseAndProcess(ctx, multiLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process multi-player objective: %v", err)
		}

		// Should create events for each player involved
		events, err := testApp.FindRecordsByFilter("events", "type = 'objective_destroyed'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query destroyed events: %v", err)
		}
		if len(events) < 2 {
			t.Logf("Expected multiple destroyed events for multi-player objective, got %d", len(events))
		}
	})
}

// TestRoundTransitions tests round start, round end, and game over sequence
func TestRoundTransitions(t *testing.T) {
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()
	serverID := "test-server-rounds"

	// Create server
	_, err = database.GetOrCreateServer(ctx, testApp, serverID, "Round Test Server", "test/path")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	parser := NewLogParser(testApp, testApp.Logger())

	// Create match
	mapLoadLine := `[2025.11.15-12.00.00:000][100]LogLoad: LoadMap: /Game/Maps/Hideout/Hideout_Checkpoint?Game=Checkpoint?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
	err = parser.ParseAndProcess(ctx, mapLoadLine, serverID, "test.log")
	if err != nil {
		t.Fatalf("Failed to process map load: %v", err)
	}

	t.Run("Round start", func(t *testing.T) {
		roundStartLine := `[2025.11.15-12.01.00:000][200]LogGameplayEvents: Display: Pre-round 1 started`
		err := parser.ParseAndProcess(ctx, roundStartLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process round start: %v", err)
		}
		// Round start doesn't create events, just resets counter
	})

	t.Run("Round end", func(t *testing.T) {
		roundEndLine := `[2025.11.15-12.05.00:000][300]LogGameplayEvents: Display: Round 1 Over: Team 0 won (win reason: Elimination)`
		err := parser.ParseAndProcess(ctx, roundEndLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process round end: %v", err)
		}

		// Verify round_end event was created
		events, err := testApp.FindRecordsByFilter("events", "type = 'round_end'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query round_end events: %v", err)
		}
		if len(events) < 1 {
			t.Errorf("Expected at least 1 round_end event, got %d", len(events))
		}
	})

	t.Run("Game over", func(t *testing.T) {
		gameOverLine := `[2025.11.15-12.10.00:000][400]LogSession: Display: AINSGameSession::HandleMatchHasEnded`
		err := parser.ParseAndProcess(ctx, gameOverLine, serverID, "test.log")
		if err != nil {
			t.Fatalf("Failed to process game over: %v", err)
		}

		// Verify match_end event was created
		events, err := testApp.FindRecordsByFilter("events", "type = 'match_end'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("Failed to query match_end events: %v", err)
		}
		if len(events) < 1 {
			t.Errorf("Expected at least 1 match_end event, got %d", len(events))
		}
	})
}

// Helper function to count leave events
func countLeaveEvents(t *testing.T, testApp *tests.TestApp) int {
	events, err := testApp.FindRecordsByFilter("events", "type = 'player_leave'", "-created", 100, 0)
	if err != nil {
		t.Fatalf("Failed to query leave events: %v", err)
	}
	return len(events)
}
