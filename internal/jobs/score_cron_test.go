package jobs

import (
	"log/slog"
	"testing"
	"time"
)

func TestParseRconListPlayers(t *testing.T) {
	// Real response format from RCON listplayers with tabs
	response := "ID\t | Name\t\t\t\t | NetID\t\t\t | IP\t\t\t | Score\t\t |\n" +
		"===============================================================================\n" +
		"0\t | \t\t\t\t | None:INVALID\t | \t | 0\t\t | " +
		"256\t | ArmoredBear\t | SteamNWI:76561198995742987\t | 127.0.0.1\t | 150\t\t | " +
		"0\t | Observer\t\t | None:INVALID\t | \t | 10\t\t | " +
		"0\t | Commander\t | None:INVALID\t | \t | 0\t\t | " +
		"0\t | Marksman\t\t | None:INVALID\t | \t | 0\t\t | "

	t.Logf("Raw response: %q", response)
	players := parseRconListPlayers(response)

	if len(players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(players))
		for i, p := range players {
			t.Logf("Player %d: Name=%s, Score=%d, NetID=%s", i, p.Name, p.Score, p.NetID)
		}
		// Debug: print lines
		lines := splitLines(response)
		t.Logf("Total lines: %d", len(lines))
		for i, line := range lines {
			t.Logf("Line %d (len=%d): %q", i, len(line), line)
		}
		return
	}

	player := players[0]
	if player.Name != "ArmoredBear" {
		t.Errorf("Expected name 'ArmoredBear', got '%s'", player.Name)
	}

	if player.Score != 150 {
		t.Errorf("Expected score 150, got %d", player.Score)
	}

	if player.NetID != "SteamNWI:76561198995742987" {
		t.Errorf("Expected NetID 'SteamNWI:76561198995742987', got '%s'", player.NetID)
	}
}

func TestParseRconListPlayers_SkipsInvalidEntries(t *testing.T) {
	// Test that invalid entries are properly filtered out
	response := "ID\t | Name\t\t\t\t | NetID\t\t\t | IP\t\t\t | Score\t\t |\n" +
		"===============================================================================\n" +
		"0\t | Observer\t\t | None:INVALID\t | \t | 0\t\t | " +
		"0\t | Commander\t | None:INVALID\t | \t | 0\t\t | " +
		"0\t | Marksman\t\t | None:INVALID\t | \t | 0\t\t | "

	players := parseRconListPlayers(response)

	if len(players) != 0 {
		t.Errorf("Expected 0 players, got %d", len(players))
		for i, p := range players {
			t.Logf("Unexpected player %d: Name=%s, Score=%d, NetID=%s", i, p.Name, p.Score, p.NetID)
		}
	}
}

func TestScoreDebouncer_TriggerScoreUpdateFixed(t *testing.T) {
	// Test that TriggerScoreUpdateFixed creates a timer with fixed delay
	// and clears the firstTriggerAt (indicating it's not using debounce logic)
	debouncer := &ScoreDebouncer{
		timers:         make(map[string]*time.Timer),
		firstTriggerAt: make(map[string]time.Time),
		logger:         slog.Default(),
		debounceWindow: 5 * time.Second,
		maxWait:        15 * time.Second,
	}

	// Set a firstTriggerAt to ensure it gets cleared
	debouncer.mu.Lock()
	debouncer.firstTriggerAt["test-server"] = time.Now()
	debouncer.mu.Unlock()

	// Trigger with fixed delay
	debouncer.TriggerScoreUpdateFixed("test-server", 100*time.Millisecond)

	// Check that timer was created
	debouncer.mu.Lock()
	_, timerExists := debouncer.timers["test-server"]
	_, firstTriggerExists := debouncer.firstTriggerAt["test-server"]
	debouncer.mu.Unlock()

	if !timerExists {
		t.Error("Expected timer to be created")
	}

	if firstTriggerExists {
		t.Error("Expected firstTriggerAt to be cleared when using fixed delay")
	}

	// Stop timer to prevent execution (we don't have a real app)
	debouncer.Stop()
	t.Log("TriggerScoreUpdateFixed correctly creates timer with fixed delay")
}

func TestScoreDebouncer_FixedDelayReplacesDebounce(t *testing.T) {
	// Test that fixed delay cancels existing timer and clears debounce state
	debouncer := &ScoreDebouncer{
		timers:         make(map[string]*time.Timer),
		firstTriggerAt: make(map[string]time.Time),
		logger:         slog.Default(),
		debounceWindow: 200 * time.Millisecond,
		maxWait:        500 * time.Millisecond,
	}

	originalFired := false

	// Create an existing debounce timer
	debouncer.mu.Lock()
	debouncer.firstTriggerAt["test-server"] = time.Now()
	debouncer.timers["test-server"] = time.AfterFunc(200*time.Millisecond, func() {
		originalFired = true
	})
	debouncer.mu.Unlock()

	// Trigger fixed delay - should cancel original
	debouncer.TriggerScoreUpdateFixed("test-server", 100*time.Millisecond)

	// Stop the new timer immediately to prevent execution
	debouncer.Stop()

	// Wait to ensure original doesn't fire
	time.Sleep(250 * time.Millisecond)

	if originalFired {
		t.Error("Original debounce timer should have been cancelled")
	}

	// Check that firstTriggerAt was cleared
	debouncer.mu.Lock()
	_, hasFirstTrigger := debouncer.firstTriggerAt["test-server"]
	debouncer.mu.Unlock()

	if hasFirstTrigger {
		t.Error("Expected firstTriggerAt to be cleared when using fixed delay")
	}

	t.Log("Fixed delay successfully cancelled and replaced debounce timer")
}
