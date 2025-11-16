package parser

import (
	"testing"
)

// MockRconSender for testing
type MockRconSender struct {
	lastCommand  string
	lastServerID string
}

func (m *MockRconSender) SendRconCommand(serverID string, command string) (string, error) {
	m.lastServerID = serverID
	m.lastCommand = command
	return "", nil
}

func TestChatCommandParsing(t *testing.T) {
	tests := []struct {
		name    string
		logLine string
		want    bool
	}{
		{
			name:    "!stats command",
			logLine: "[2025.10.21-20.09.21:472][427]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats",
			want:    true,
		},
		{
			name:    "!kdr command",
			logLine: "[2025.10.21-20.10.00:497][763]LogChat: Display: TestPlayer(76561198995742999) Global Chat: !kdr",
			want:    true,
		},
		{
			name:    "!top command",
			logLine: "[2025.10.21-20.10.31:940][644]LogChat: Display: Player1(76561198995742988) Global Chat: !top",
			want:    true,
		},
		{
			name:    "!guns command",
			logLine: "[2025.10.21-20.11.28:617][ 34]LogChat: Display: Player2(76561198995742989) Global Chat: !guns",
			want:    true,
		},
		{
			name:    "!map command (should match pattern but not be handled)",
			logLine: "[2025.10.21-20.09.21:472][427]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !map",
			want:    true,
		},
		{
			name:    "!maplist command (should match pattern but not be handled)",
			logLine: "[2025.10.21-20.09.21:472][427]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !maplist",
			want:    true,
		},
		{
			name:    "regular chat (not a command)",
			logLine: "[2025.10.21-20.11.28:617][ 34]LogChat: Display: Player3(76561198995742990) Global Chat: hello world",
			want:    false,
		},
		{
			name:    "not a chat line",
			logLine: "[2025.10.21-20.11.28:617][ 34]LogKill: Display: Player killed enemy",
			want:    false,
		},
	}

	parser := &LogParser{
		patterns: NewLogPatterns(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if the pattern matches
			matches := parser.patterns.ChatCommand.FindStringSubmatch(tt.logLine)
			got := len(matches) >= 5

			if got != tt.want {
				t.Errorf("ChatCommand pattern match = %v, want %v", got, tt.want)
				if got {
					t.Logf("Matches: %v", matches)
				}
			}
		})
	}
}

func TestChatCommandHandler(t *testing.T) {
	// This would require setting up a test database
	// For now, we just test that the parser recognizes the commands
	mockRcon := &MockRconSender{}

	parser := &LogParser{
		patterns: NewLogPatterns(),
	}

	handler := NewChatCommandHandler(parser, mockRcon)

	if handler == nil {
		t.Fatal("Failed to create chat command handler")
	}

	if handler.rconSender != mockRcon {
		t.Error("RCON sender not properly set")
	}
}

func TestUnsupportedCommandsIgnored(t *testing.T) {
	// Test that unsupported commands like !map, !maplist are detected but not handled
	unsupportedCommands := []string{
		"!map",
		"!maplist",
		"!rank",
		"!votemap",
		"!help",
		"!admin",
	}

	parser := &LogParser{
		patterns: NewLogPatterns(),
	}

	for _, cmd := range unsupportedCommands {
		t.Run(cmd, func(t *testing.T) {
			logLine := "[2025.10.21-20.09.21:472][427]LogChat: Display: TestPlayer(76561198995742987) Global Chat: " + cmd

			// Pattern should match (because it starts with !)
			matches := parser.patterns.ChatCommand.FindStringSubmatch(logLine)
			if len(matches) < 5 {
				t.Errorf("Command %s should match ChatCommand pattern but didn't", cmd)
				return
			}

			// Verify the command was extracted correctly
			extractedCmd := matches[4]
			if extractedCmd != cmd {
				t.Errorf("Expected command %s, got %s", cmd, extractedCmd)
			}

			// Note: The actual handler will ignore these commands (no case in switch statement)
			// This test just verifies the pattern matches so we can log them
			t.Logf("Command %s correctly detected (will be ignored by handler)", cmd)
		})
	}
}
