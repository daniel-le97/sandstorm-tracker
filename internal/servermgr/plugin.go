package servermgr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// Config defines the plugin configuration
type Config struct {
	// DefaultSAWPath is the default path to Sandstorm Admin Wrapper installation
	DefaultSAWPath string
}

// Plugin manages Insurgency server processes as a PocketBase plugin
type Plugin struct {
	app     core.App
	config  Config
	mu      sync.RWMutex
	servers map[string]*ManagedServer
}

// MustRegister registers the server manager plugin (panics on error)
func MustRegister(app core.App, rootCmd *cobra.Command, config Config) *Plugin {
	p, err := Register(app, rootCmd, config)
	if err != nil {
		panic(err)
	}
	return p
}

// Register registers the server manager plugin
func Register(app core.App, rootCmd *cobra.Command, config Config) (*Plugin, error) {
	p := &Plugin{
		app:     app,
		config:  config,
		servers: make(map[string]*ManagedServer),
	}

	// Register CLI commands
	p.registerCommands(rootCmd)

	// Register API routes on serve
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		return p.registerRoutes(e)
	})

	// Cleanup on terminate
	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		p.StopAll()
		return e.Next()
	})

	return p, nil
}

// registerCommands adds server management CLI commands
func (p *Plugin) registerCommands(rootCmd *cobra.Command) {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Manage Insurgency servers",
		Long:  "Start, stop, and manage Insurgency: Sandstorm server instances",
	}

	// server start command
	startCmd := &cobra.Command{
		Use:   "start [server-id]",
		Short: "Start an Insurgency server",
		Long:  "Start an Insurgency server from SAW configuration. If no server ID is provided, the first available server will be started.",
		RunE: func(cmd *cobra.Command, args []string) error {
			showLogs, _ := cmd.Flags().GetBool("logs")
			sawPath, _ := cmd.Flags().GetString("saw-path")

			if sawPath == "" {
				sawPath = p.config.DefaultSAWPath
			}
			if sawPath == "" {
				return fmt.Errorf("SAW path not provided. Use --saw-path flag or set sawPath in config")
			}

			configs, err := p.LoadSAWConfigs(sawPath)
			if err != nil {
				return fmt.Errorf("failed to load SAW configs: %w", err)
			}

			var serverID string
			var serverConfig SAWServerConfig

			if len(args) > 0 {
				serverID = args[0]
				var ok bool
				serverConfig, ok = configs[serverID]
				if !ok {
					return fmt.Errorf("server ID '%s' not found in SAW configs", serverID)
				}
			} else {
				for id, cfg := range configs {
					serverID = id
					serverConfig = cfg
					break
				}
			}

			if serverID == "" {
				return fmt.Errorf("no servers found in SAW configuration")
			}

			fmt.Printf("Starting server: %s\n", serverID)
			if showLogs {
				fmt.Println("Server logs will be displayed in console (Press Ctrl+C to stop)")
			}

			return p.StartServer(serverID, serverConfig, sawPath, showLogs)
		},
	}
	startCmd.Flags().Bool("logs", false, "Show server logs in console")
	startCmd.Flags().String("saw-path", "", "Path to Sandstorm Admin Wrapper installation")

	// server stop command
	stopCmd := &cobra.Command{
		Use:   "stop [server-id]",
		Short: "Stop a running Insurgency server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID := args[0]
			fmt.Printf("Stopping server: %s\n", serverID)

			if err := p.StopServer(serverID); err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}

			fmt.Println("Server stopped successfully")
			return nil
		},
	}

	// server status command
	statusCmd := &cobra.Command{
		Use:   "status [server-id]",
		Short: "Check server status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				serverID := args[0]
				running, err := p.GetServerStatus(serverID)
				if err != nil {
					return err
				}

				status := "stopped"
				if running {
					status = "running"
				}
				fmt.Printf("Server %s: %s\n", serverID, status)
			} else {
				servers := p.ListServers()
				if len(servers) == 0 {
					fmt.Println("No managed servers found")
					return nil
				}

				fmt.Println("Managed Servers:")
				fmt.Println(strings.Repeat("-", 50))
				for id, running := range servers {
					status := "stopped"
					if running {
						status = "running"
					}
					fmt.Printf("%-40s %s\n", id, status)
				}
			}
			return nil
		},
	}

	// server list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available servers from SAW configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			sawPath, _ := cmd.Flags().GetString("saw-path")

			if sawPath == "" {
				sawPath = p.config.DefaultSAWPath
			}
			if sawPath == "" {
				return fmt.Errorf("SAW path not provided. Use --saw-path flag or set sawPath in config")
			}

			configs, err := p.LoadSAWConfigs(sawPath)
			if err != nil {
				return fmt.Errorf("failed to load SAW configs: %w", err)
			}

			if len(configs) == 0 {
				fmt.Println("No servers found in SAW configuration")
				return nil
			}

			fmt.Printf("Available Servers (%d found):\n", len(configs))
			fmt.Println(strings.Repeat("-", 80))
			for id, cfg := range configs {
				fmt.Printf("ID:          %s\n", id)
				fmt.Printf("Name:        %s\n", cfg.ServerHostname)
				fmt.Printf("Map:         %s\n", cfg.ServerDefaultMap)
				fmt.Printf("Mode:        %s\n", cfg.ServerScenarioMode)
				fmt.Printf("Port:        %s\n", cfg.ServerGamePort)
				fmt.Printf("Max Players: %s\n", cfg.ServerMaxPlayers)
				fmt.Println(strings.Repeat("-", 80))
			}

			return nil
		},
	}
	listCmd.Flags().String("saw-path", "", "Path to Sandstorm Admin Wrapper installation")

	serverCmd.AddCommand(startCmd, stopCmd, statusCmd, listCmd)
	rootCmd.AddCommand(serverCmd)
}

