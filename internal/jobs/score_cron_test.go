package jobs

import (
	"testing"
)

func TestParseRconListPlayers(t *testing.T) {
	// Example response from RCON listplayers
	response := `ID       | Name                          | NetID                         | IP                    | Score                 |
===============================================================================
256      | ArmoredBear   | SteamNWI:76561198995742987    | 127.0.0.1     | 20            | 0     |                               | None:INVALID  |       | 0  
| 0      | Observer              | None:INVALID  |       | 0             | 0     | Commander     | None:INVALID  |       | 0             | 0     | Marksman   | None:INVALID   |       | 0             |`

	players := parseRconListPlayers(response)

	if len(players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(players))
		for i, p := range players {
			t.Logf("Player %d: Name=%s, Score=%d, NetID=%s", i, p.Name, p.Score, p.NetID)
		}
		return
	}

	player := players[0]
	if player.Name != "ArmoredBear" {
		t.Errorf("Expected name 'ArmoredBear', got '%s'", player.Name)
	}

	if player.Score != 20 {
		t.Errorf("Expected score 20, got %d", player.Score)
	}

	if player.NetID != "SteamNWI:76561198995742987" {
		t.Errorf("Expected NetID 'SteamNWI:76561198995742987', got '%s'", player.NetID)
	}
}
