package jobs

import (
	"testing"
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
