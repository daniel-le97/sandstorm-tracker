package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
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
	mu             sync.RWMutex
	servers        map[string]*ManagedServer
	logger         *slog.Logger
	defaultSAWPath string
}

// ProcessInfo holds information about a running process
type ProcessInfo struct {
	Pid  int
	Name string
}

func main() {
	godotenv.Load() // Load .env file if present
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	sm := &ServerManager{
		servers: make(map[string]*ManagedServer),
		logger:  logger,
	}

	rootCmd := &cobra.Command{
		Use:   "servermgr",
		Short: "Insurgency: Sandstorm Server Manager",
		Long:  "A standalone tool for managing Insurgency: Sandstorm dedicated server instances",
	}

	// Set default SAW path from environment or flag
	rootCmd.PersistentFlags().StringVar(&sm.defaultSAWPath, "saw-path", os.Getenv("SAW_PATH"), "Path to Sandstorm Admin Wrapper installation")

	sm.registerCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// registerCommands adds server management CLI commands
func (sm *ServerManager) registerCommands(rootCmd *cobra.Command) {
	// Start command
	startCmd := &cobra.Command{
		Use:   "start [server-id]",
		Short: "Start an Insurgency server",
		Long:  "Start an Insurgency server from SAW configuration. Use --all to start all servers, or provide a server ID to start a specific server.",
		RunE:  sm.startCommand,
	}
	startCmd.Flags().Bool("logs", false, "Show server logs in console (default: log to file)")
	startCmd.Flags().Bool("all", false, "Start all servers from SAW configuration")

	// Stop command
	stopCmd := &cobra.Command{
		Use:   "stop [server-id]",
		Short: "Stop a running Insurgency server",
		Long:  "Stop a running Insurgency server. Use --all to stop all servers, or provide a server ID to stop a specific server.",
		RunE:  sm.stopCommand,
	}
	stopCmd.Flags().Bool("all", false, "Stop all running servers")

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check server status",
		RunE:  sm.statusCommand,
	}

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available servers from SAW configuration",
		RunE:  sm.listCommand,
	}

	// Update SteamCMD command
	updateSteamCmdCmd := &cobra.Command{
		Use:   "update-steamcmd",
		Short: "Update SteamCMD to the latest version",
		RunE:  sm.updateSteamCmdCommand,
	}

	// Update game command
	updateGameCmd := &cobra.Command{
		Use:   "update-game",
		Short: "Update Insurgency: Sandstorm server to the latest version",
		Long:  "Update the Insurgency: Sandstorm dedicated server using SteamCMD. This will validate and update all server files.",
		RunE:  sm.updateGameCommand,
	}
	updateGameCmd.Flags().Bool("validate", false, "Validate all server files (slower but more thorough)")
	updateGameCmd.Flags().Bool("force", false, "Force update even if servers are running (not recommended)")

	rootCmd.AddCommand(startCmd, stopCmd, statusCmd, listCmd, updateSteamCmdCmd, updateGameCmd)
}

// startCommand handles the start command
func (sm *ServerManager) startCommand(cmd *cobra.Command, args []string) error {
	showLogs, _ := cmd.Flags().GetBool("logs")
	startAll, _ := cmd.Flags().GetBool("all")
	sawPath := sm.getSAWPath()

	if sawPath == "" {
		return fmt.Errorf("SAW path not provided. Use --saw-path flag or set SAW_PATH environment variable")
	}

	configs, err := sm.loadSAWConfigs(sawPath)
	if err != nil {
		return fmt.Errorf("failed to load SAW configs: %w", err)
	}

	if len(configs) == 0 {
		fmt.Println("No servers found in SAW configuration")
		return nil
	}

	// Start all servers if --all flag is set
	if startAll {
		fmt.Printf("Starting %d server(s)...\n", len(configs))
		successCount := 0
		failCount := 0

		for serverID, serverConfig := range configs {
			fmt.Printf("  Starting %s (%s)... ", serverID, serverConfig.ServerHostname)
			if err := sm.startServer(serverID, serverConfig, sawPath, false); err != nil {
				fmt.Printf("FAILED: %v\n", err)
				failCount++
			} else {
				fmt.Println("OK")
				successCount++
			}
		}

		fmt.Printf("\nStarted %d/%d servers successfully\n", successCount, len(configs))
		if failCount > 0 {
			return fmt.Errorf("%d server(s) failed to start", failCount)
		}
		return nil
	}

	// Start single server
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
	} else {
		fmt.Printf("Server logs will be written to: %s.log\n", serverID)
	}

	return sm.startServer(serverID, serverConfig, sawPath, showLogs)
}

