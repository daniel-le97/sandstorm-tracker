package a2s

import (
	"bytes"
	"testing"
	"time"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.timeout != DEFAULT_TIMEOUT {
		t.Errorf("Expected timeout %v, got %v", DEFAULT_TIMEOUT, client.timeout)
	}
}

// TestNewClientWithTimeout tests client creation with custom timeout
func TestNewClientWithTimeout(t *testing.T) {
	customTimeout := 10 * time.Second
	client := NewClientWithTimeout(customTimeout)
	if client == nil {
		t.Fatal("NewClientWithTimeout returned nil")
	}
	if client.timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.timeout)
	}
}

// Example test - requires a running Insurgency: Sandstorm server
// To run: go test -v -run TestQueryInfo_Live
// Skip by default as it requires a live server
func TestQueryInfo_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live server test in short mode")
	}

	// Replace with your actual server address and query port
	// Default Insurgency query port is game port + 1 (e.g., 27102 for game port 27101)
	address := "localhost:27102"

	client := NewClient()
	info, err := client.QueryInfo(address)

	// If no server is running, this test will fail - that's expected
	if err != nil {
		t.Skipf("No server available at %s: %v", address, err)
		return
	}

	// Validate response
	if info.Name == "" {
		t.Error("Server name is empty")
	}
	if info.Map == "" {
		t.Error("Server map is empty")
	}

	t.Logf("Server: %s", info.Name)
	t.Logf("Map: %s", info.Map)
	t.Logf("Players: %d/%d", info.Players, info.MaxPlayers)
	t.Logf("Bots: %d", info.Bots)
	t.Logf("Game: %s", info.Game)
	t.Logf("Version: %s", info.Version)
}

// Example test - requires a running Insurgency: Sandstorm server
func TestQueryPlayers_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live server test in short mode")
	}

	address := "localhost:27102"

	client := NewClient()
	players, err := client.QueryPlayers(address)

	if err != nil {
		t.Skipf("No server available at %s: %v", address, err)
		return
	}

	t.Logf("Found %d players", len(players))
	for _, player := range players {
		t.Logf("  [%d] %s - Score: %d, Duration: %.2fs",
			player.Index, player.Name, player.Score, player.Duration)
	}
}

// Example test - requires a running Insurgency: Sandstorm server
func TestQueryRules_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live server test in short mode")
	}

	address := "localhost:27102"

	client := NewClient()
	rules, err := client.QueryRules(address)

	if err != nil {
		t.Skipf("No server available at %s: %v", address, err)
		return
	}

	t.Logf("Found %d rules", len(rules))
	for _, rule := range rules {
		t.Logf("  %s = %s", rule.Name, rule.Value)
	}
}

// TestReadString tests the readString helper function
func TestReadString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple string",
			input:    []byte("hello\x00"),
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "Empty string",
			input:    []byte("\x00"),
			expected: "",
			wantErr:  false,
		},
		{
			name:     "String with spaces",
			input:    []byte("hello world\x00"),
			expected: "hello world",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := readString(bytes.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("readString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("readString() = %v, want %v", result, tt.expected)
			}
		})
	}
}
