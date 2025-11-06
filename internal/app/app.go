package app

import (
	"fmt"
	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/handlers"
	"sandstorm-tracker/internal/jobs"
	"sandstorm-tracker/internal/parser"
	"sandstorm-tracker/internal/util"
	"sandstorm-tracker/internal/watcher"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

// App wraps PocketBase with application-specific components and methods
type App struct {
	*pocketbase.PocketBase // Embed PocketBase - all its methods are available

	// Custom application components
	Config  *config.Config
	Parser  *parser.LogParser
	Watcher *watcher.Watcher
}

// New creates and initializes the sandstorm-tracker application
func New() (*App, error) {
	app := &App{
		PocketBase: pocketbase.New(),
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	app.Config = cfg

	// Initialize parser
	app.Parser = parser.NewLogParser(app.PocketBase)

	// Initialize watcher
	w, err := watcher.NewWatcher(app.PocketBase, cfg.Servers)
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}
	app.Watcher = w

	// Setup default plugins
	app.setupPlugins()

	return app, nil
}

// setupPlugins configures PocketBase plugins
func (app *App) setupPlugins() {
	// Auto-migrate database
	migratecmd.MustRegister(app.PocketBase, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// Add other plugins here (ghupdate, jsvm, etc.)
}

// Bootstrap initializes all application components and registers hooks
func (app *App) Bootstrap() error {
	// Register lifecycle hooks
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		return app.onServe(e)
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		return app.onTerminate(e)
	})

	return nil
}

// onServe is called when the server starts
func (app *App) onServe(e *core.ServeEvent) error {
	// Ensure servers from config are in database
	err := app.Config.EnsureServersInDatabase(app.PocketBase, util.GetServerIdFromPath)
	if err != nil {
		return fmt.Errorf("failed to ensure servers in database: %w", err)
	}

	// Register web routes
	handlers.Register(app.PocketBase)

	// Register background jobs (A2S cron)
	jobs.RegisterA2S(app.PocketBase, app.Config)

	// Start file watcher
	for _, serverCfg := range app.Config.Servers {
		if serverCfg.Enabled {
			if err := app.Watcher.AddPath(serverCfg.LogPath); err != nil {
				return fmt.Errorf("failed to add path %s to watcher: %w", serverCfg.LogPath, err)
			}
		}
	}

	go app.Watcher.Start()

	return e.Next()
}

// onTerminate is called when the application shuts down
func (app *App) onTerminate(e *core.TerminateEvent) error {
	// Cleanup resources
	if app.Watcher != nil {
		app.Watcher.Stop()
	}

	return e.Next()
}

// Custom application methods

// GetEnabledServers returns all enabled servers from config
func (app *App) GetEnabledServers() []config.ServerConfig {
	var enabled []config.ServerConfig
	for _, server := range app.Config.Servers {
		if server.Enabled {
			enabled = append(enabled, server)
		}
	}
	return enabled
}

// GetServerByName finds a server config by name
func (app *App) GetServerByName(name string) (*config.ServerConfig, error) {
	for _, server := range app.Config.Servers {
		if server.Name == name {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server '%s' not found in config", name)
}