// registerRoutes adds API endpoints for server management
func (p *Plugin) registerRoutes(e *core.ServeEvent) error {
	// POST /api/server/start - Start a server
	e.Router.POST("/api/server/start", func(re *core.RequestEvent) error {
		data := struct {
			ServerID string `json:"server_id"`
			SAWPath  string `json:"saw_path"`
			ShowLogs bool   `json:"show_logs"`
		}{}

		if err := re.BindBody(&data); err != nil {
			return re.BadRequestError("Invalid request body", err)
		}

		if data.ServerID == "" {
			return re.BadRequestError("server_id is required", nil)
		}

		sawPath := data.SAWPath
		if sawPath == "" {
			sawPath = p.config.DefaultSAWPath
		}
		if sawPath == "" {
			return re.BadRequestError("saw_path is required", nil)
		}

		configs, err := p.LoadSAWConfigs(sawPath)
		if err != nil {
			return re.BadRequestError("Failed to load SAW configs", err)
		}

		config, ok := configs[data.ServerID]
		if !ok {
			return re.NotFoundError("Server ID not found", nil)
		}

		if err := p.StartServer(data.ServerID, config, sawPath, data.ShowLogs); err != nil {
			return re.InternalServerError("Failed to start server", err)
		}

		return re.JSON(200, map[string]any{
			"success": true,
			"message": "Server started successfully",
		})
	})

	// POST /api/server/stop - Stop a server
	e.Router.POST("/api/server/stop", func(re *core.RequestEvent) error {
		data := struct {
			ServerID string `json:"server_id"`
		}{}

		if err := re.BindBody(&data); err != nil {
			return re.BadRequestError("Invalid request body", err)
		}

		if data.ServerID == "" {
			return re.BadRequestError("server_id is required", nil)
		}

		if err := p.StopServer(data.ServerID); err != nil {
			return re.InternalServerError("Failed to stop server", err)
		}

		return re.JSON(200, map[string]any{
			"success": true,
			"message": "Server stopped successfully",
		})
	})

	// GET /api/server/status - Get status of all managed servers
	e.Router.GET("/api/server/status", func(re *core.RequestEvent) error {
		servers := p.ListServers()
		return re.JSON(200, map[string]any{
			"servers": servers,
		})
	})

	// GET /api/server/list - List available servers from SAW
	e.Router.GET("/api/server/list", func(re *core.RequestEvent) error {
		sawPath := re.Request.URL.Query().Get("saw_path")
		if sawPath == "" {
			sawPath = p.config.DefaultSAWPath
		}
		if sawPath == "" {
			return re.BadRequestError("saw_path query parameter is required", nil)
		}

		configs, err := p.LoadSAWConfigs(sawPath)
		if err != nil {
			return re.BadRequestError("Failed to load SAW configs", err)
		}

		serverList := make([]map[string]any, 0, len(configs))
		for id, cfg := range configs {
			serverList = append(serverList, map[string]any{
				"id":          id,
				"name":        cfg.ServerHostname,
				"map":         cfg.ServerDefaultMap,
				"mode":        cfg.ServerScenarioMode,
				"port":        cfg.ServerGamePort,
				"query_port":  cfg.ServerQueryPort,
				"max_players": cfg.ServerMaxPlayers,
			})
		}

		return re.JSON(200, map[string]any{
			"servers": serverList,
		})
	})

	return e.Next()
}

// LoadSAWConfigs loads server configurations from SAW installation
func (p *Plugin) LoadSAWConfigs(sawPath string) (map[string]SAWServerConfig, error) {
	sawPath = strings.ReplaceAll(sawPath, "\\", "/")

	configPath := filepath.Join(sawPath, "admin-interface", "config", "server-configs.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read server configs: %w", err)
	}

	var configs map[string]SAWServerConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse server configs: %w", err)
	}

	return configs, nil
}

