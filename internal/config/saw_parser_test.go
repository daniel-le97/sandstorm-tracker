package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromSAW(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Create SAW directory structure
	adminInterfaceDir := filepath.Join(tempDir, "admin-interface")
	if err := os.MkdirAll(adminInterfaceDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test server-configs.json
	testConfig := map[string]SAWServerConfig{
		"test-server-uuid-1": {
			ID:                 "test-server-uuid-1",
			ServerConfigName:   "Test Config",
			ServerHostname:     "Test Server 1",
			ServerRconEnabled:  "true",
			ServerRconPort:     "27015",
			ServerRconPassword: "password123",
			ServerQueryPort:    "27131",
			ServerGamePort:     "7777",
		},
		"test-server-uuid-2": {
			ID:                 "test-server-uuid-2",
			ServerConfigName:   "Test Config 2",
			ServerHostname:     "Test Server 2",
			ServerRconEnabled:  "false", // Should be skipped
			ServerRconPort:     "27016",
			ServerRconPassword: "password456",
			ServerQueryPort:    "27132",
			ServerGamePort:     "7778",
		},
	}

	configData, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	configPath := filepath.Join(adminInterfaceDir, "server-configs.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading from SAW
	config, err := LoadFromSAW(tempDir)
	if err != nil {
		t.Fatalf("LoadFromSAW() error = %v", err)
	}

	// Should only load servers with RCON enabled
	if len(config.Servers) != 1 {
		t.Errorf("Expected 1 server (RCON disabled server should be skipped), got %d", len(config.Servers))
	}

	server := config.Servers[0]

	// Verify server name
	if server.Name != "Test Server 1" {
		t.Errorf("Server name = %s, want Test Server 1", server.Name)
	}

	// Verify log path construction
	expectedLogPath := filepath.Join(tempDir, "sandstorm-server", "Insurgency", "Saved", "Logs", "test-server-uuid-1.log")
	if server.LogPath != expectedLogPath {
		t.Errorf("LogPath = %s, want %s", server.LogPath, expectedLogPath)
	}

	// Verify RCON address
	if server.RconAddress != "127.0.0.1:27015" {
		t.Errorf("RconAddress = %s, want 127.0.0.1:27015", server.RconAddress)
	}

	// Verify RCON password
	if server.RconPassword != "password123" {
		t.Errorf("RconPassword = %s, want password123", server.RconPassword)
	}

	// Verify query address
	if server.QueryAddress != "127.0.0.1:27131" {
		t.Errorf("QueryAddress = %s, want 127.0.0.1:27131", server.QueryAddress)
	}

	// Verify default timeout
	if server.RconTimeout != 5 {
		t.Errorf("RconTimeout = %d, want 5", server.RconTimeout)
	}

	// Verify enabled
	if !server.Enabled {
		t.Error("Server should be enabled")
	}
}

func TestLoadFromSAW_NoRconServers(t *testing.T) {
	tempDir := t.TempDir()
	adminInterfaceDir := filepath.Join(tempDir, "admin-interface")
	if err := os.MkdirAll(adminInterfaceDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create config with all RCON disabled
	testConfig := map[string]SAWServerConfig{
		"test-server-uuid": {
			ID:                "test-server-uuid",
			ServerHostname:    "Test Server",
			ServerRconEnabled: "false",
		},
	}

	configData, _ := json.MarshalIndent(testConfig, "", "  ")
	configPath := filepath.Join(adminInterfaceDir, "server-configs.json")
	os.WriteFile(configPath, configData, 0644)

	// Should return error when no enabled servers found
	_, err := LoadFromSAW(tempDir)
	if err == nil {
		t.Error("LoadFromSAW() should return error when no RCON-enabled servers found")
	}
}

func TestLoadFromSAW_MissingFile(t *testing.T) {
	tempDir := t.TempDir()

	_, err := LoadFromSAW(tempDir)
	if err == nil {
		t.Error("LoadFromSAW() should return error when server-configs.json doesn't exist")
	}
}