// stopCommand handles the stop command
func (sm *ServerManager) stopCommand(cmd *cobra.Command, args []string) error {
	stopAll, _ := cmd.Flags().GetBool("all")
	sawPath := sm.getSAWPath()

	if sawPath == "" {
		return fmt.Errorf("SAW path not provided. Use --saw-path flag or set SAW_PATH environment variable")
	}

	// Stop all servers if --all flag is set
	if stopAll {
		configs, err := sm.loadSAWConfigs(sawPath)
		if err != nil {
			return fmt.Errorf("failed to load SAW configs: %w", err)
		}

		if len(configs) == 0 {
			fmt.Println("No servers found in SAW configuration")
			return nil
		}

		fmt.Printf("Stopping all servers...\n")
		successCount := 0
		failCount := 0
		staleCount := 0

		for serverID := range configs {
			// Check if PID file exists (server might not be running)
			pid, err := sm.loadPIDFile(serverID)
			if err != nil {
				// No PID file, skip
				continue
			}

			// Check if the process is actually running
			if !sm.isProcessRunning(pid) {
				// Clean up stale PID file
				sm.logger.Info("Cleaning up stale PID file", "serverID", serverID, "pid", pid)
				if err := sm.removePIDFile(serverID); err != nil {
					sm.logger.Warn("Failed to remove stale PID file", "error", err)
				}
				staleCount++
				continue
			}

			fmt.Printf("  Stopping %s... ", serverID)
			if err := sm.stopServer(serverID, sawPath); err != nil {
				fmt.Printf("FAILED: %v\n", err)
				failCount++
			} else {
				fmt.Println("OK")
				successCount++
			}
		}

		if staleCount > 0 {
			fmt.Printf("Cleaned up %d stale PID file(s)\n", staleCount)
		}
		fmt.Printf("\nStopped %d server(s) successfully\n", successCount)
		if failCount > 0 {
			return fmt.Errorf("%d server(s) failed to stop", failCount)
		}
		return nil
	}

	// Stop single server
	if len(args) == 0 {
		return fmt.Errorf("server ID is required (or use --all to stop all servers)")
	}

	serverID := args[0]
	fmt.Printf("Stopping server: %s\n", serverID)

	if err := sm.stopServer(serverID, sawPath); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	fmt.Println("Server stopped successfully")
	return nil
}

// statusCommand handles the status command
func (sm *ServerManager) statusCommand(cmd *cobra.Command, args []string) error {
	sawPath := sm.getSAWPath()

	// Check actual running processes instead of in-memory state
	procs, err := sm.getRunningServerProcesses()
	if err != nil {
		return fmt.Errorf("failed to check running processes: %w", err)
	}

	// Load SAW configs to check for stale PID files
	configs, _ := sm.loadSAWConfigs(sawPath)

	// Check for stale PID files
	var stalePIDs []string
	for sid := range configs {
		if pid, err := sm.loadPIDFile(sid); err == nil {
			// PID file exists, check if process is running
			isRunning := false
			for _, proc := range procs {
				if proc.Pid == pid {
					isRunning = true
					break
				}
			}
			if !isRunning {
				stalePIDs = append(stalePIDs, fmt.Sprintf("%s (PID: %d)", sid, pid))
			}
		}
	}

	if len(procs) == 0 && len(stalePIDs) == 0 {
		fmt.Println("No Insurgency server processes are currently running")
		return nil
	}

	if len(procs) > 0 {
		fmt.Printf("Running Insurgency Server Processes (%d):\n", len(procs))
		fmt.Println(strings.Repeat("-", 50))

		for _, proc := range procs {
			serverID := ""
			// Try to find which server this PID belongs to
			for sid := range configs {
				if pid, err := sm.loadPIDFile(sid); err == nil && pid == proc.Pid {
					serverID = fmt.Sprintf(" (Server: %s)", sid)
					break
				}
			}
			fmt.Printf("PID %-8d %s%s\n", proc.Pid, proc.Name, serverID)
		}
	}

	if len(stalePIDs) > 0 {
		if len(procs) > 0 {
			fmt.Println()
		}
		fmt.Printf("⚠️  Stale PID Files Detected (%d):\n", len(stalePIDs))
		fmt.Println(strings.Repeat("-", 50))
		for _, stale := range stalePIDs {
			fmt.Printf("  %s\n", stale)
		}
		fmt.Println("\nRun 'servermgr stop --all' to clean up stale PID files")
	}

	return nil
}

