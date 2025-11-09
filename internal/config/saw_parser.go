package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// SAWServerConfig represents a single server configuration from SAW's server-configs.json
type SAWServerConfig struct {
	ID                 string `json:"id"`
	ServerConfigName   string `json:"server-config-name"`
	ServerHostname     string `json:"server_hostname"`
	ServerRconEnabled  string `json:"server_rcon_enabled"`
	ServerRconPort     string `json:"server_rcon_port"`
	ServerRconPassword string `json:"server_rcon_password"`
	ServerQueryPort    string `json:"server_query_port"`
	ServerGamePort     string `json:"server_game_port"`
}

// LoadFromSAW loads server configurations from Sandstorm Admin Wrapper's server-configs.json
// sawPath should be the root directory of the SAW installation
func LoadFromSAW(sawPath string) (*Config, error) {
	// Read server-configs.json
	configPath := filepath.Join(sawPath, "admin-interface", "config", "server-configs.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SAW config at %s: %w", configPath, err)
	}

	// Parse JSON - it's a map of server ID to config
	var sawConfigs map[string]SAWServerConfig
	if err := json.Unmarshal(data, &sawConfigs); err != nil {
		return nil, fmt.Errorf("failed to parse SAW config: %w", err)
	}

	// Convert SAW configs to our internal config format
	config := &Config{
		Servers: make([]ServerConfig, 0, len(sawConfigs)),
		Logging: LoggingConfig{
			Level:            "info",
			EnableServerLogs: true,
		},
	}

	for serverID, sawConfig := range sawConfigs {
		// Skip if RCON is disabled
		if sawConfig.ServerRconEnabled != "true" {
			continue
		}

		// Construct log path: {SAW_PATH}\sandstorm-server\Insurgency\Saved\Logs\{SERVER_ID}.log
		logPath := filepath.Join(sawPath, "sandstorm-server", "Insurgency", "Saved", "Logs", serverID+".log")

		// Default timeout
		timeout := 5

		server := ServerConfig{
			Name:         sawConfig.ServerHostname,
			LogPath:      logPath,
			RconAddress:  fmt.Sprintf("127.0.0.1:%s", sawConfig.ServerRconPort),
			RconPassword: sawConfig.ServerRconPassword,
			RconTimeout:  timeout,
			QueryAddress: fmt.Sprintf("127.0.0.1:%s", sawConfig.ServerQueryPort),
			Enabled:      true,
		}

		config.Servers = append(config.Servers, server)
	}

	if len(config.Servers) == 0 {
		return nil, fmt.Errorf("no enabled servers found in SAW config (RCON must be enabled)")
	}

	return config, nil
}

// LoadWithSAWPath loads configuration, checking for SAW path first, then falling back to manual config
func LoadWithSAWPath(sawPath string) (*Config, error) {
	// If SAW path is provided, use that
	if sawPath != "" {
		return LoadFromSAW(sawPath)
	}

	// Otherwise fall back to manual config file
	return Load()
}

// Helper to parse SAW boolean strings
func parseSAWBool(s string) bool {
	return s == "true"
}

// Helper to parse SAW int strings
func parseSAWInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
