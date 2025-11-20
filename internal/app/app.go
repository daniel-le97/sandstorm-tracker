package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"sandstorm-tracker/internal/a2s"
	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/ghupdate"
	"sandstorm-tracker/internal/handlers"
	"sandstorm-tracker/internal/jobs"
	"sandstorm-tracker/internal/logger"
	"sandstorm-tracker/internal/parser"
	"sandstorm-tracker/internal/rcon"
	"sandstorm-tracker/internal/updater"
	"sandstorm-tracker/internal/util"
	"sandstorm-tracker/internal/watcher"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/spf13/cobra"
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
	// ServerManager *servermgr.Plugin  // Server manager plugin
	// logFileWriter *logger.FileWriter // File writer for PocketBase logs
	customLogger *slog.Logger // Logger with TeeHandler (writes to both console and file)
	updater      *updater.Updater

	// Version information (injected at build time via ldflags)
	Version string
	Commit  string
	Date    string
}

// New creates and initializes the sandstorm-tracker application
func New() (*App, error) {
	return NewWithVersion("dev", "unknown", "unknown")
}

func (app *App) bindConfig() {
	app.Config = app.Store().GetOrSet("config", func() any {
		cfg, err := config.Load()
		if err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
		return cfg
	}).(*config.Config)
}
func (app *App) bindPoolers() {
	rconPool := app.Store().GetOrSet("rconpool", func() any {
		return rcon.NewClientPool(app.Logger().WithGroup("RCON"))
	}).(*rcon.ClientPool)
	app.RconPool = rconPool
}

func (app *App) setupParser() {
	app.Parser = app.Store().GetOrSet("parser", func() any {
		return parser.NewLogParser(app, app.Logger().With("component", "PARSER"))
	}).(*parser.LogParser)

	app.A2SPool = app.Store().GetOrSet("a2spool", func() any {
		return a2s.NewServerPool()
	}).(*a2s.ServerPool)
}

// NewWithVersion creates a new app with version information
func NewWithVersion(version, commit, date string) (*App, error) {
	app := &App{
		PocketBase: pocketbase.New(),
		Version:    version,
		Commit:     commit,
		Date:       date,
	}

	app.bindConfig()

	app.setupLogger()

	app.bindPoolers()

	// Note: Logger setup happens in OnBootstrap hook (after PocketBase is fully initialized)

	// Initialize parser with logger
	app.setupParser()

	// Initialize RCON pool (servers added in onServe)

	// Initialize A2S pool (servers added in onServe)

	// Setup default plugins (includes server manager)
	app.setupPlugins()

	return app, nil
}

// setupPlugins configures PocketBase plugins
func (app *App) setupPlugins() {
	// Auto-migrate database
	migratecmd.MustRegister(app.PocketBase, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	updater := ghupdate.MustRegister(app.PocketBase, app.RootCmd, ghupdate.Config{
		Owner:             "daniel-le97",
		Repo:              "sandstorm-tracker",
		ArchiveExecutable: "sandstorm-tracker",
	})

	// Register version command
	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("sandstorm-tracker version %s\n", app.Version)
			fmt.Printf("Commit: %s\n", app.Commit)
			fmt.Printf("Date: %s\n", app.Date)
			hasUpdates, err := updater.CheckForUpdate(context.Background())
			if err != nil {
				fmt.Printf("Update check failed: %v\n", err)
				return
			}
			if hasUpdates != "" {
				fmt.Println("A new update is available! Run 'sandstorm-tracker update' to update.")
			} else {
				fmt.Println("You are running the latest version.")
			}
		},
	})

	// Add other plugins here (jsvm, etc.)
}

// Bootstrap initializes all application components and registers hooks
func (app *App) Bootstrap() error {
	// Register lifecycle hooks
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Write PID file for graceful shutdown/update coordination
		pidData := []byte(fmt.Sprintf("%d", os.Getpid()))
		if err := os.WriteFile("sandstorm-tracker.pid", pidData, 0644); err != nil {
			app.Logger().Warn("Failed to write PID file", "error", err)
		}
		return app.onServe(e)
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		os.Remove("sandstorm-tracker.pid")
		return app.onTerminate(e)
	})

	return nil
}

