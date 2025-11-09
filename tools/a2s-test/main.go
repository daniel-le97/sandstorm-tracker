package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sandstorm-tracker/internal/a2s"
	"strings"
	"time"
)

func main() {
	// Parse command line flags
	address := flag.String("address", "127.0.0.1:27131", "Server address to query (host:port)")
	timeout := flag.Duration("timeout", 5*time.Second, "Query timeout")
	continuous := flag.Bool("continuous", false, "Run continuous queries every 5 seconds")
	flag.Parse()

	fmt.Printf("A2S Query Test Tool\n")
	fmt.Printf("===================\n")
	fmt.Printf("Target: %s\n", *address)
	fmt.Printf("Timeout: %v\n\n", *timeout)

	// Create A2S client
	client := a2s.NewClientWithTimeout(*timeout)

	if *continuous {
		fmt.Println("Running continuous queries (Ctrl+C to stop)...")
		runContinuous(client, *address)
	} else {
		runOnce(client, *address)
	}
}

func runOnce(client *a2s.Client, address string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query server info
	fmt.Println("Querying server info...")
	info, err := client.QueryInfoContext(ctx, address)
	if err != nil {
		log.Fatalf("Failed to query server info: %v", err)
	}

	printServerInfo(info)

	// Query players
	fmt.Println("\nQuerying players...")
	players, err := client.QueryPlayersContext(ctx, address)
	if err != nil {
		fmt.Printf("Failed to query players: %v\n", err)
		fmt.Println("\nThis might mean:")
		fmt.Println("  - Server doesn't respond to A2S_PLAYER queries")
		fmt.Println("  - Wrong port (should be query port, not game port)")
		fmt.Println("  - Server is not running")
		fmt.Println("  - Firewall blocking UDP packets")
		os.Exit(1)
	}

	printPlayers(players)
}

func runContinuous(client *a2s.Client, address string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// First query immediately
	queryAndPrint(client, address)

	for range ticker.C {
		fmt.Println("\n" + strings.Repeat("=", 60))
		queryAndPrint(client, address)
	}
}

func queryAndPrint(client *a2s.Client, address string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\n[%s] Querying %s\n", timestamp, address)

	// Query server info
	info, err := client.QueryInfoContext(ctx, address)
	if err != nil {
		fmt.Printf("❌ Failed to query server info: %v\n", err)
		return
	}

	fmt.Printf("✓ Server Info: %s\n", info.Name)
	fmt.Printf("  Map: %s | Players: %d/%d | Bots: %d\n",
		info.Map, info.Players, info.MaxPlayers, info.Bots)

	// Query players
	players, err := client.QueryPlayersContext(ctx, address)
	if err != nil {
		fmt.Printf("❌ Failed to query players: %v\n", err)
		return
	}

	fmt.Printf("✓ Player Query: %d players\n", len(players))
	for i, player := range players {
		fmt.Printf("  %d. %-20s Score: %4d  Time: %6.1fs\n",
			i+1, player.Name, player.Score, player.Duration)
	}
}

func printServerInfo(info *a2s.ServerInfo) {
	fmt.Println("\n=== Server Information ===")
	fmt.Printf("Name:        %s\n", info.Name)
	fmt.Printf("Map:         %s\n", info.Map)
	fmt.Printf("Game:        %s\n", info.Game)
	fmt.Printf("Folder:      %s\n", info.Folder)
	fmt.Printf("Players:     %d/%d\n", info.Players, info.MaxPlayers)
	fmt.Printf("Bots:        %d\n", info.Bots)
	fmt.Printf("Server Type: %c\n", info.ServerType)
	fmt.Printf("Environment: %c\n", info.Environment)
	fmt.Printf("Visibility:  %d (0=public, 1=private)\n", info.Visibility)
	fmt.Printf("VAC:         %d (0=unsecured, 1=secured)\n", info.VAC)
	fmt.Printf("Version:     %s\n", info.Version)

	if info.Port != nil {
		fmt.Printf("Port:        %d\n", *info.Port)
	}
	if info.Keywords != nil {
		fmt.Printf("Keywords:    %s\n", *info.Keywords)
	}
	if info.GameID != nil {
		fmt.Printf("Game ID:     %d\n", *info.GameID)
	}
}

func printPlayers(players []a2s.Player) {
	fmt.Printf("\n=== Players (%d) ===\n", len(players))
	if len(players) == 0 {
		fmt.Println("No players currently on the server")
		return
	}

	fmt.Printf("%-4s %-30s %10s %12s\n", "#", "Name", "Score", "Duration")
	fmt.Println(strings.Repeat("-", 60))

	for i, player := range players {
		duration := formatDuration(player.Duration)
		fmt.Printf("%-4d %-30s %10d %12s\n",
			i+1, player.Name, player.Score, duration)
	}
}

func formatDuration(seconds float32) string {
	d := time.Duration(seconds * float32(time.Second))

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
