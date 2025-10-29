package a2s

import (
	"fmt"
	"log"
	"time"
)

// Example demonstrates how to use the A2S client
func Example() {
	// Create a new client with default timeout
	client := NewClient()

	// Query server information
	serverAddress := "yourserver.com:27102" // Query port, not game port

	info, err := client.QueryInfo(serverAddress)
	if err != nil {
		log.Printf("Failed to query server info: %v", err)
		return
	}

	// Display server information
	fmt.Printf("=== Server Info ===\n")
	fmt.Printf("Name: %s\n", info.Name)
	fmt.Printf("Map: %s\n", info.Map)
	fmt.Printf("Game: %s\n", info.Game)
	fmt.Printf("Players: %d/%d (Bots: %d)\n", info.Players, info.MaxPlayers, info.Bots)
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("VAC Secured: %v\n", info.VAC == 1)

	// Check optional fields
	if info.Keywords != nil {
		fmt.Printf("Keywords: %s\n", *info.Keywords)
	}
	if info.SteamID != nil {
		fmt.Printf("Steam ID: %d\n", *info.SteamID)
	}

	// Query players
	players, err := client.QueryPlayers(serverAddress)
	if err != nil {
		log.Printf("Failed to query players: %v", err)
		return
	}

	fmt.Printf("\n=== Players (%d) ===\n", len(players))
	for _, player := range players {
		minutes := int(player.Duration / 60)
		seconds := int(player.Duration) % 60
		fmt.Printf("  %s - Score: %d, Time: %dm%ds\n",
			player.Name, player.Score, minutes, seconds)
	}
}

// ExampleMonitoring demonstrates continuous server monitoring
func ExampleMonitoring() {
	client := NewClient()
	serverAddress := "yourserver.com:27102"

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		info, err := client.QueryInfo(serverAddress)
		if err != nil {
			log.Printf("Server query failed: %v", err)
			continue
		}

		log.Printf("Server: %s | Map: %s | Players: %d/%d",
			info.Name, info.Map, info.Players, info.MaxPlayers)

		// Alert if server is full
		if info.Players >= info.MaxPlayers {
			log.Printf("⚠️  Server is full!")
		}

		// Alert if server is empty
		if info.Players == 0 {
			log.Printf("ℹ️  Server is empty")
		}
	}
}

// ExampleIntegrationWithTracker shows how to integrate A2S with the tracker
func ExampleIntegrationWithTracker() {
	client := NewClient()
	serverAddress := "yourserver.com:27102"

	// Get current server state
	info, err := client.QueryInfo(serverAddress)
	if err != nil {
		log.Printf("Failed to query server: %v", err)
		return
	}

	// Use this info to:
	// 1. Verify server is online before processing logs
	// 2. Update server metadata in database
	// 3. Track player count trends
	// 4. Monitor server health

	fmt.Printf("Processing logs for: %s (Map: %s, Players: %d/%d)\n",
		info.Name, info.Map, info.Players, info.MaxPlayers)

	// Get player list to cross-reference with log data
	players, err := client.QueryPlayers(serverAddress)
	if err != nil {
		log.Printf("Failed to get player list: %v", err)
		return
	}

	// Compare log-based player list with live player list
	fmt.Printf("Currently online: %d players\n", len(players))
	for _, p := range players {
		fmt.Printf("  - %s (Score: %d)\n", p.Name, p.Score)
	}
}

// ExampleWithRetries shows how to implement retry logic
func ExampleWithRetries() {
	client := NewClientWithTimeout(3 * time.Second)
	serverAddress := "yourserver.com:27102"

	maxRetries := 3
	retryDelay := time.Second

	var info *ServerInfo
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		info, err = client.QueryInfo(serverAddress)
		if err == nil {
			break
		}

		log.Printf("Attempt %d/%d failed: %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		log.Fatalf("Failed after %d attempts: %v", maxRetries, err)
	}

	fmt.Printf("Successfully queried: %s\n", info.Name)
}

// ExampleBatchQuery demonstrates querying multiple servers
func ExampleBatchQuery() {
	client := NewClient()

	servers := []string{
		"server1.com:27102",
		"server2.com:27102",
		"server3.com:27102",
	}

	type ServerStatus struct {
		Address string
		Info    *ServerInfo
		Error   error
	}

	results := make(chan ServerStatus, len(servers))

	// Query servers concurrently
	for _, addr := range servers {
		go func(address string) {
			info, err := client.QueryInfo(address)
			results <- ServerStatus{
				Address: address,
				Info:    info,
				Error:   err,
			}
		}(addr)
	}

	// Collect results
	fmt.Println("=== Server Status ===")
	for i := 0; i < len(servers); i++ {
		status := <-results

		if status.Error != nil {
			fmt.Printf("❌ %s - OFFLINE (%v)\n", status.Address, status.Error)
			continue
		}

		fmt.Printf("✅ %s - %s | %d/%d players | Map: %s\n",
			status.Address,
			status.Info.Name,
			status.Info.Players,
			status.Info.MaxPlayers,
			status.Info.Map)
	}
}
