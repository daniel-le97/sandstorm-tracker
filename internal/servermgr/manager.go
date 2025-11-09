package servermgr

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// SAWServerConfig represents a single server configuration from SAW's server-configs.json
type SAWServerConfig struct {
	ID                       string   `json:"id"`
	ServerConfigName         string   `json:"server-config-name"`
	ServerDefaultMap         string   `json:"server_default_map"`
	ServerLightingDay        string   `json:"server_lighting_day"`
	ServerDefaultSide        string   `json:"server_default_side"`
	ServerMaxPlayers         string   `json:"server_max_players"`
	ServerMaxPlayersOverride string   `json:"server_max_players_override"`
	ServerGameMode           string   `json:"server_game_mode"`
	ServerScenarioMode       string   `json:"server_scenario_mode"`
	ServerRuleSet            string   `json:"server_rule_set"`
	ServerMutatorsCustom     string   `json:"server_mutators_custom"`
	ServerMutators           []string `json:"server_mutators"`
	ServerCheats             string   `json:"server_cheats"`
	ServerHostname           string   `json:"server_hostname"`
	ServerPassword           string   `json:"server_password"`
	ServerGamePort           string   `json:"server_game_port"`
	ServerQueryPort          string   `json:"server_query_port"`
	ServerRconEnabled        string   `json:"server_rcon_enabled"`
	ServerRconPort           string   `json:"server_rcon_port"`
	ServerRconPassword       string   `json:"server_rcon_password"`
	ServerCustomServerArgs   string   `json:"server_custom_server_args"`
	ServerCustomTravelArgs   string   `json:"server_custom_travel_args"`
}

// ManagedServer represents a running server instance
type ManagedServer struct {
	ID        string
	Config    SAWServerConfig
	SAWPath   string
	Cmd       *exec.Cmd
	IsRunning bool
}

// ServerManager manages Insurgency server processes
type ServerManager struct {
	mu      sync.RWMutex
	servers map[string]*ManagedServer
	logger  *slog.Logger
}

// NewServerManager creates a new server manager
func NewServerManager(logger *slog.Logger) *ServerManager {
	return &ServerManager{
		servers: make(map[string]*ManagedServer),
		logger:  logger,
	}
}

// LoadSAWConfigs loads server configurations from SAW installation
func (sm *ServerManager) LoadSAWConfigs(sawPath string) (map[string]SAWServerConfig, error) {
	// Normalize path
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
func (sm *ServerManager) StartServer(serverID string, config SAWServerConfig, sawPath string, showLogs bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if already running
	if server, exists := sm.servers[serverID]; exists && server.IsRunning {
		return fmt.Errorf("server %s is already running", serverID)
	}

	// Normalize path
	sawPath = strings.ReplaceAll(sawPath, "\\", "/")

	// Get server executable path from env var or construct from SAW path
	serverExe := os.Getenv("INSURGENCY_SERVER_PATH")
	if serverExe == "" {
		serverExe = filepath.Join(sawPath, "sandstorm-server", "Insurgency", "Binaries", "Win64", "InsurgencyServer-Win64-Shipping.exe")
	}

	// Make server executable path absolute
	absServerExe, err := filepath.Abs(serverExe)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for server executable: %w", err)
	}
	serverExe = absServerExe

	// Check if server executable exists
	if _, err := os.Stat(serverExe); os.IsNotExist(err) {
		return fmt.Errorf("server executable not found at: %s", serverExe)
	}

	// Build the map/scenario travel string
	scenarioName := fmt.Sprintf("Scenario_%s_%s", config.ServerDefaultMap, config.ServerScenarioMode)
	travelArgs := config.ServerDefaultMap + "?Scenario=" + scenarioName

	if config.ServerMaxPlayers != "" {
		travelArgs += "?MaxPlayers=" + config.ServerMaxPlayers
	}

	// Add lighting
	if config.ServerLightingDay == "true" {
		travelArgs += "?Lighting=Day"
	} else {
		travelArgs += "?Lighting=Night"
	}

	// Add custom travel args
	if config.ServerCustomTravelArgs != "" {
		travelArgs += "?" + config.ServerCustomTravelArgs
	}

	// Build server arguments
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

	// Add log output destination
	if showLogs {
		args = append(args, "-stdout") // Output to console
	} else {
		args = append(args, "-log="+serverID+".log") // Output to file
	}

	// Add password if set
	if config.ServerPassword != "" {
		args = append(args, "-Password="+config.ServerPassword)
	}

	// Add mutators
	if len(config.ServerMutators) > 0 {
		mutators := strings.Join(config.ServerMutators, ",")
		args = append(args, "-Mutators="+mutators)
	}
	if config.ServerMutatorsCustom != "" {
		args = append(args, "-Mutators="+config.ServerMutatorsCustom)
	}

	// Add cheats
	if config.ServerCheats == "true" {
		args = append(args, "-CmdServerCheats")
	}

	// Add custom server args
	if config.ServerCustomServerArgs != "" {
		customArgs := strings.Fields(config.ServerCustomServerArgs)
		args = append(args, customArgs...)
	}

	// Get absolute path for SAW directory
	absSAWPath, err := filepath.Abs(sawPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for SAW: %w", err)
	}

	sm.logger.Info("Starting Insurgency server",
		"serverID", serverID,
		"name", config.ServerHostname,
		"executable", serverExe,
		"workDir", absSAWPath,
	)

	// Create command
	cmd := exec.Command(serverExe, args...)
	cmd.Dir = absSAWPath

	// If showing logs, pipe to stdout/stderr
	if showLogs {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Store managed server
	sm.servers[serverID] = &ManagedServer{
		ID:        serverID,
		Config:    config,
		SAWPath:   sawPath,
		Cmd:       cmd,
		IsRunning: true,
	}

	// Monitor server process in background
	go sm.monitorServer(serverID, cmd)

	return nil
}

// monitorServer monitors a server process and updates status when it exits
func (sm *ServerManager) monitorServer(serverID string, cmd *exec.Cmd) {
	err := cmd.Wait()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if server, exists := sm.servers[serverID]; exists {
		server.IsRunning = false
		if err != nil {
			sm.logger.Error("Server exited with error", "serverID", serverID, "error", err)
		} else {
			sm.logger.Info("Server stopped", "serverID", serverID)
		}
	}
}

// StopServer stops a running server
func (sm *ServerManager) StopServer(serverID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	server, exists := sm.servers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if !server.IsRunning {
		return fmt.Errorf("server %s is not running", serverID)
	}

	if server.Cmd == nil || server.Cmd.Process == nil {
		return fmt.Errorf("server %s has no process", serverID)
	}

	sm.logger.Info("Stopping server", "serverID", serverID)

	// Kill the process
	if err := server.Cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill server process: %w", err)
	}

	server.IsRunning = false
	return nil
}

// GetServerStatus returns the status of a server
func (sm *ServerManager) GetServerStatus(serverID string) (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	server, exists := sm.servers[serverID]
	if !exists {
		return false, fmt.Errorf("server %s not found", serverID)
	}

	return server.IsRunning, nil
}

// ListServers returns all managed servers
func (sm *ServerManager) ListServers() map[string]bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	status := make(map[string]bool)
	for id, server := range sm.servers {
		status[id] = server.IsRunning
	}
	return status
}

// StopAll stops all running servers
func (sm *ServerManager) StopAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, server := range sm.servers {
		if server.IsRunning && server.Cmd != nil && server.Cmd.Process != nil {
			sm.logger.Info("Stopping server", "serverID", id)
			server.Cmd.Process.Kill()
			server.IsRunning = false
		}
	}
}
