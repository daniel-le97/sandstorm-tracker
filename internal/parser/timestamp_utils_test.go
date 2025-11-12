package parser

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestReplaceTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		baseTime time.Time
		want     string
	}{
		{
			name:     "Join succeeded",
			line:     "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player",
			baseTime: time.Date(2025, 11, 11, 14, 30, 0, 123000000, time.UTC),
			want:     "[2025.11.11-14.30.00:123][866]LogNet: Join succeeded: Player",
		},
		{
			name:     "Kill event",
			line:     "[2025.10.04-21.29.24:398][699]LogGameplayEvents: Display: Player[123, team 0] killed Enemy[456, team 1] with Weapon",
			baseTime: time.Date(2025, 11, 11, 14, 30, 0, 0, time.UTC),
			want:     "[2025.11.11-14.30.00:000][699]LogGameplayEvents: Display: Player[123, team 0] killed Enemy[456, team 1] with Weapon",
		},
		{
			name:     "Chat command",
			line:     "[2025.10.04-21.28.22:312][561]LogChat: Display: Player(123) Global Chat: !stats",
			baseTime: time.Date(2025, 11, 11, 14, 30, 0, 500000000, time.UTC),
			want:     "[2025.11.11-14.30.00:500][561]LogChat: Display: Player(123) Global Chat: !stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceTimestamp(tt.line, tt.baseTime)
			if got != tt.want {
				t.Errorf("ReplaceTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceTimestampWithOffset(t *testing.T) {
	referenceTime := time.Date(2025, 10, 4, 21, 27, 0, 0, time.UTC)
	baseTime := time.Date(2025, 11, 11, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		line string
		want string // Expected timestamp should be baseTime + offset from reference
	}{
		{
			name: "51.78 seconds after reference",
			line: "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player",
			want: "[2025.11.11-14.30.51:780][866]LogNet: Join succeeded: Player",
		},
		{
			name: "144.398 seconds after reference",
			line: "[2025.10.04-21.29.24:398][699]LogGameplayEvents: Display: Kill event",
			want: "[2025.11.11-14.32.24:398][699]LogGameplayEvents: Display: Kill event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceTimestampWithOffset(tt.line, baseTime, referenceTime)
			if got != tt.want {
				t.Errorf("ReplaceTimestampWithOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateLogTimestamps(t *testing.T) {
	baseTime := time.Date(2025, 11, 11, 14, 30, 0, 0, time.UTC)

	lines := []string{
		"[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player1",
		"[2025.10.04-21.27.56:890][490]LogNet: Join succeeded: Player2",
		"[2025.10.04-21.28.22:312][561]LogChat: Display: Player1(123) Global Chat: !stats",
	}

	expected := []string{
		"[2025.11.11-14.30.00:000][866]LogNet: Join succeeded: Player1",                    // First timestamp = baseTime
		"[2025.11.11-14.30.05:110][490]LogNet: Join succeeded: Player2",                    // +5.11 seconds
		"[2025.11.11-14.30.30:532][561]LogChat: Display: Player1(123) Global Chat: !stats", // +30.532 seconds
	}

	result, err := UpdateLogTimestamps(lines, baseTime)
	if err != nil {
		t.Fatalf("UpdateLogTimestamps() error = %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf("UpdateLogTimestamps() returned %d lines, want %d", len(result), len(expected))
	}

	for i, got := range result {
		if got != expected[i] {
			t.Errorf("Line %d: got %v, want %v", i, got, expected[i])
		}
	}
}

func TestParseInsurgencyTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		ts      string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Valid timestamp",
			ts:      "2025.10.04-21.27.51:780",
			want:    time.Date(2025, 10, 4, 21, 27, 51, 780000000, time.UTC),
			wantErr: false,
		},
		{
			name:    "Single digit milliseconds",
			ts:      "2025.10.04-21.27.51:5",
			want:    time.Date(2025, 10, 4, 21, 27, 51, 5000000, time.UTC),
			wantErr: false,
		},
		{
			name:    "Zero milliseconds",
			ts:      "2025.10.04-21.27.51:000",
			want:    time.Date(2025, 10, 4, 21, 27, 51, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Invalid format",
			ts:      "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInsurgencyTimestamp(tt.ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInsurgencyTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseInsurgencyTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatInsurgencyTimestamp(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "Standard timestamp",
			time: time.Date(2025, 11, 11, 14, 30, 0, 123000000, time.UTC),
			want: "[2025.11.11-14.30.00:123]",
		},
		{
			name: "Zero milliseconds",
			time: time.Date(2025, 11, 11, 14, 30, 0, 0, time.UTC),
			want: "[2025.11.11-14.30.00:000]",
		},
		{
			name: "Single digit values",
			time: time.Date(2025, 1, 5, 9, 5, 7, 50000000, time.UTC),
			want: "[2025.01.05-09.05.07:050]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatInsurgencyTimestamp(tt.time)
			if got != tt.want {
				t.Errorf("formatInsurgencyTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseAllEventsWithUpdatedTimestamps tests that the parser correctly handles
// all event types from a real log file with updated timestamps
func TestParseAllEventsWithUpdatedTimestamps(t *testing.T) {
	// Read test log file
	testLogLines := []string{
		"Log file open, $TIME",
		"[2025.10.04-21.18.15:445][  0]LogLoad: LoadMap: /Game/Maps/Town/Town?Name=Player?Scenario=Scenario_Hideout_Checkpoint_Security?MaxPlayers=10?Game=CheckpointHardcore?Lighting=Day",
		"[2025.10.04-21.27.51:780][866]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (76561198262186571) Result: (EOS_Success)",
		"[2025.10.04-21.27.51:781][866]LogNet: Join succeeded: -=312th=- Rabbit",
		"[2025.10.04-21.28.22:312][561]LogChat: Display: -=312th=- Rabbit(76561198262186571) Global Chat: !maplist",
		"[2025.10.04-21.29.24:398][699]LogGameplayEvents: Display: *OSS*0rigin[76561198007416544, team 0] killed Rifleman[INVALID, team 1] with BP_Firearm_M4A1_C_2147480587",
		"[2025.10.04-21.30.15:593][108]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (76561198995742987) Result: (EOS_Success)",
		"[2025.10.04-21.30.15:594][108]LogNet: Join succeeded: ArmoredBear",
		"[2025.10.04-21.31.01:003][404]LogGameplayEvents: Display: Objective 0 was captured for team 0 from team 1 by *OSS*0rigin[76561198007416544], -=312th=- Rabbit[76561198262186571].",
		"[2025.10.04-21.32.55:614][297]LogGameplayEvents: Display: Objective 1 owned by team 1 was destroyed for team 0 by -=312th=- Rabbit[76561198262186571], ArmoredBear[76561198995742987].",
		"[2025.10.04-21.36.00:296][254]LogGameplayEvents: Display: Rifleman[INVALID, team 1] killed ArmoredBear[76561198995742987, team 0] with BP_Firearm_M16A4_C_2147473468",
	}

	// Update timestamps to be within the last hour
	baseTime := time.Now().Add(-30 * time.Minute)
	updatedLines, err := UpdateLogTimestamps(testLogLines, baseTime)
	if err != nil {
		t.Fatalf("UpdateLogTimestamps() error = %v", err)
	}

	// Verify timestamps were updated
	if strings.Contains(updatedLines[2], "2025.10.04") {
		t.Errorf("Timestamp was not updated in line: %s", updatedLines[2])
	}

	// Create a test parser instance (we can't fully test parsing without DB, but we can verify patterns)
	parser := &LogParser{
		patterns: newLogPatterns(),
	}

	// Test that each event type can be matched
	eventTests := []struct {
		name    string
		line    string
		pattern *regexp.Regexp
	}{
		{"LoadMap", updatedLines[1], parser.patterns.MapLoad},
		{"PlayerRegister", updatedLines[2], parser.patterns.PlayerRegister},
		{"PlayerJoin", updatedLines[3], parser.patterns.PlayerJoin},
		{"ChatCommand", updatedLines[4], parser.patterns.ChatCommand},
		{"PlayerKill", updatedLines[5], parser.patterns.PlayerKill},
		{"ObjectiveCaptured", updatedLines[8], parser.patterns.ObjectiveCaptured},
		{"ObjectiveDestroyed", updatedLines[9], parser.patterns.ObjectiveDestroyed},
	}

	for _, tt := range eventTests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.pattern.MatchString(tt.line) {
				t.Errorf("%s pattern did not match line: %s", tt.name, tt.line)
			}
		})
	}
}

// TestLoginRequestEvent specifically tests the new login request parser
func TestLoginRequestEvent(t *testing.T) {
	line := "[2025.11.10-20.58.50:166][881]LogNet: Login request: ?InitialConnectTimeout=30?Name=ArmoredBear userId: SteamNWI:76561198995742987 platform: SteamNWI"

	parser := &LogParser{
		patterns: newLogPatterns(),
	}

	if !parser.patterns.PlayerLogin.MatchString(line) {
		t.Errorf("PlayerLogin pattern did not match line: %s", line)
	}

	matches := parser.patterns.PlayerLogin.FindStringSubmatch(line)
	if len(matches) < 4 {
		t.Fatalf("Expected at least 4 capture groups, got %d", len(matches))
	}

	// Verify captured data
	timestamp := matches[1]
	playerName := matches[2]
	steamID := matches[3]
	platform := matches[4]

	if timestamp != "2025.11.10-20.58.50:166" {
		t.Errorf("Expected timestamp '2025.11.10-20.58.50:166', got '%s'", timestamp)
	}
	if playerName != "ArmoredBear" {
		t.Errorf("Expected player name 'ArmoredBear', got '%s'", playerName)
	}
	if steamID != "76561198995742987" {
		t.Errorf("Expected Steam ID '76561198995742987', got '%s'", steamID)
	}
	if platform != "SteamNWI" {
		t.Errorf("Expected platform 'SteamNWI', got '%s'", platform)
	}

	// Test with updated timestamp
	baseTime := time.Now()
	updatedLine := ReplaceTimestamp(line, baseTime)

	if !parser.patterns.PlayerLogin.MatchString(updatedLine) {
		t.Errorf("PlayerLogin pattern did not match updated line: %s", updatedLine)
	}
}
