# A2S (Source Engine Query Protocol) Client

A Go implementation of the A2S protocol for querying Source Engine game servers, including Insurgency: Sandstorm.

## Features

- **Server Info Query (A2S_INFO)**: Get server name, map, player count, game version, etc.
- **Player List Query (A2S_PLAYER)**: Get list of connected players with scores and playtime
- **Server Rules Query (A2S_RULES)**: Get server configuration variables (cvars)
- **Challenge Support**: Automatic challenge number handling for player and rules queries
- **Configurable Timeout**: Customize network timeout for queries
- **Context Support**: Full context.Context integration for cancellation and timeouts
- **Rate Limiting**: Built-in rate limiting (1 query/sec per server) to prevent blocking
- **ServerPool**: High-level API for managing and monitoring multiple servers concurrently
- **Concurrent Queries**: Efficiently query multiple servers in parallel

## Production Ready Features (New!)

✅ **Context Support** - Cancel queries, set deadlines, propagate timeouts  
✅ **Rate Limiting** - Automatic per-server rate limiting to prevent query blocking  
✅ **ServerPool** - High-level API for managing multiple servers  
✅ **Concurrent Monitoring** - Query 6+ servers efficiently in parallel  
✅ **Health Tracking** - Track server online/offline state and last query times

## Protocol Reference

Based on the Valve A2S protocol specification:

