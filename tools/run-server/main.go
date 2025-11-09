package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: run-server <path-to-saw-root> [server-id]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  run-server.exe C:\\sandstorm-admin-wrapper")
		fmt.Println("  run-server.exe C:\\sandstorm-admin-wrapper 1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde")
		os.Exit(1)
	}

	sawPath := os.Args[1]
	var serverID string
	if len(os.Args) > 2 {
		serverID = os.Args[2]
	}

	// Load server configs
	configPath := filepath.Join(sawPath, "admin-interface", "config", "server-configs.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read server configs: %v", err)
	}

	var configs map[string]SAWServerConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		log.Fatalf("Failed to parse server configs: %v", err)
	}

	// If no server ID specified, use the first one
	var config SAWServerConfig
	if serverID == "" {
		for id, cfg := range configs {
			serverID = id
			config = cfg
			break
		}
	} else {
		var ok bool
		config, ok = configs[serverID]
		if !ok {
			log.Fatalf("Server ID %s not found in configs", serverID)
		}
	}

	fmt.Printf("Starting Insurgency: Sandstorm Server\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Server ID:   %s\n", serverID)
	fmt.Printf("Name:        %s\n", config.ServerHostname)
	fmt.Printf("Map:         %s\n", config.ServerDefaultMap)
	fmt.Printf("Mode:        %s\n", config.ServerScenarioMode)
	fmt.Printf("Game Port:   %s\n", config.ServerGamePort)
	fmt.Printf("Query Port:  %s\n", config.ServerQueryPort)
	fmt.Printf("RCON Port:   %s\n", config.ServerRconPort)
	fmt.Printf("Max Players: %s\n", config.ServerMaxPlayers)
	fmt.Printf("=====================================\n\n")

	// Get server executable path from env var or construct from SAW path
	serverExe := os.Getenv("INSURGENCY_SERVER_PATH")
	if serverExe == "" {
		serverExe = filepath.Join(sawPath, "sandstorm-server", "Insurgency", "Binaries", "Win64", "InsurgencyServer-Win64-Shipping.exe")
	}

	// Make server executable path absolute
	absServerExe, err := filepath.Abs(serverExe)
	if err != nil {
		log.Fatalf("Failed to get absolute path for server executable: %v", err)
	}
	serverExe = absServerExe

	fmt.Printf("Server executable: %s\n\n", serverExe)

	// Check if server executable exists
	if _, err := os.Stat(serverExe); os.IsNotExist(err) {
		log.Fatalf("Server executable not found at: %s\n\nPlease set INSURGENCY_SERVER_PATH environment variable or ensure server is installed", serverExe)
	}

	// Build the map/scenario travel string
	// Format: MapName?Scenario=Scenario_MapName_Mode
	// SAW uses format like: Ministry?Scenario=Scenario_Ministry_Checkpoint_Security
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

	// Build server arguments (matching SAW format)
	args := []string{
		travelArgs,
		"-Hostname=" + config.ServerHostname,
		"-MaxPlayers=" + config.ServerMaxPlayers,
		"-Port=" + config.ServerGamePort,
		"-QueryPort=" + config.ServerQueryPort,
		"-log=" + serverID + ".log",
		"-LogCmds=LogGameplayEvents Log",
		"-LOCALLOGTIMES",
		"-AdminList=Admins",
		"-MapCycle=MapCycle",
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
		log.Fatalf("Failed to get absolute path for SAW: %v", err)
	}

	fmt.Printf("Working Directory: %s\n", absSAWPath)
	fmt.Printf("Command: %s %s\n\n", serverExe, strings.Join(args, " "))

	// Setup config management
	serverInstancePath := filepath.Join(absSAWPath, "servers", serverID)
	localConfigDir := filepath.Join(absSAWPath, "server-config", serverID)

	// Ensure local config directory exists
	if err := os.MkdirAll(localConfigDir, 0755); err != nil {
		log.Fatalf("Failed to create local config directory: %v", err)
	}

	// Copy config files from local storage to server instance before launch
	fmt.Println("Applying server configuration...")
	if err := applyServerConfig(serverInstancePath, localConfigDir); err != nil {
		log.Fatalf("Failed to apply server config: %v", err)
	}

	fmt.Println("Starting server... (Press Ctrl+C to stop)")
	fmt.Println()

	// Run the server with working directory set to SAW path
	cmd := exec.Command(serverExe, args...)
	cmd.Dir = absSAWPath // Set working directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Server exited with error: %v", err)
	}

	fmt.Println("\nServer stopped. Syncing bans...")

	// After server closes, sync ban list back to local storage
	if err := syncBansFromServer(serverInstancePath, localConfigDir); err != nil {
		log.Printf("Warning: Failed to sync bans: %v", err)
	}

	fmt.Println("Server shutdown complete.")
}