// onServe is called when the server starts
func (app *App) onServe(e *core.ServeEvent) error {
	logger := app.Logger().With("component", "APP")
	logger.Info("Starting sandstorm-tracker application")
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
	w, err := watcher.NewWatcher(app.PocketBase, app.Parser, app.RconPool, app.A2SPool, app.Logger().With("component", "WATCHER"), app.Config.Servers)
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
	handlers.Register(app, e)

	// Create score debouncer for event-driven score updates
	// Scores update 10 seconds after any kill/objective/round event
	scoreDebouncer := jobs.NewScoreDebouncer(app, app.Config, 10*time.Second, 30*time.Second)
	app.Logger().Info("Initialized event-driven score updater", "component", "APP", "debounce", "10s", "maxWait", "30s")

	// Register event handlers for hook-based processing
	// Handlers process events created by the parser and trigger score updates
	gameEventHandlers := handlers.NewGameEventHandlers(app, scoreDebouncer)
	gameEventHandlers.RegisterHooks()
	app.Logger().Info("Registered game event handlers", "component", "APP")

	// Start file watcher
	for _, serverCfg := range app.Config.Servers {
		if serverCfg.Enabled {
			if err := app.Watcher.AddPath(serverCfg.LogPath); err != nil {
				return fmt.Errorf("failed to add path %s to watcher: %w", serverCfg.LogPath, err)
			}
		}
	}

	// Register archive cron job for data older than 30 days
	jobs.RegisterArchiveOldData(app.PocketBase, app.Logger().With("component", "ARCHIVE_JOB"))

	// Register update checker cron job (every 30 minutes)
	jobs.RegisterUpdateChecker(app, app.Config, app.Logger())

	// Start watcher with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				app.Logger().Error("Watcher panic recovered", "component", "WATCHER", "panic", r)
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

	// Note: ServerManager plugin handles its own cleanup via OnTerminate hook

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

func (app *App) Logger() *slog.Logger {
	if app.customLogger != nil {
		return app.customLogger
	}
	return app.PocketBase.Logger()
}

// GetUpdater returns the updater instance for use in hooks
func (app *App) GetUpdater() *updater.Updater {
	return app.updater
}

// setupLogFileWriter initializes the file writer.
// This must be called AFTER app.Config is loaded.
func (app *App) setupLogger() {
	// Determine log file path
	logFilePath := filepath.Join(".", "logs", "app.log")

	// Get max backups from config (default to 5 if not set)
	maxBackups := app.Config.Logging.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 5
	}

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		writer := app.Store().Get("logger:filewriter")
		if writer != nil {
			fw := writer.(*logger.FileWriter)
			if err := fw.Close(); err != nil {
				app.Logger().Error("Failed to close log file writer", "component", "APP", "error", err)
			}
		}
		return e.Next()
	})

	app.OnModelCreate(core.LogsTableName).BindFunc(func(e *core.ModelEvent) error {
		// get or create global file writer instance if not already created
		writer := e.App.Store().GetOrSet("logger:filewriter", func() any {
			fw, err := logger.NewFileWriter(logger.FileWriterConfig{
				FilePath:   logFilePath,
				MaxSize:    10 * 1024 * 1024, // 10MB max file size
				MaxBackups: maxBackups,
				BufferSize: 8192,                   // 8KB buffer
				FlushEvery: 500 * time.Millisecond, // Flush every 500ms (more responsive for development)
			})
			if err != nil {
				panic(fmt.Sprintf("failed to create file writer: %v", err))
			}

			return fw
		}).(*logger.FileWriter)
		l := e.Model.(*core.Log)
		writer.WriteLog(l)
		return e.Next()
	})
}
