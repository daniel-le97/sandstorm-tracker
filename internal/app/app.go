package app

import (
	"fmt"
	"sandstorm-tracker/internal/a2s"
	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/handlers"
	"sandstorm-tracker/internal/jobs"
	"sandstorm-tracker/internal/parser"
	"sandstorm-tracker/internal/rcon"
	"sandstorm-tracker/internal/util"
	"sandstorm-tracker/internal/watcher"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

// App wraps PocketBase with application-specific components and methods
type App struct {
	*pocketbase.PocketBase // Embed PocketBase - all its methods are available

	// Custom application components
	Config   *config.Config
	Parser   *parser.LogParser
	RconPool *rcon.ClientPool
	A2SPool  *a2s.ServerPool
	Watcher  *watcher.Watcher
}

// New creates and initializes the sandstorm-tracker application
func New() (*App, error) {
	app := &App{
		PocketBase: pocketbase.New(),
	}

	// Load configuration (lightweight - no validation yet)
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	app.Config = cfg

	// Initialize parser
	app.Parser = parser.NewLogParser(app.PocketBase)

	// Initialize RCON pool (servers added in onServe)
	app.RconPool = rcon.NewClientPool(app.PocketBase.Logger())

	// Initialize A2S pool (servers added in onServe)
	app.A2SPool = a2s.NewServerPool()

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

	// Auto-update from GitHub releases
	ghupdate.MustRegister(app.PocketBase, app.RootCmd, ghupdate.Config{
		Owner: "daniel-le97",
		Repo:  "sandstorm-trackerv2",
	})

	// Add other plugins here (jsvm, etc.)
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
	// Validate configuration before starting
	if err := app.Config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup servers in RCON and A2S pools
	for _, sc := range app.Config.Servers {
		if !sc.Enabled {
			continue
		}

		serverID, err := util.GetServerIdFromPath(sc.LogPath)
		if err != nil {
			return fmt.Errorf("failed to get server ID from path %s: %w", sc.LogPath, err)
		}

		// Add to RCON pool
		if sc.RconAddress != "" && sc.RconPassword != "" {
			timeout := 5 * time.Second
			if sc.RconTimeout > 0 {
				timeout = time.Duration(sc.RconTimeout) * time.Second
			}

			app.RconPool.AddServer(serverID, &rcon.ServerConfig{
				Address:  sc.RconAddress,
				Password: sc.RconPassword,
				Timeout:  timeout,
			})
		}

		// Add to A2S pool
		queryAddr := sc.RconAddress
		if sc.QueryAddress != "" {
			queryAddr = sc.QueryAddress
		}
		if queryAddr != "" {
			app.A2SPool.AddServer(queryAddr, sc.Name)
		}
	}

	// Initialize watcher with configured servers
	w, err := watcher.NewWatcher(app.PocketBase, app.Parser, app.RconPool, app.Config.Servers)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	app.Watcher = w

	// Ensure servers from config are in database
	err = app.Config.EnsureServersInDatabase(app.PocketBase, util.GetServerIdFromPath)
	if err != nil {
		return fmt.Errorf("failed to ensure servers in database: %w", err)
	}

	// Register web routes
	handlers.Register(app)

	// Register background jobs (A2S cron)
	jobs.RegisterA2S(app, app.Config)

	// Start file watcher
	for _, serverCfg := range app.Config.Servers {
		if serverCfg.Enabled {
			if err := app.Watcher.AddPath(serverCfg.LogPath); err != nil {
				return fmt.Errorf("failed to add path %s to watcher: %w", serverCfg.LogPath, err)
			}
		}
	}

	// Start watcher with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				app.Logger().Error("Watcher panic recovered", "panic", r)
			}
		}()
		app.Watcher.Start()
	}()

	return e.Next()
}

// onTerminate is called when the application shuts down
func (app *App) onTerminate(e *core.TerminateEvent) error {
	// Cleanup resources
	if app.Watcher != nil {
		app.Watcher.Stop()
	}

	if app.RconPool != nil {
		app.RconPool.CloseAll()
	}

	return e.Next()
}

// Custom application methods

// SendRconCommand sends an RCON command to a specific server
func (app *App) SendRconCommand(serverID string, command string) (string, error) {
	if app.RconPool == nil {
		return "", fmt.Errorf("RCON pool not initialized")
	}

	return app.RconPool.SendCommand(serverID, command)
}

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

// GetRconPoolStatus returns the current status of the RCON pool
func (app *App) GetRconPoolStatus() map[string]any {
	if app.RconPool == nil {
		return map[string]any{
			"available": false,
		}
	}

	servers := app.RconPool.ListServers()
	connectedServers := []string{}
	for _, serverID := range servers {
		if app.RconPool.IsConnected(serverID) {
			connectedServers = append(connectedServers, serverID)
		}
	}

	return map[string]any{
		"available":         true,
		"total_servers":     len(servers),
		"connected_servers": len(connectedServers),
		"connected_list":    connectedServers,
	}
}

// GetA2SPool returns the A2S server pool
func (app *App) GetA2SPool() *a2s.ServerPool {
	return app.A2SPool
}