// applyServerConfig copies config files from local storage to server instance
// This prevents the server from overwriting our manual changes
func applyServerConfig(serverInstancePath, localConfigDir string) error {
	// Config files to manage (relative to Saved/Config/LinuxServer or Saved/Config/WindowsServer)
	configFiles := []string{
		"Game.ini",
		"Engine.ini",
		"Admins.txt",
		"Bans.txt",
		"MapCycle.txt",
		"Motd.txt",
	}

	// Determine if this is Windows or Linux
	savedPath := filepath.Join(serverInstancePath, "Saved")
	configBasePath := ""

	// Check for WindowsServer first, then LinuxServer
	windowsConfig := filepath.Join(savedPath, "Config", "WindowsServer")
	linuxConfig := filepath.Join(savedPath, "Config", "LinuxServer")

	if _, err := os.Stat(windowsConfig); err == nil {
		configBasePath = windowsConfig
	} else if _, err := os.Stat(linuxConfig); err == nil {
		configBasePath = linuxConfig
	} else {
		// Create WindowsServer directory if neither exists
		configBasePath = windowsConfig
		if err := os.MkdirAll(configBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Copy each config file from local storage to server instance
	for _, filename := range configFiles {
		localFile := filepath.Join(localConfigDir, filename)
		serverFile := filepath.Join(configBasePath, filename)

		// If local file doesn't exist, create an empty one
		if _, err := os.Stat(localFile); os.IsNotExist(err) {
			// Create empty file
			if err := os.WriteFile(localFile, []byte{}, 0644); err != nil {
				log.Printf("Warning: Failed to create %s: %v", filename, err)
				continue
			}
		}

		// Copy from local to server
		if err := copyFile(localFile, serverFile); err != nil {
			log.Printf("Warning: Failed to copy %s: %v", filename, err)
		} else {
			fmt.Printf("  ✓ Applied %s\n", filename)
		}
	}

	return nil
}

// syncBansFromServer copies updated ban list from server instance back to local storage
func syncBansFromServer(serverInstancePath, localConfigDir string) error {
	savedPath := filepath.Join(serverInstancePath, "Saved")

	// Determine config path
	windowsConfig := filepath.Join(savedPath, "Config", "WindowsServer")
	linuxConfig := filepath.Join(savedPath, "Config", "LinuxServer")

	var configBasePath string
	if _, err := os.Stat(windowsConfig); err == nil {
		configBasePath = windowsConfig
	} else if _, err := os.Stat(linuxConfig); err == nil {
		configBasePath = linuxConfig
	} else {
		return fmt.Errorf("config directory not found")
	}

	serverBansFile := filepath.Join(configBasePath, "Bans.txt")
	localBansFile := filepath.Join(localConfigDir, "Bans.txt")

	// Copy bans from server back to local storage
	if _, err := os.Stat(serverBansFile); err == nil {
		if err := copyFile(serverBansFile, localBansFile); err != nil {
			return fmt.Errorf("failed to sync bans: %w", err)
		}
		fmt.Println("  ✓ Synced ban list")
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