// listCommand handles the list command
func (sm *ServerManager) listCommand(cmd *cobra.Command, args []string) error {
	sawPath := sm.getSAWPath()

	if sawPath == "" {
		return fmt.Errorf("SAW path not provided. Use --saw-path flag or set SAW_PATH environment variable")
	}

	configs, err := sm.loadSAWConfigs(sawPath)
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
}

// updateSteamCmdCommand handles the update-steamcmd command
func (sm *ServerManager) updateSteamCmdCommand(cmd *cobra.Command, args []string) error {
	sawPath := sm.getSAWPath()

	if sawPath == "" {
		return fmt.Errorf("SAW path not provided. Use --saw-path flag or set SAW_PATH environment variable")
	}

	fmt.Println("Updating SteamCMD...")
	if err := sm.updateSteamCMD(sawPath); err != nil {
		return fmt.Errorf("failed to update SteamCMD: %w", err)
	}

	fmt.Println("SteamCMD updated successfully!")
	return nil
}

// updateGameCommand handles the update-game command
func (sm *ServerManager) updateGameCommand(cmd *cobra.Command, args []string) error {
	sawPath := sm.getSAWPath()

	if sawPath == "" {
		return fmt.Errorf("SAW path not provided. Use --saw-path flag or set SAW_PATH environment variable")
	}

	force, _ := cmd.Flags().GetBool("force")
	validate, _ := cmd.Flags().GetBool("validate")

	// Check if any servers are running
	procs, err := sm.getRunningServerProcesses()
	if err == nil && len(procs) > 0 && !force {
		fmt.Printf("\n⚠️  WARNING: %d server process(es) are currently running!\n", len(procs))
		fmt.Println("Updating game files while servers are running can cause:")
		fmt.Println("  - Server crashes")
		fmt.Println("  - Data corruption")
		fmt.Println("  - Player disconnects")
		fmt.Println("\nPlease stop all servers first using: servermgr stop --all")
		fmt.Println("Or use --force to update anyway (not recommended)")
		return fmt.Errorf("servers are running - stop them first or use --force")
	}

	fmt.Println("Updating Insurgency: Sandstorm server...")
	if validate {
		fmt.Println("(This will validate all files - may take longer)")
	}

	if err := sm.updateInsurgencyServer(sawPath, validate); err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	fmt.Println("Insurgency: Sandstorm server updated successfully!")
	return nil
}

// getSAWPath returns the SAW path from flag or environment
func (sm *ServerManager) getSAWPath() string {
	if sm.defaultSAWPath != "" {
		return sm.defaultSAWPath
	}
	return os.Getenv("SAW_PATH")
}

// loadSAWConfigs loads server configurations from SAW installation
func (sm *ServerManager) loadSAWConfigs(sawPath string) (map[string]SAWServerConfig, error) {
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

// getPIDFilePath returns the path to the PID file for a server
func (sm *ServerManager) getPIDFilePath(serverID string) string {
	dataDir := "data"
	os.MkdirAll(dataDir, 0755)
	return filepath.Join(dataDir, fmt.Sprintf("%s.pid", serverID))
}

// savePIDFile saves the server's PID to a file
func (sm *ServerManager) savePIDFile(serverID string, pid int) error {
	pidFile := sm.getPIDFilePath(serverID)
	pidData := fmt.Sprintf("%d", pid)
	return os.WriteFile(pidFile, []byte(pidData), 0644)
}

// loadPIDFile loads the server's PID from a file
func (sm *ServerManager) loadPIDFile(serverID string) (int, error) {
	pidFile := sm.getPIDFilePath(serverID)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, fmt.Errorf("invalid PID file format: %w", err)
	}

	return pid, nil
}

// removePIDFile removes the PID file for a server
func (sm *ServerManager) removePIDFile(serverID string) error {
	pidFile := sm.getPIDFilePath(serverID)
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// isProcessRunning checks if a process with the given PID is currently running
func (sm *ServerManager) isProcessRunning(pid int) bool {
	psCmd := fmt.Sprintf("Get-Process -Id %d -ErrorAction SilentlyContinue | Select-Object Id", pid)
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// getRunningServerProcesses returns all running Insurgency server processes
func (sm *ServerManager) getRunningServerProcesses() ([]ProcessInfo, error) {
	var procs []ProcessInfo

	cmd := exec.Command("powershell", "-Command",
		"Get-Process -Name '*InsurgencyServer*' -ErrorAction SilentlyContinue | Select-Object Id, ProcessName | ConvertTo-Json")

	output, err := cmd.Output()
	if err != nil {
		return procs, nil
	}

	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return procs, nil
	}

	// Handle single object or array
	switch v := result.(type) {
	case map[string]interface{}:
		if id, ok := v["Id"].(float64); ok {
			if name, ok := v["ProcessName"].(string); ok {
				procs = append(procs, ProcessInfo{Pid: int(id), Name: name})
			}
		}
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if id, ok := m["Id"].(float64); ok {
					if name, ok := m["ProcessName"].(string); ok {
						procs = append(procs, ProcessInfo{Pid: int(id), Name: name})
					}
				}
			}
		}
	}

	return procs, nil
}

