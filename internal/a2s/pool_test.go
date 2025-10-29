package a2s

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewServerPool(t *testing.T) {
	pool := NewServerPool()
	if pool == nil {
		t.Fatal("NewServerPool returned nil")
	}
	if pool.client == nil {
		t.Error("Pool client is nil")
	}
	if pool.servers == nil {
		t.Error("Pool servers map is nil")
	}
	if pool.rateLimiter == nil {
		t.Error("Pool rate limiter is nil")
	}
}

func TestAddRemoveServer(t *testing.T) {
	pool := NewServerPool()

	// Add server
	pool.AddServer("localhost:27102", "Test Server")

	servers := pool.ListServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	// Get server
	server, err := pool.GetServer("localhost:27102")
	if err != nil {
		t.Errorf("Failed to get server: %v", err)
	}
	if server.Name != "Test Server" {
		t.Errorf("Expected name 'Test Server', got '%s'", server.Name)
	}

	// Remove server
	pool.RemoveServer("localhost:27102")

	servers = pool.ListServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers after removal, got %d", len(servers))
	}

	// Try to get removed server
	_, err = pool.GetServer("localhost:27102")
	if err == nil {
		t.Error("Expected error when getting removed server, got nil")
	}
}

func TestServerPoolMultipleServers(t *testing.T) {
	pool := NewServerPool()

	// Add multiple servers
	pool.AddServer("server1:27102", "Server 1")
	pool.AddServer("server2:27102", "Server 2")
	pool.AddServer("server3:27102", "Server 3")
	pool.AddServer("server4:27102", "Server 4")
	pool.AddServer("server5:27102", "Server 5")
	pool.AddServer("server6:27102", "Server 6")

	servers := pool.ListServers()
	if len(servers) != 6 {
		t.Errorf("Expected 6 servers, got %d", len(servers))
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(100 * time.Millisecond)

	start := time.Now()

	// First call should not block
	limiter.Wait("test-server")
	elapsed1 := time.Since(start)

	// Second call should block for ~100ms
	limiter.Wait("test-server")
	elapsed2 := time.Since(start)

	if elapsed1 > 50*time.Millisecond {
		t.Errorf("First call blocked for too long: %v", elapsed1)
	}

	if elapsed2 < 90*time.Millisecond {
		t.Errorf("Second call didn't block long enough: %v", elapsed2)
	}

	if elapsed2 > 150*time.Millisecond {
		t.Errorf("Second call blocked for too long: %v", elapsed2)
	}
}

func TestRateLimiterMultipleServers(t *testing.T) {
	limiter := NewRateLimiter(100 * time.Millisecond)

	start := time.Now()

	// Different servers should not block each other
	limiter.Wait("server1")
	limiter.Wait("server2")
	limiter.Wait("server3")

	elapsed := time.Since(start)

	// Should complete quickly since they're different servers
	if elapsed > 50*time.Millisecond {
		t.Errorf("Multi-server rate limiting blocked incorrectly: %v", elapsed)
	}
}

func TestQueryAll_NoServers(t *testing.T) {
	pool := NewServerPool()
	ctx := context.Background()

	results := pool.QueryAll(ctx)

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty pool, got %d", len(results))
	}
}

func TestQueryAll_ContextCancellation(t *testing.T) {
	pool := NewServerPool()
	pool.AddServer("server1:27102", "Server 1")
	pool.AddServer("server2:27102", "Server 2")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	results := pool.QueryAll(ctx)

	// Should still return results even if cancelled
	// but all should have errors
	for addr, status := range results {
		if status.Online {
			t.Errorf("Server %s reported online despite cancelled context", addr)
		}
		if status.Error == nil {
			t.Errorf("Server %s has no error despite cancelled context", addr)
		}
	}
}

func TestQueryServer_NotFound(t *testing.T) {
	pool := NewServerPool()
	ctx := context.Background()

	_, err := pool.QueryServer(ctx, "nonexistent:27102")
	if err == nil {
		t.Error("Expected error when querying non-existent server")
	}
}

func TestServerIsOnline(t *testing.T) {
	server := &Server{
		Address: "test:27102",
		Name:    "Test",
	}

	// Initially should be offline
	if server.IsOnline() {
		t.Error("New server should not be online")
	}

	// Update with info
	info := &ServerInfo{
		Name: "Test Server",
	}
	server.updateStatus(info, nil)

	if !server.IsOnline() {
		t.Error("Server should be online after successful update")
	}

	// Update with error
	server.updateStatus(nil, fmt.Errorf("connection failed"))

	if server.IsOnline() {
		t.Error("Server should be offline after error")
	}
}

func TestMonitor_Cancellation(t *testing.T) {
	pool := NewServerPool()
	pool.AddServer("localhost:27102", "Test")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	callCount := 0
	done := make(chan bool)

	go func() {
		pool.Monitor(ctx, 50*time.Millisecond, func(results map[string]*ServerStatus) {
			callCount++
		})
		done <- true
	}()

	<-done

	// Should have been called at least once (initial query)
	if callCount < 1 {
		t.Errorf("Monitor callback not called, count: %d", callCount)
	}

	// Should not run forever
	if callCount > 5 {
		t.Errorf("Monitor ran too many times: %d", callCount)
	}
}

// Example of how to use the pool
func ExampleServerPool() {
	// Create a pool
	pool := NewServerPool()

	// Add your 6 servers
	pool.AddServer("server1.example.com:27102", "Server 1")
	pool.AddServer("server2.example.com:27102", "Server 2")
	pool.AddServer("server3.example.com:27102", "Server 3")
	pool.AddServer("server4.example.com:27102", "Server 4")
	pool.AddServer("server5.example.com:27102", "Server 5")
	pool.AddServer("server6.example.com:27102", "Server 6")

	// Query all servers
	ctx := context.Background()
	results := pool.QueryAll(ctx)

	for addr, status := range results {
		if status.Online {
			fmt.Printf("%s: %s - %d/%d players\n",
				addr, status.Info.Name, status.Info.Players, status.Info.MaxPlayers)
		} else {
			fmt.Printf("%s: OFFLINE - %v\n", addr, status.Error)
		}
	}
}

// Example of continuous monitoring
func ExampleServerPool_Monitor() {
	pool := NewServerPool()
	pool.AddServer("localhost:27102", "My Server")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Monitor every 30 seconds
	pool.Monitor(ctx, 30*time.Second, func(results map[string]*ServerStatus) {
		for addr, status := range results {
			if status.Online {
				fmt.Printf("[%s] %s - Map: %s, Players: %d/%d\n",
					status.LastQuery.Format("15:04:05"),
					status.Info.Name,
					status.Info.Map,
					status.Info.Players,
					status.Info.MaxPlayers)
			} else {
				fmt.Printf("[%s] %s - OFFLINE: %v\n",
					status.LastQuery.Format("15:04:05"),
					addr,
					status.Error)
			}
		}
	})
}
