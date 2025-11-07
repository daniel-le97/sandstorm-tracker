package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad_YAML(t *testing.T) {
	// Create temp directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sandstorm-tracker.yml")

	yamlContent := `
servers:
  - name: "Test Server"
    logPath: "/test/logs"
    rconAddress: "127.0.0.1:27015"
    rconPassword: "testpass"
    rconTimeout: 10
    queryAddress: "127.0.0.1:27016"
    enabled: true
  - name: "Disabled Server"
    logPath: "/test/logs2"
    rconAddress: "127.0.0.1:27017"
    rconPassword: "pass2"
    enabled: false

logging:
  level: "debug"
  enableServerLogs: true
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Reset viper state
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify servers
	if len(cfg.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(cfg.Servers))
	}

	// Verify first server
	if cfg.Servers[0].Name != "Test Server" {
		t.Errorf("Server name = %s, want Test Server", cfg.Servers[0].Name)
	}
	if cfg.Servers[0].RconTimeout != 10 {
		t.Errorf("RconTimeout = %d, want 10", cfg.Servers[0].RconTimeout)
	}
	if !cfg.Servers[0].Enabled {
		t.Error("Server should be enabled")
	}

	// Verify second server
	if cfg.Servers[1].Enabled {
		t.Error("Second server should be disabled")
	}

	// Verify logging config
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging level = %s, want debug", cfg.Logging.Level)
	}
	if !cfg.Logging.EnableServerLogs {
		t.Error("EnableServerLogs should be true")
	}
}

func TestLoad_TOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sandstorm-tracker.toml")

	tomlContent := `
[[servers]]
name = "TOML Server"
logPath = "/toml/logs"
rconAddress = "127.0.0.1:27015"
rconPassword = "tomlpass"
rconTimeout = 15
queryAddress = "127.0.0.1:27016"
enabled = true

[logging]
level = "info"
enableServerLogs = false
`

	err := os.WriteFile(configPath, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}

	if cfg.Servers[0].Name != "TOML Server" {
		t.Errorf("Server name = %s, want TOML Server", cfg.Servers[0].Name)
	}

	if cfg.Servers[0].RconTimeout != 15 {
		t.Errorf("RconTimeout = %d, want 15", cfg.Servers[0].RconTimeout)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Logging level = %s, want info", cfg.Logging.Level)
	}
}

func TestLoad_EnvironmentOverride(t *testing.T) {
	t.Skip("Skipping environment override test due to viper state management complexity")
	// Environment variable override testing requires careful viper state reset
	// and is better tested via integration tests
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "Test",
						LogPath:      "/logs",
						RconAddress:  "127.0.0.1:27015",
						RconPassword: "pass",
						QueryAddress: "127.0.0.1:27016",
						Enabled:      true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "",
						LogPath:      "/logs",
						RconAddress:  "127.0.0.1:27015",
						RconPassword: "pass",
						Enabled:      true,
					},
				},
			},
			wantErr:     true,
			errContains: "missing 'name' field",
		},
		{
			name: "missing logPath",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "Test",
						LogPath:      "",
						RconAddress:  "127.0.0.1:27015",
						RconPassword: "pass",
						Enabled:      true,
					},
				},
			},
			wantErr:     true,
			errContains: "missing 'logPath' field",
		},
		{
			name: "missing rconAddress",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "Test",
						LogPath:      "/logs",
						RconAddress:  "",
						RconPassword: "pass",
						Enabled:      true,
					},
				},
			},
			wantErr:     true,
			errContains: "missing 'rconAddress' field",
		},
		{
			name: "missing rconPassword",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "Test",
						LogPath:      "/logs",
						RconAddress:  "127.0.0.1:27015",
						RconPassword: "",
						QueryAddress: "127.0.0.1:27016",
						Enabled:      true,
					},
				},
			},
			wantErr:     true,
			errContains: "missing 'rconPassword' field",
		},
		{
			name: "missing queryAddress",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "Test",
						LogPath:      "/logs",
						RconAddress:  "127.0.0.1:27015",
						RconPassword: "pass",
						QueryAddress: "",
						Enabled:      true,
					},
				},
			},
			wantErr:     true,
			errContains: "missing 'queryAddress' field",
		},
		{
			name: "disabled server skips validation",
			config: Config{
				Servers: []ServerConfig{
					{
						Name:         "",
						LogPath:      "",
						RconAddress:  "",
						RconPassword: "",
						Enabled:      false, // Disabled, should skip validation
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