// StartServer starts an Insurgency server
func (p *Plugin) StartServer(serverID string, config SAWServerConfig, sawPath string, showLogs bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if server, exists := p.servers[serverID]; exists && server.IsRunning {
		return fmt.Errorf("server %s is already running", serverID)
	}

	sawPath = strings.ReplaceAll(sawPath, "\\", "/")

	serverExe := os.Getenv("INSURGENCY_SERVER_PATH")
	if serverExe == "" {
		serverExe = filepath.Join(sawPath, "sandstorm-server", "Insurgency", "Binaries", "Win64", "InsurgencyServer-Win64-Shipping.exe")
	}

	absServerExe, err := filepath.Abs(serverExe)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for server executable: %w", err)
	}
	serverExe = absServerExe

	if _, err := os.Stat(serverExe); os.IsNotExist(err) {
		return fmt.Errorf("server executable not found at: %s", serverExe)
	}

	scenarioName := fmt.Sprintf("Scenario_%s_%s", config.ServerDefaultMap, config.ServerScenarioMode)
	travelArgs := config.ServerDefaultMap + "?Scenario=" + scenarioName

	if config.ServerMaxPlayers != "" {
		travelArgs += "?MaxPlayers=" + config.ServerMaxPlayers
	}

	if config.ServerLightingDay == "true" {
		travelArgs += "?Lighting=Day"
	} else {
		travelArgs += "?Lighting=Night"
	}

	if config.ServerCustomTravelArgs != "" {
		travelArgs += "?" + config.ServerCustomTravelArgs
	}

	args := []string{
		travelArgs,
		"-Hostname=" + config.ServerHostname,
		"-MaxPlayers=" + config.ServerMaxPlayers,
		"-Port=" + config.ServerGamePort,
		"-QueryPort=" + config.ServerQueryPort,
		"-LogCmds=LogGameplayEvents Log",
		"-LOCALLOGTIMES",
		"-AdminList=Admins",
		"-MapCycle=MapCycle",
	}

	if showLogs {
		args = append(args, "-stdout")
	} else {
		args = append(args, "-log="+serverID+".log")
	}

	if config.ServerPassword != "" {
		args = append(args, "-Password="+config.ServerPassword)
	}

	if len(config.ServerMutators) > 0 {
		mutators := strings.Join(config.ServerMutators, ",")
		args = append(args, "-Mutators="+mutators)
	}
	if config.ServerMutatorsCustom != "" {
		args = append(args, "-Mutators="+config.ServerMutatorsCustom)
	}

	if config.ServerCheats == "true" {
		args = append(args, "-CmdServerCheats")
	}

	if config.ServerCustomServerArgs != "" {
		customArgs := strings.Fields(config.ServerCustomServerArgs)
		args = append(args, customArgs...)
	}

	absSAWPath, err := filepath.Abs(sawPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for SAW: %w", err)
	}

	p.app.Logger().Info("Starting Insurgency server",
		"serverID", serverID,
		"name", config.ServerHostname,
		"executable", serverExe,
		"workDir", absSAWPath,
	)

	cmd := exec.Command(serverExe, args...)
	cmd.Dir = absSAWPath

	if showLogs {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	p.servers[serverID] = &ManagedServer{
		ID:        serverID,
		Config:    config,
		SAWPath:   sawPath,
		Cmd:       cmd,
		IsRunning: true,
	}

	go p.monitorServer(serverID, cmd)

	return nil
}

// monitorServer monitors a server process and updates status when it exits
func (p *Plugin) monitorServer(serverID string, cmd *exec.Cmd) {
	err := cmd.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()

	if server, exists := p.servers[serverID]; exists {
		server.IsRunning = false
		if err != nil {
			p.app.Logger().Error("Server exited with error", "serverID", serverID, "error", err)
		} else {
			p.app.Logger().Info("Server stopped", "serverID", serverID)
		}
	}
}

// StopServer stops a running server
func (p *Plugin) StopServer(serverID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	server, exists := p.servers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if !server.IsRunning {
		return fmt.Errorf("server %s is not running", serverID)
	}

	if server.Cmd == nil || server.Cmd.Process == nil {
		return fmt.Errorf("server %s has no process", serverID)
	}

	p.app.Logger().Info("Stopping server", "serverID", serverID)

	if err := server.Cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill server process: %w", err)
	}

	server.IsRunning = false
	return nil
}

// GetServerStatus returns the status of a server
func (p *Plugin) GetServerStatus(serverID string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	server, exists := p.servers[serverID]
	if !exists {
		return false, fmt.Errorf("server %s not found", serverID)
	}

	return server.IsRunning, nil
}

// ListServers returns all managed servers
func (p *Plugin) ListServers() map[string]bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := make(map[string]bool)
	for id, server := range p.servers {
		status[id] = server.IsRunning
	}
	return status
}

// StopAll stops all running servers
func (p *Plugin) StopAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, server := range p.servers {
		if server.IsRunning && server.Cmd != nil && server.Cmd.Process != nil {
			p.app.Logger().Info("Stopping server", "serverID", id)
			server.Cmd.Process.Kill()
			server.IsRunning = false
		}
	}
}

// GetPlugin returns the plugin instance from a PocketBase app
func GetPlugin(app *pocketbase.PocketBase) *Plugin {
	// This would require storing the plugin instance in app context
	// For now, we'll implement it when needed
	return nil
}
