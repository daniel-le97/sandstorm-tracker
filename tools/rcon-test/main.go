package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"sandstorm-tracker/internal/rcon"
)

func main() {
	address := flag.String("address", "127.0.0.1:27015", "RCON server address")
	password := flag.String("password", "", "RCON password")
	command := flag.String("command", "listplayers", "RCON command to execute")
	flag.Parse()

	if *password == "" {
		log.Fatal("Password is required. Use -password flag")
	}

	fmt.Printf("Connecting to %s...\n", *address)

	conn, err := net.DialTimeout("tcp", *address, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	config := rcon.DefaultConfig()
	client := rcon.NewRconClient(conn, config)

	fmt.Println("Authenticating...")
	if !client.Auth(*password) {
		log.Fatal("Authentication failed")
	}

	fmt.Printf("\nExecuting command: %s\n", *command)
	response, err := client.Send(*command)
	if err != nil {
		log.Fatalf("Failed to send command: %v", err)
	}

	fmt.Println("\n=== Response ===")
	fmt.Println(response)

	// Parse listplayers response if that's what we ran
	if *command == "listplayers" {
		parseListPlayersResponse(response)
	}
}

func parseListPlayersResponse(response string) {
	lines := strings.Split(response, "\n")

	fmt.Println("\n=== Parsed Players ===")
	playerCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: [ID] PlayerName - Score: X, Time: Y
		// or similar variations
		if strings.Contains(line, "Score:") || strings.Contains(line, "Kills:") {
			playerCount++
			fmt.Printf("  %s\n", line)
		}
	}

	fmt.Printf("\nTotal players found: %d\n", playerCount)
}