- [Valve Server Queries](https://developer.valvesoftware.com/wiki/Server_queries)

## Usage

### Managing Multiple Servers (Recommended for 6+ servers)

```go
package main

import (
	"context"
	"log"
	"time"

	"sandstorm-tracker/internal/a2s"
)

func main() {
	// Create a server pool
	pool := a2s.NewServerPool()

	// Add your 6 servers
	pool.AddServer("server1.com:27102", "US East")
	pool.AddServer("server2.com:27102", "US West")
	pool.AddServer("server3.com:27102", "EU")
	pool.AddServer("server4.com:27102", "Asia")
	pool.AddServer("server5.com:27102", "Test")
	pool.AddServer("server6.com:27102", "Event")

	ctx := context.Background()

	// Query all servers concurrently (efficient!)
	results := pool.QueryAll(ctx)

	for addr, status := range results {
		if status.Online {
			log.Printf("%s: %d/%d players on %s",
				status.Info.Name,
				status.Info.Players,
				status.Info.MaxPlayers,
				status.Info.Map)
		} else {
			log.Printf("%s: OFFLINE - %v", addr, status.Error)
		}
	}
}
```

### Continuous Monitoring

```go
func monitorServers() {
	pool := a2s.NewServerPool()
	pool.AddServer("localhost:27102", "My Server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor every 30 seconds with automatic rate limiting
	pool.Monitor(ctx, 30*time.Second, func(results map[string]*a2s.ServerStatus) {
		totalPlayers := 0
		for _, status := range results {
			if status.Online {
				totalPlayers += int(status.Info.Players)
			}
		}
		log.Printf("Total players across all servers: %d", totalPlayers)
	})
}
```

### Basic Server Info Query

```go
package main

import (
	"fmt"
	"log"

	"sandstorm-tracker/internal/a2s"
)

func main() {
	// Create client
	client := a2s.NewClient()

	// Query server info
	// Note: Query port is usually game port + 1
	// For example, if game port is 27101, query port is 27102
	info, err := client.QueryInfo("yourserver.com:27102")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server: %s\n", info.Name)
	fmt.Printf("Map: %s\n", info.Map)
	fmt.Printf("Players: %d/%d\n", info.Players, info.MaxPlayers)
	fmt.Printf("Bots: %d\n", info.Bots)
	fmt.Printf("Game: %s\n", info.Game)
	fmt.Printf("Version: %s\n", info.Version)
}
```

### Query with Context (Cancellation & Timeout)

```go
// Query with 5 second timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

client := a2s.NewClient()
info, err := client.QueryInfoContext(ctx, "server.com:27102")
if err != nil {
	log.Fatal(err)
}
```

### Query Players

```go
client := a2s.NewClient()

players, err := client.QueryPlayers("yourserver.com:27102")
if err != nil {
	log.Fatal(err)
}

for _, player := range players {
	fmt.Printf("[%d] %s - Score: %d, Time: %.0fs\n",
		player.Index, player.Name, player.Score, player.Duration)
}
```

### Query Server Rules

```go
client := a2s.NewClient()

rules, err := client.QueryRules("yourserver.com:27102")
if err != nil {
	log.Fatal(err)
}

for _, rule := range rules {
	fmt.Printf("%s = %s\n", rule.Name, rule.Value)
}
```

### Custom Timeout

```go
import "time"

// Create client with 10 second timeout
client := a2s.NewClientWithTimeout(10 * time.Second)

info, err := client.QueryInfo("yourserver.com:27102")
// ...
```

## Data Structures

### ServerInfo

```go
type ServerInfo struct {
	Protocol    byte    // Protocol version
	Name        string  // Server name
	Map         string  // Current map
	Folder      string  // Game directory
	Game        string  // Game name
	ID          uint16  // Steam App ID
	Players     byte    // Current players
	MaxPlayers  byte    // Maximum players
	Bots        byte    // Number of bots
	ServerType  byte    // 'd' = dedicated, 'l' = listen, 'p' = proxy
	Environment byte    // 'l' = Linux, 'w' = Windows, 'm' = Mac
	Visibility  byte    // 0 = public, 1 = private
	VAC         byte    // 0 = unsecured, 1 = VAC secured
	Version     string  // Game version

	// Optional Extended Data (may be nil)
	Port        *uint16
	SteamID     *uint64
	SourceTVPort *uint16
	SourceTVName *string
	Keywords    *string
	GameID      *uint64
}
```

### Player

```go
type Player struct {
	Index    byte    // Player index
	Name     string  // Player name
	Score    int32   // Player score
	Duration float32 // Time on server in seconds
}
```

### Rule

```go
type Rule struct {
	Name  string // Rule/cvar name
	Value string // Rule/cvar value
}
```

## Finding Your Query Port

For Insurgency: Sandstorm servers:

- **Game Port**: Usually 27101 (configured in server settings)
- **Query Port**: Game port + 1 (e.g., 27102)

You can verify the query port in your server configuration or by checking the server's network bindings.

## Testing

Run all tests:

```bash
go test ./internal/a2s/
```

Run tests with live server (skip by default):

```bash
# Update address in test files first
go test ./internal/a2s/ -v -run TestQueryInfo_Live
```

## Integration with Sandstorm Tracker

This A2S client can be used to:

- Monitor server population in real-time
- Verify server status before log processing
- Display current server stats in dashboards
- Track player join/leave events independently of logs
- **NEW**: Efficiently monitor 6+ servers concurrently with automatic rate limiting
- **NEW**: Get alerts when servers go online/offline
- **NEW**: Find least populated server for load balancing

### Example Integration

```go
// In your tracker startup
pool := a2s.NewServerPool()

// Load servers from config
for _, server := range config.Servers {
	pool.AddServer(server.Address, server.Name)
}

// Background monitoring
go pool.Monitor(ctx, 30*time.Second, func(results map[string]*a2s.ServerStatus) {
	// Update database with current server states
	for addr, status := range results {
		updateServerStatus(addr, status)
	}
})
```

## Error Handling

Common errors:

- **Connection timeout**: Server is offline or firewall blocking
- **Unexpected response**: Server protocol mismatch or corrupted packet
- **Challenge failure**: Server not responding to challenge requests

Always check errors and implement retry logic for production use:

```go
client := a2s.NewClient()

var info *a2s.ServerInfo
var err error

for i := 0; i < 3; i++ {
	info, err = client.QueryInfo(address)
	if err == nil {
		break
	}
	time.Sleep(time.Second)
}

if err != nil {
	log.Printf("Failed after retries: %v", err)
}
```

## Protocol Constants

```go
const (
	A2S_INFO   = 0x54 // Request server info
	A2S_PLAYER = 0x55 // Request player list
	A2S_RULES  = 0x56 // Request server rules

	DEFAULT_TIMEOUT = 5 * time.Second
)
```

## Limitations

- Maximum packet size ~1400 bytes (UDP limitation)
- ~~Some servers may rate-limit queries~~ **✅ FIXED: Built-in rate limiting (1 query/sec per server)**
- Challenge numbers expire (handled automatically)
- Multi-packet responses not yet implemented (rarely needed for <50 players)

## Production Readiness

**For monitoring 6 servers:**

- ✅ **Fully Production Ready**
- ✅ Context support for cancellation
- ✅ Automatic rate limiting per server
- ✅ Concurrent querying with ServerPool
- ✅ Error handling and retry support
- ✅ Comprehensive test coverage

**Recommended for:**

- ✅ Monitoring 1-20 servers
- ✅ Query intervals of 30+ seconds
- ✅ Servers with <50 players
- ✅ Integration with logging/tracking systems

**Future Enhancements** (not needed for typical use):

- Multi-packet response handling (for 50+ player servers)
- Connection pooling (marginal benefit for UDP)
- BZIP2 decompression support
