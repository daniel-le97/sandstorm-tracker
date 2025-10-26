package config

import (
	// "flag"
	// "fmt"
	// "log"
	// "os"
	// "os/signal"
	// "sandstorm-tracker/internal/db"
	// "sandstorm-tracker/internal/utils"
	// "sandstorm-tracker/internal/watcher"
	// "strings"
	// "syscall"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Name         string `mapstructure:"name"`
	LogPath      string `mapstructure:"logPath"`
	RconAddress  string `mapstructure:"rconAddress"`
	RconPassword string `mapstructure:"rconPassword"`
	Enabled      bool   `mapstructure:"enabled"`
}

type DatabaseConfig struct {
	Path      string `mapstructure:"path"`
	EnableWAL bool   `mapstructure:"enableWAL"`
	CacheSize int    `mapstructure:"cacheSize"`
}

type LoggingConfig struct {
	Level            string `mapstructure:"level"`
	EnableServerLogs bool   `mapstructure:"enableServerLogs"`
}

type AppConfig struct {
	Servers  []ServerConfig `mapstructure:"servers"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

func InitConfig() (*AppConfig, error) {
	viper.SetConfigName("sandstorm-tracker") // name of config file (without extension)
	viper.SetConfigType("yml")               // or viper.SetConfigType("json")
	viper.AddConfigPath(".")                 // look for config in the working directory

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

	return &config, nil
}
