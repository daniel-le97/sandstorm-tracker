package rcon

import (
	"fmt"
	"log/slog"
)

// ExamplePoolUsage demonstrates how to use the ClientPool
func ExamplePoolUsage() {
	// Create a pool with a logger
	pool := NewClientPool(slog.Default())

	// Add servers to the pool
	pool.AddServer("server1", &ServerConfig{
		Address:  "127.0.0.1:27015",
		Password: "mypassword",
	})

	pool.AddServer("server2", &ServerConfig{
		Address:  "127.0.0.1:27016",
		Password: "anotherpassword",
	})

	// Send commands to specific servers
	response, err := pool.SendCommand("server1", "listplayers")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Server 1 response: %s\n", response)

	// Get a client directly for multiple commands
	client, err := pool.GetClient("server2")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Send multiple commands using the same client
	response1, _ := client.Send("listplayers")
	response2, _ := client.Send("listmaps")
	fmt.Printf("Players: %s\n", response1)
	fmt.Printf("Maps: %s\n", response2)

	// List all servers in the pool
	servers := pool.ListServers()
	fmt.Printf("Servers in pool: %v\n", servers)

	// Check if a server is connected
	if pool.IsConnected("server1") {
		fmt.Println("Server 1 is connected")
	}

	// Close all connections when done
	pool.CloseAll()
}
