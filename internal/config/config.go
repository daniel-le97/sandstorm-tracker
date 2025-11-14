package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sandstorm-tracker/assets"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Name         string `mapstructure:"name"`
	LogPath      string `mapstructure:"logPath"`
	RconAddress  string `mapstructure:"rconAddress"`
	RconPassword string `mapstructure:"rconPassword"`
	RconTimeout  int    `mapstructure:"rconTimeout"` // timeout in seconds, default 5
	QueryAddress string `mapstructure:"queryAddress"`
	Enabled      bool   `mapstructure:"enabled"`
}

type LoggingConfig struct {
	Level            string `mapstructure:"level"`
	EnableServerLogs bool   `mapstructure:"enableServerLogs"`
	MaxBackups       int    `mapstructure:"maxBackups"` // Number of rotated log files to keep (default: 5)
}

type Config struct {
	SAWPath string         `mapstructure:"sawPath"` // Path to Sandstorm Admin Wrapper installation
	Servers []ServerConfig `mapstructure:"servers"`
	Logging LoggingConfig  `mapstructure:"logging"`
}

func Load() (*Config, error) {
	viper.SetConfigName("sandstorm-tracker") // name of config file (without extension)
	viper.AddConfigPath(".")                 // look for config in the working directory

	// Enable automatic environment variable reading
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SANDSTORM") // Optional: all env vars must start with SANDSTORM_

	// Bind SAW_PATH environment variable
	viper.BindEnv("sawPath", "SAW_PATH")

	// Try YAML first, then TOML
	viper.SetConfigType("yml")
	err := viper.ReadInConfig()
	if err != nil {
		viper.SetConfigType("toml")
		err = viper.ReadInConfig()
		if err != nil {
			// No config file found - return empty config (will be handled by serve command)
			return &Config{
				Logging: LoggingConfig{
					Level:            "info",
					EnableServerLogs: true,
				},
			}, nil
		}
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	// If SAW path is provided, load from SAW and merge with manual config
	if config.SAWPath != "" {
		sawConfig, err := LoadFromSAW(config.SAWPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load from SAW: %w", err)
		}
		// Preserve logging config from file
		sawConfig.Logging = config.Logging
		sawConfig.SAWPath = config.SAWPath

		// Merge manual servers - they override SAW-discovered servers by name
		if len(config.Servers) > 0 {
			sawConfig.Servers = mergeServerConfigs(sawConfig.Servers, config.Servers)
		}

		return sawConfig, nil
	}

	// Dynamically bind environment variables for each server's RCON password
	// Format: RCON_PASSWORD_0, RCON_PASSWORD_1, etc. will override servers[i].rconPassword
	for i := range config.Servers {
		viper.BindEnv(fmt.Sprintf("servers.%d.rconPassword", i), fmt.Sprintf("RCON_PASSWORD_%d", i))
	}

	// Re-unmarshal to apply environment variable overrides
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	// Validate that all enabled servers have required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate checks that all enabled servers have required configuration fields
func (c *Config) Validate() error {
	for i, server := range c.Servers {
		if !server.Enabled {
			continue // Skip validation for disabled servers
		}

		if server.Name == "" {
			return fmt.Errorf("server at index %d is missing 'name' field", i)
		}

		if server.LogPath == "" {
			return fmt.Errorf("server '%s' (index %d) is missing 'logPath' field", server.Name, i)
		}

		if server.RconAddress == "" {
			return fmt.Errorf("server '%s' (index %d) is missing 'rconAddress' field", server.Name, i)
		}

		if server.RconPassword == "" {
			return fmt.Errorf("server '%s' (index %d) is missing 'rconPassword' field", server.Name, i)
		}

		if server.QueryAddress == "" {
			return fmt.Errorf("server '%s' (index %d) is missing 'queryAddress' field (A2S query port, usually game port + 29)", server.Name, i)
		}
	}

	return nil
}

// EnsureServersInDatabase creates server records in PocketBase for all servers in config
// This is a helper that bridges config and database concerns
func (c *Config) EnsureServersInDatabase(pbApp core.App, getServerID func(string) (string, error)) error {
	for _, serverCfg := range c.Servers {
		if !serverCfg.Enabled {
			continue
		}

		// Normalize path to absolute path for consistent comparison
		absPath, err := filepath.Abs(serverCfg.LogPath)
		if err != nil {
			absPath = serverCfg.LogPath
		}

		// Extract server UUID from the log file path
		serverID, err := getServerID(absPath)
		if err != nil {
			return fmt.Errorf("failed to extract server ID from path %s: %w", absPath, err)
		}

		// Check if server already exists by external_id
		exists, err := pbApp.FindRecordsByFilter(
			"servers",
			"external_id = {:external_id}",
			"",
			1,
			0,
			map[string]any{"external_id": serverID},
		)

		if err == nil && len(exists) > 0 {
			// Server already exists
			continue
		}

		// Create new server record
		collection, err := pbApp.FindCollectionByNameOrId("servers")
		if err != nil {
			return fmt.Errorf("failed to find servers collection: %w", err)
		}

		record := core.NewRecord(collection)
		record.Set("name", serverCfg.Name)  // Friendly name from config
		record.Set("external_id", serverID) // UUID from filename
		record.Set("path", absPath)

		if err := pbApp.Save(record); err != nil {
			return fmt.Errorf("failed to create server record for %s: %w", serverCfg.Name, err)
		}

		fmt.Printf("Created server: name='%s', external_id='%s', path='%s'\n", serverCfg.Name, serverID, absPath)
	}

	return nil
}

// mergeServerConfigs merges manual server configs into SAW-discovered configs
// Manual configs override SAW configs by name, or can disable servers with enabled:false
func mergeServerConfigs(sawServers []ServerConfig, manualServers []ServerConfig) []ServerConfig {
	// Build map of manual configs by name
	manualMap := make(map[string]ServerConfig)
	for _, srv := range manualServers {
		manualMap[srv.Name] = srv
	}

	// Merge: for each SAW server, check if there's a manual override
	result := make([]ServerConfig, 0, len(sawServers))
	for _, sawSrv := range sawServers {
		if manualSrv, exists := manualMap[sawSrv.Name]; exists {
			// Manual config exists - merge it (manual values override SAW values)
			merged := sawSrv // Start with SAW config

			// Override with manual values if they're set (non-zero)
			if manualSrv.LogPath != "" {
				merged.LogPath = manualSrv.LogPath
			}
			if manualSrv.RconAddress != "" {
				merged.RconAddress = manualSrv.RconAddress
			}
			if manualSrv.RconPassword != "" {
				merged.RconPassword = manualSrv.RconPassword
			}
			if manualSrv.RconTimeout > 0 {
				merged.RconTimeout = manualSrv.RconTimeout
			}
			if manualSrv.QueryAddress != "" {
				merged.QueryAddress = manualSrv.QueryAddress
			}

			// Enabled is always taken from manual config (allows disabling)
			merged.Enabled = manualSrv.Enabled

			result = append(result, merged)
			delete(manualMap, sawSrv.Name) // Mark as processed
		} else {
			// No manual override - use SAW config as-is
			result = append(result, sawSrv)
		}
	}

	// Add any manual servers that weren't in SAW config
	for _, manualSrv := range manualMap {
		result = append(result, manualSrv)
	}

	return result
}

// GenerateExample writes an example config file to the specified path
// format can be "yml" or "toml"
func GenerateExample(path string, format string) error {
	webAssets := assets.GetWebAssets()
	return webAssets.WriteExampleConfig(path, format)
}

// Exists checks if a config file exists in the current directory
func Exists() bool {
	// Check for YAML config
	if _, err := os.Stat("sandstorm-tracker.yml"); err == nil {
		return true
	}
	if _, err := os.Stat("sandstorm-tracker.yaml"); err == nil {
		return true
	}

	// Check for TOML config
	if _, err := os.Stat("sandstorm-tracker.toml"); err == nil {
		return true
	}

	return false
}