// startServer starts an Insurgency server
func (sm *ServerManager) startServer(serverID string, config SAWServerConfig, sawPath string, showLogs bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check for stale PID file and clean it up
	if pid, err := sm.loadPIDFile(serverID); err == nil {
		if !sm.isProcessRunning(pid) {
			sm.logger.Info("Cleaning up stale PID file before starting", "serverID", serverID, "pid", pid)
			if err := sm.removePIDFile(serverID); err != nil {
				sm.logger.Warn("Failed to remove stale PID file", "error", err)
			}
		} else {
			return fmt.Errorf("server %s is already running (PID: %d)", serverID, pid)
		}
	}

	if server, exists := sm.servers[serverID]; exists && server.IsRunning {
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

	// Build scenario name - for Checkpoint and Push modes, include the side
	scenarioName := fmt.Sprintf("Scenario_%s_%s", config.ServerDefaultMap, config.ServerScenarioMode)
	if config.ServerScenarioMode == "Checkpoint" || config.ServerScenarioMode == "Push" {
		scenarioName += "_" + config.ServerDefaultSide
	}
	travelArgs := config.ServerDefaultMap + "?Scenario=" + scenarioName

	if config.ServerMaxPlayers != "" {
		travelArgs += "?MaxPlayers=" + config.ServerMaxPlayers
	}

	// Add Game mode if specified
	if config.ServerGameMode != "" && config.ServerGameMode != "None" {
		travelArgs += "?Game=" + config.ServerGameMode
	}

	// Add password if specified
	if config.ServerPassword != "" {
		travelArgs += "?Password=" + config.ServerPassword
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

	// Apply server configuration before starting
	serverInstancePath := filepath.Join(absSAWPath, "sandstorm-server", "Insurgency")
	localConfigDir := filepath.Join(absSAWPath, "server-config", serverID)

	if err := sm.applyServerConfig(serverInstancePath, localConfigDir); err != nil {
		sm.logger.Warn("Failed to apply server config", "error", err)
	}

	sm.logger.Info("Starting Insurgency server",
		"serverID", serverID,
		"name", config.ServerHostname,
		"executable", serverExe,
		"workDir", absSAWPath,
	)

	// For servers without console logs, use PowerShell Start-Process to detach
	if !showLogs {
		argString := strings.Join(args, " ")

		psCmd := fmt.Sprintf(`
			$proc = Start-Process -FilePath '%s' -ArgumentList '%s' -WorkingDirectory '%s' -WindowStyle Hidden -PassThru
			Write-Output $proc.Id
		`, serverExe, argString, absSAWPath)

		cmd := exec.Command("powershell", "-Command", psCmd)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		pidStr := strings.TrimSpace(string(output))
		var pid int
		if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
			sm.logger.Warn("Failed to parse server PID", "output", pidStr)
		} else {
			if err := sm.savePIDFile(serverID, pid); err != nil {
				sm.logger.Warn("Failed to save PID file", "error", err)
			}
		}

		sm.logger.Info("Server started in detached mode", "pid", pid)
		return nil
	}

	// For console logs, use regular exec
	cmd := exec.Command(serverExe, args...)
	cmd.Dir = absSAWPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	sm.servers[serverID] = &ManagedServer{
		ID:        serverID,
		Config:    config,
		SAWPath:   sawPath,
		Cmd:       cmd,
		IsRunning: true,
	}

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

// stopServer stops a running server
func (sm *ServerManager) stopServer(serverID string, sawPath string) error {
	// First try to get PID from file
	pid, err := sm.loadPIDFile(serverID)
	if err == nil {
		if !sm.isProcessRunning(pid) {
			sm.logger.Info("Process not running, cleaning up stale PID file", "serverID", serverID, "pid", pid)
			if err := sm.removePIDFile(serverID); err != nil {
				sm.logger.Warn("Failed to remove stale PID file", "error", err)
			}
			return fmt.Errorf("server %s is not running (stale PID file cleaned up)", serverID)
		}

		sm.logger.Info("Stopping server via PID file", "serverID", serverID, "pid", pid)

		psCmd := fmt.Sprintf("Stop-Process -Id %d -Force -ErrorAction SilentlyContinue", pid)
		cmd := exec.Command("powershell", "-Command", psCmd)
		if err := cmd.Run(); err != nil {
			sm.logger.Warn("Failed to kill process via PID", "pid", pid, "error", err)
		}

		if err := sm.removePIDFile(serverID); err != nil {
			sm.logger.Warn("Failed to remove PID file", "error", err)
		}

		sm.mu.Lock()
		if server, exists := sm.servers[serverID]; exists {
			server.IsRunning = false
		}
		sm.mu.Unlock()

		return nil
	}

	// Fallback to in-memory tracking
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

	if err := server.Cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill server process: %w", err)
	}

	server.IsRunning = false
	return nil
}

// applyServerConfig copies configuration files to the server instance directory
func (sm *ServerManager) applyServerConfig(serverInstancePath, localConfigDir string) error {
	configFiles := []string{
		"Game.ini",
		"Engine.ini",
		"Admins.txt",
		"Bans.txt",
		"MapCycle.txt",
		"Motd.txt",
	}

	if err := os.MkdirAll(localConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create local config directory: %w", err)
	}

	savedPath := filepath.Join(serverInstancePath, "Saved")
	configBasePath := ""

	windowsConfig := filepath.Join(savedPath, "Config", "WindowsServer")
	linuxConfig := filepath.Join(savedPath, "Config", "LinuxServer")

	if _, err := os.Stat(windowsConfig); err == nil {
		configBasePath = windowsConfig
	} else if _, err := os.Stat(linuxConfig); err == nil {
		configBasePath = linuxConfig
	} else {
		configBasePath = windowsConfig
		if err := os.MkdirAll(configBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	for _, filename := range configFiles {
		localFile := filepath.Join(localConfigDir, filename)
		serverFile := filepath.Join(configBasePath, filename)

		if _, err := os.Stat(localFile); os.IsNotExist(err) {
			if err := os.WriteFile(localFile, []byte{}, 0644); err != nil {
				sm.logger.Warn("Failed to create config file", "file", filename, "error", err)
				continue
			}
		}

		if err := copyFile(localFile, serverFile); err != nil {
			sm.logger.Warn("Failed to copy config file", "file", filename, "error", err)
		} else {
			sm.logger.Debug("Applied config file", "file", filename)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, 0644)
}

// updateSteamCMD updates SteamCMD to the latest version
func (sm *ServerManager) updateSteamCMD(sawPath string) error {
	sawPath = strings.ReplaceAll(sawPath, "\\", "/")
	steamCmdPath := filepath.Join(sawPath, "steamcmd", "installation", "steamcmd.exe")

	if _, err := os.Stat(steamCmdPath); os.IsNotExist(err) {
		return fmt.Errorf("steamcmd.exe not found at: %s", steamCmdPath)
	}

	sm.logger.Info("Updating SteamCMD", "path", steamCmdPath)

	cmd := exec.Command(steamCmdPath, "+quit")
	cmd.Dir = filepath.Dir(steamCmdPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("steamcmd update failed: %w", err)
	}

	sm.logger.Info("SteamCMD updated successfully")
	return nil
}

// updateInsurgencyServer updates the Insurgency: Sandstorm dedicated server
func (sm *ServerManager) updateInsurgencyServer(sawPath string, validate bool) error {
	sawPath = strings.ReplaceAll(sawPath, "\\", "/")
	steamCmdPath := filepath.Join(sawPath, "steamcmd", "installation", "steamcmd.exe")
	serverPath := filepath.Join(sawPath, "sandstorm-server")

	if _, err := os.Stat(steamCmdPath); os.IsNotExist(err) {
		return fmt.Errorf("steamcmd.exe not found at: %s", steamCmdPath)
	}

	if err := os.MkdirAll(serverPath, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	sm.logger.Info("Updating Insurgency: Sandstorm server",
		"steamcmd", steamCmdPath,
		"serverPath", serverPath,
		"validate", validate,
	)

	args := []string{
		"+force_install_dir", serverPath,
		"+login", "anonymous",
		"+app_update", "581330",
	}

	if validate {
		args = append(args, "validate")
	}

	args = append(args, "+quit")

	cmd := exec.Command(steamCmdPath, args...)
	cmd.Dir = filepath.Dir(steamCmdPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("\nDownloading/Updating server files...")
	fmt.Println("This may take several minutes depending on your connection speed.")
	fmt.Println(strings.Repeat("-", 80))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("server update failed: %w", err)
	}

	sm.logger.Info("Insurgency: Sandstorm server updated successfully")
	return nil
}
