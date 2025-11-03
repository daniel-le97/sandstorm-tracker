package app

import (
	"fmt"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Name         string `mapstructure:"name"`
	LogPath      string `mapstructure:"logPath"`
	RconAddress  string `mapstructure:"rconAddress"`
	RconPassword string `mapstructure:"rconPassword"`
	QueryAddress string `mapstructure:"queryAddress"`
	Enabled      bool   `mapstructure:"enabled"`
}

type LoggingConfig struct {
	Level            string `mapstructure:"level"`
	EnableServerLogs bool   `mapstructure:"enableServerLogs"`
}

type AppConfig struct {
	Servers []ServerConfig `mapstructure:"servers"`
	Logging LoggingConfig  `mapstructure:"logging"`
}

func InitConfig() (*AppConfig, error) {
	viper.SetConfigName("sandstorm-tracker") // name of config file (without extension)
	viper.AddConfigPath(".")                 // look for config in the working directory

	// Enable automatic environment variable reading
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SANDSTORM") // Optional: all env vars must start with SANDSTORM_

	// Try YAML first, then TOML
	viper.SetConfigType("yml")
	err := viper.ReadInConfig()
	if err != nil {
		viper.SetConfigType("toml")
		err = viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	var config AppConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
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

	return &config, nil
}

// EnsureServersInDatabase creates server records in PocketBase for all servers in config
func (c *AppConfig) EnsureServersInDatabase(pbApp core.App) error {
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
		serverID, err := GetServerIdFromPath(absPath)
		if err != nil {
			return fmt.Errorf("failed to extract server ID from path %s: %w", absPath, err)
		}

		// Check if server already exists by path
		exists, err := pbApp.FindRecordsByFilter(
			"servers",
			"path = {:path}",
			"",
			1,
			0,
			map[string]any{"path": absPath},
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
