package a2s

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Example showing how to monitor 6 servers efficiently
func ExampleMultiServerMonitoring() {
	// Create a server pool
	pool := NewServerPool()

	// Add your 6 servers (replace with actual addresses)
	servers := map[string]string{
		"server1.example.com:27102": "US East Server",
		"server2.example.com:27102": "US West Server",
		"server3.example.com:27102": "EU Server",
		"server4.example.com:27102": "Asia Server",
		"server5.example.com:27102": "Test Server",
		"server6.example.com:27102": "Event Server",
	}

	for addr, name := range servers {
		pool.AddServer(addr, name)
	}

	// One-time query of all servers
	ctx := context.Background()
	results := pool.QueryAll(ctx)

	fmt.Println("=== Server Status ===")
	for addr, status := range results {
		if status.Online {
			fmt.Printf("✅ %s: %s\n", status.Info.Name, status.Info.Map)
			fmt.Printf("   Players: %d/%d | Bots: %d | Ping: %v\n",
				status.Info.Players, status.Info.MaxPlayers,
				status.Info.Bots, status.QueryTime)
		} else {
			fmt.Printf("❌ %s: OFFLINE (%v)\n", addr, status.Error)
		}
	}
}

// Example of continuous monitoring with callbacks
func ExampleContinuousMonitoring() {
	pool := NewServerPool()

	// Add servers
	pool.AddServer("localhost:27102", "Server 1")
	pool.AddServer("localhost:27103", "Server 2")
	pool.AddServer("localhost:27104", "Server 3")
	pool.AddServer("localhost:27105", "Server 4")
	pool.AddServer("localhost:27106", "Server 5")
	pool.AddServer("localhost:27107", "Server 6")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor every 30 seconds
	go pool.Monitor(ctx, 30*time.Second, func(results map[string]*ServerStatus) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		log.Printf("[%s] Queried %d servers", timestamp, len(results))

		totalPlayers := 0
		onlineCount := 0

		for addr, status := range results {
			if status.Online {
				onlineCount++
				totalPlayers += int(status.Info.Players)
				log.Printf("  %s: %d players on %s",
					status.Info.Name, status.Info.Players, status.Info.Map)
			} else {
				log.Printf("  %s: OFFLINE", addr)
			}
		}

		log.Printf("Summary: %d/%d servers online, %d total players",
			onlineCount, len(results), totalPlayers)
	})

	// Keep running (in real app, wait for signal or condition)
	time.Sleep(5 * time.Minute)
}

// Example of querying specific server when needed
func ExampleTargetedQuery() {
	pool := NewServerPool()

	// Add servers
	pool.AddServer("main-server:27102", "Main Server")
	pool.AddServer("backup-server:27102", "Backup Server")

	ctx := context.Background()

	// Query just the main server
	status, err := pool.QueryServer(ctx, "main-server:27102")
	if err != nil {
		log.Printf("Failed to query main server: %v", err)

		// Try backup
		status, err = pool.QueryServer(ctx, "backup-server:27102")
		if err != nil {
			log.Printf("Backup server also unavailable: %v", err)
			return
		}
	}

	if status.Online {
		fmt.Printf("Connected to: %s\n", status.Info.Name)
		fmt.Printf("Players: %d/%d\n", status.Info.Players, status.Info.MaxPlayers)
	}
}

// Example of integrating with your tracker database
func ExampleDatabaseIntegration() {
	pool := NewServerPool()

	// Load servers from config/database
	pool.AddServer("server1:27102", "Server 1")
	pool.AddServer("server2:27102", "Server 2")

	ctx := context.Background()

	// Query all servers
	results := pool.QueryAll(ctx)

	// Update database with current status
	for addr, status := range results {
		if status.Online {
			// Update server metadata in database
			fmt.Printf("UPDATE servers SET online=true, current_map='%s', "+
				"current_players=%d, last_seen=NOW() WHERE address='%s'\n",
				status.Info.Map, status.Info.Players, addr)

			// Compare with log-based player list
			if len(status.Players) > 0 {
				fmt.Printf("Currently online players for %s:\n", addr)
				for _, p := range status.Players {
					fmt.Printf("  - %s (Score: %d, Time: %.0fs)\n",
						p.Name, p.Score, p.Duration)
				}
			}
		} else {
			// Mark as offline
			fmt.Printf("UPDATE servers SET online=false WHERE address='%s'\n", addr)
		}
	}
}

// Example of handling server alerts
func ExampleServerAlerts() {
	pool := NewServerPool()

	pool.AddServer("server1:27102", "Main Server")
	pool.AddServer("server2:27102", "Secondary Server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track previous state
	previousState := make(map[string]bool)

	pool.Monitor(ctx, 30*time.Second, func(results map[string]*ServerStatus) {
		for addr, status := range results {
			wasOnline := previousState[addr]
			isOnline := status.Online

			// State changed
			if wasOnline != isOnline {
				if isOnline {
					log.Printf("🟢 SERVER UP: %s is now online", status.Info.Name)
				} else {
					log.Printf("🔴 SERVER DOWN: %s went offline", addr)
				}
			}

			// Check if server is full
			if isOnline && status.Info.Players >= status.Info.MaxPlayers {
				log.Printf("⚠️  SERVER FULL: %s (%d/%d)",
					status.Info.Name, status.Info.Players, status.Info.MaxPlayers)
			}

			// Check if server is empty for too long
			if isOnline && status.Info.Players == 0 {
				log.Printf("ℹ️  Empty server: %s", status.Info.Name)
			}

			previousState[addr] = isOnline
		}
	})

	time.Sleep(10 * time.Minute)
}

// Example with custom timeout and context
func ExampleWithTimeout() {
	// Create client with custom timeout
	client := NewClientWithTimeout(3 * time.Second)
	pool := NewServerPoolWithClient(client)

	pool.AddServer("slow-server:27102", "Slow Server")
	pool.AddServer("fast-server:27102", "Fast Server")

	// Query with overall timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results := pool.QueryAll(ctx)

	for addr, status := range results {
		if status.Online {
			fmt.Printf("%s: OK (%.2fs)\n", addr, status.QueryTime.Seconds())
		} else {
			fmt.Printf("%s: Failed after %.2fs - %v\n",
				addr, status.QueryTime.Seconds(), status.Error)
		}
	}
}

// Example of load balancing / server selection
func ExampleServerSelection() {
	pool := NewServerPool()

	// Add all servers
	for i := 1; i <= 6; i++ {
		pool.AddServer(
			fmt.Sprintf("server%d:27102", i),
			fmt.Sprintf("Server %d", i),
		)
	}

	ctx := context.Background()
	results := pool.QueryAll(ctx)

	// Find least populated server
	var bestServer *ServerStatus
	minPlayers := 999

	for _, status := range results {
		if status.Online {
			players := int(status.Info.Players)
			if players < minPlayers && players < int(status.Info.MaxPlayers) {
				minPlayers = players
				bestServer = status
			}
		}
	}

	if bestServer != nil {
		fmt.Printf("Recommended server: %s (%d/%d players)\n",
			bestServer.Info.Name,
			bestServer.Info.Players,
			bestServer.Info.MaxPlayers)
	} else {
		fmt.Println("No available servers found")
	}
}
