package events

import (
	"testing"
	"time"
)

func TestTimestampParsing(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expected  time.Time
	}{
		{
			name:      "Standard timestamp",
			timestamp: "2025.10.04-15.23.38:790",
			expected:  time.Date(2025, 10, 4, 15, 23, 38, 790000000, time.UTC),
		},
		{
			name:      "Another timestamp",
			timestamp: "2025.10.04-13.46.26:141",
			expected:  time.Date(2025, 10, 4, 13, 46, 26, 141000000, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimestamp(tt.timestamp)
			if err != nil {
				t.Fatalf("parseTimestamp(%q) returned error: %v", tt.timestamp, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("parseTimestamp(%q) = %v, want %v", tt.timestamp, result, tt.expected)
			}
		})
	}
}

func TestPlayerKillRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		killer      string
		killerSteam string
		killerTeam  string
		victim      string
		victimSteam string
		victimTeam  string
		weapon      string
	}{
		{
			name:        "Player kills AI bot",
			logLine:     "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Marksman[INVALID, team 1] with BP_Firearm_M16A4_C_2147481419",
			shouldMatch: true,
			killer:      "ArmoredBear",
			killerSteam: "76561198995742987",
			killerTeam:  "0",
			victim:      "Marksman",
			victimSteam: "INVALID",
			victimTeam:  "1",
			weapon:      "BP_Firearm_M16A4_C_2147481419",
		},
		{
			name:        "Team kill",
			logLine:     "[2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Rabbit[76561198995742956, team 0] with BP_Firearm_M16A4_C_2147481419",
			shouldMatch: true,
			killer:      "ArmoredBear",
			killerSteam: "76561198995742987",
			killerTeam:  "0",
			victim:      "Rabbit",
			victimSteam: "76561198995742956",
			victimTeam:  "0",
			weapon:      "BP_Firearm_M16A4_C_2147481419",
		},
		{
			name:        "Suicide",
			logLine:     "[2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Character_Player_C_2147481498",
			shouldMatch: true,
			killer:      "ArmoredBear",
			killerSteam: "76561198995742987",
			killerTeam:  "0",
			victim:      "ArmoredBear",
			victimSteam: "76561198995742987",
			victimTeam:  "0",
			weapon:      "BP_Character_Player_C_2147481498",
		},
		{
			name:        "AI bot suicide",
			logLine:     "[2025.10.04-19.51.47:382][508]LogGameplayEvents: Display: ? killed Observer[INVALID, team 1] with BP_Projectile_Mortar_HE_C_2147480348",
			shouldMatch: true,
			killer:      "?",
			killerSteam: "",
			killerTeam:  "",
			victim:      "Observer",
			victimSteam: "INVALID",
			victimTeam:  "1",
			weapon:      "BP_Projectile_Mortar_HE_C_2147480348",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.PlayerKill.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				// Check extracted values
				if matches[2] != tt.killer {
					t.Errorf("Expected killer %q, got %q", tt.killer, matches[2])
				}
				if matches[5] != tt.victim {
					t.Errorf("Expected victim %q, got %q", tt.victim, matches[5])
				}
				if matches[8] != tt.weapon {
					t.Errorf("Expected weapon %q, got %q", tt.weapon, matches[8])
				}
			} else {
				if len(matches) > 0 {
					t.Fatalf("Expected regex not to match but got matches for: %s", tt.logLine)
				}
			}
		})
	}
}

func TestPlayerJoinRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		player      string
	}{
		{
			name:        "Player join",
			logLine:     "[2025.10.04-15.35.33:666][194]LogNet: Join succeeded: ArmoredBear",
			shouldMatch: true,
			player:      "ArmoredBear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.PlayerJoin.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.player {
					t.Errorf("Expected player %q, got %q", tt.player, matches[2])
				}
			} else {
				if len(matches) > 0 {
					t.Fatalf("Expected regex not to match but got matches for: %s", tt.logLine)
				}
			}
		})
	}
}

func TestPlayerDisconnectRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		steamID     string
	}{
		{
			name:        "Player disconnect",
			logLine:     "[2025.10.04-13.50.55:457][944]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198995742987), Result: (EOS_Success)",
			shouldMatch: true,
			steamID:     "76561198995742987",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.PlayerDisconnect.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.steamID {
					t.Errorf("Expected steamID %q, got %q", tt.steamID, matches[2])
				}
			}
		})
	}
}

func TestPlayerRconLeaveRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		player      string
	}{
		{
			name:        "Player RCON leave",
			logLine:     "[2025.10.04-15.37.59:204][779]LogRcon: 127.0.0.1:58877 << say See you later, ArmoredBear!",
			shouldMatch: true,
			player:      "ArmoredBear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.PlayerRconLeave.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.player {
					t.Errorf("Expected player %q, got %q", tt.player, matches[2])
				}
			}
		})
	}
}

func TestRoundEndRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		team        string
		reason      string
	}{
		{
			name:        "Round end - GameMode",
			logLine:     "[2025.10.04-15.21.46:114][183]LogGameMode: Display: Round O ver: Team 1 won (win reason: Elimination)",
			shouldMatch: true,
			team:        "1",
			reason:      "Elimination",
		},
		{
			name:        "Round end - GameplayEvents with round number",
			logLine:     "[2025.10.04-15.21.46:114][183]LogGameplayEvents: Display: Round 2 Over: Team 1 won (win reason: Elimination)",
			shouldMatch: true,
			team:        "1",
			reason:      "Elimination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.RoundEnd.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				// Team is at index 3 (after round number which might be empty)
				teamIndex := 3
				reasonIndex := 4

				if matches[teamIndex] != tt.team {
					t.Errorf("Expected team %q, got %q", tt.team, matches[teamIndex])
				}
				if matches[reasonIndex] != tt.reason {
					t.Errorf("Expected reason %q, got %q", tt.reason, matches[reasonIndex])
				}
			}
		})
	}
}

func TestGameOverRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
	}{
		{
			name:        "Game over",
			logLine:     "[2025.10.04-15.23.38:790][979]LogGameplayEvents: Display: Game over",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.GameOver.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}
			}
		})
	}
}

func TestMapLoadRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		mapName     string
		scenario    string
		maxPlayers  string
		lighting    string
	}{
		{
			name:        "Map load",
			logLine:     "[2025.10.04-13.46.26:141][  0]LogLoad: LoadMap: /Game/Maps/Canyon/Canyon?Name=Player?Scenario=Scenario_Crossing_Checkpoint_Security?MaxPlayers=8?Lighting=Day",
			shouldMatch: true,
			mapName:     "Canyon",
			scenario:    "Scenario_Crossing_Checkpoint_Security",
			maxPlayers:  "8",
			lighting:    "Day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.MapLoad.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.mapName {
					t.Errorf("Expected mapName %q, got %q", tt.mapName, matches[2])
				}
				if matches[3] != tt.scenario {
					t.Errorf("Expected scenario %q, got %q", tt.scenario, matches[3])
				}
				if matches[4] != tt.maxPlayers {
					t.Errorf("Expected maxPlayers %q, got %q", tt.maxPlayers, matches[4])
				}
				if matches[5] != tt.lighting {
					t.Errorf("Expected lighting %q, got %q", tt.lighting, matches[5])
				}
			}
		})
	}
}

func TestDifficultyChangeRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		difficulty  string
	}{
		{
			name:        "Difficulty change",
			logLine:     "[2025.10.04-15.34.56:470][  0]LogAI: Warning: AI difficulty set to 0.5",
			shouldMatch: true,
			difficulty:  "0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.DifficultyChange.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.difficulty {
					t.Errorf("Expected difficulty %q, got %q", tt.difficulty, matches[2])
				}
			}
		})
	}
}

func TestMapVoteRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
	}{
		{
			name:        "Map vote existing",
			logLine:     "[2025.10.04-15.23.38:812][979]LogMapVoteManager: Display: Existing Vote Options:",
			shouldMatch: true,
		},
		{
			name:        "Map vote new",
			logLine:     "[2025.10.04-15.23.38:812][979]LogMapVoteManager: Display: New Vote Options:",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.MapVote.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}
			}
		})
	}
}

func TestChatCommandRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		player      string
		steamID     string
		command     string
	}{
		{
			name:        "Stats command",
			logLine:     "[2025.10.04-16.42.23:199][613]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats",
			shouldMatch: true,
			player:      "ArmoredBear",
			steamID:     "76561198995742987",
			command:     "!stats",
		},
		{
			name:        "Stats with player name",
			logLine:     "[2025.10.04-16.42.23:199][613]LogChat: Display: Rabbit(76561198995742956) Global Chat: !stats Armoredbear",
			shouldMatch: true,
			player:      "Rabbit",
			steamID:     "76561198995742956",
			command:     "!stats Armoredbear",
		},
		{
			name:        "KDR command",
			logLine:     "[2025.10.04-16.42.26:896][833]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !kdr",
			shouldMatch: true,
			player:      "ArmoredBear",
			steamID:     "76561198995742987",
			command:     "!kdr",
		},
		{
			name:        "Top command",
			logLine:     "[2025.10.04-16.42.26:896][833]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !top",
			shouldMatch: true,
			player:      "ArmoredBear",
			steamID:     "76561198995742987",
			command:     "!top",
		},
		{
			name:        "Guns command",
			logLine:     "[2025.10.04-16.42.31:683][118]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !guns",
			shouldMatch: true,
			player:      "ArmoredBear",
			steamID:     "76561198995742987",
			command:     "!guns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.ChatCommand.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.player {
					t.Errorf("Expected player %q, got %q", tt.player, matches[2])
				}
				if matches[3] != tt.steamID {
					t.Errorf("Expected steamID %q, got %q", tt.steamID, matches[3])
				}
				if matches[4] != tt.command {
					t.Errorf("Expected command %q, got %q", tt.command, matches[4])
				}
			}
		})
	}
}

func TestFallDamageRegex(t *testing.T) {
	patterns := NewLogPatterns()

	tests := []struct {
		name        string
		logLine     string
		shouldMatch bool
		damage      string
	}{
		{
			name:        "Fall damage",
			logLine:     "[2025.10.04-15.12.17:472][441]LogSoldier: Applying 268.43 fall damage, downward velocity on landing was -1821.08",
			shouldMatch: true,
			damage:      "268.43",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := patterns.FallDamage.FindStringSubmatch(tt.logLine)

			if tt.shouldMatch {
				if len(matches) == 0 {
					t.Fatalf("Expected regex to match but got no matches for: %s", tt.logLine)
				}

				if matches[2] != tt.damage {
					t.Errorf("Expected damage %q, got %q", tt.damage, matches[2])
				}
			}
		})
	}
}
