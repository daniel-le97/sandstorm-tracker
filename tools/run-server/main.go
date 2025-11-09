package main

import (
	"encoding/json"
	"fmt"
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
	fmt.Println("Starting server... (Press Ctrl+C to stop)")
	fmt.Println()

	// Run the server with working directory set to SAW path
	cmd := exec.Command(serverExe, args...)
	cmd.Dir = absSAWPath // Set working directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}

	fmt.Println("\nServer stopped.")
}
