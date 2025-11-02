package app

import (
	"fmt"

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
