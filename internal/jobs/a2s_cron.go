package jobs

import (
	"context"
	"fmt"
	"log"
	"sandstorm-tracker/internal/a2s"
	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/util"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// AppInterface defines the methods jobs need from the app
type AppInterface interface {
	core.App
	GetA2SPool() *a2s.ServerPool
	SendRconCommand(serverID string, command string) (string, error)
}

// RegisterA2S sets up a cron job that queries all configured servers via A2S
// and updates current player scores every minute
func RegisterA2S(app AppInterface, cfg *config.Config) {
	scheduler := app.Cron()

	// Run A2S queries every minute on all configured servers
	scheduler.MustAdd("a2s_player_scores", "* * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		updatePlayerScoresFromA2S(ctx, app, cfg)
	})

	log.Println("[A2S] Registered cron job to update player scores every minute")
}

// RegisterA2SForServer sets up a cron job for a specific server
// This is called when a server becomes active (log rotation detected)
func RegisterA2SForServer(app AppInterface, cfg *config.Config, serverID string) {
	// Find the server config for this serverID
	var serverCfg *config.ServerConfig
	for i, sc := range cfg.Servers {
		if !sc.Enabled {
			continue
		}

		// Extract serverID from logPath using util function
		pathServerID, err := util.GetServerIdFromPath(sc.LogPath)
		if err != nil {
			continue
		}

		if pathServerID == serverID {
			serverCfg = &cfg.Servers[i]
			break
		}
	}

	if serverCfg == nil {
		log.Printf("[A2S] Could not find server config for serverID %s", serverID)
		return
	}

	scheduler := app.Cron()

	// Create unique job name for this server
	jobName := fmt.Sprintf("a2s_player_scores_%s", serverID)

	// Run A2S query every minute for this specific server
	// Note: Cron only supports standard 5-field expressions (no @every syntax)
	scheduler.MustAdd(jobName, "* * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		pool := app.GetA2SPool()
		if pool == nil {
			return
		}

		queryAddr := serverCfg.QueryAddress
		if queryAddr == "" {
			queryAddr = serverCfg.RconAddress
		}

		// Query just this server
		status, err := pool.QueryServer(ctx, queryAddr)
		if err != nil {
			log.Printf("[A2S] Failed to query server %s at %s: %v", serverCfg.Name, queryAddr, err)
			return
		}
		if !status.Online {
			log.Printf("[A2S] Server %s at %s is offline", serverCfg.Name, queryAddr)
			return
		}

		// Log server info for debugging
		if status.Info != nil {
			log.Printf("[A2S] Server %s info: Players=%d/%d, Map=%s",
				serverCfg.Name, status.Info.Players, status.Info.MaxPlayers, status.Info.Map)
		}

		// Process this server's players using the existing logic
		processServerStatus(ctx, app, *serverCfg, status)
	})

	log.Printf("[A2S] Registered cron job for server %s (%s)", serverID, serverCfg.Name)
}

// UnregisterA2SForServer removes the cron job for a specific server
// This is called when a server becomes inactive (no log activity for 10s)
func UnregisterA2SForServer(app AppInterface, serverID string) {
	scheduler := app.Cron()
	jobName := fmt.Sprintf("a2s_player_scores_%s", serverID)

	scheduler.Remove(jobName)
	log.Printf("[A2S] Unregistered cron job for server %s", serverID)
}

// processServerStatus processes a single server's A2S query result
func processServerStatus(ctx context.Context, app AppInterface, serverCfg config.ServerConfig, status *a2s.ServerStatus) {
	if !status.Online {
		return
	}

	// Only log if there are actually players or if A2S returned player data
	if status.Info != nil && (status.Info.Players > 0 || len(status.Players) > 0) {
		log.Printf("[A2S] Queried %d players from %s (Info reports %d players)",
			len(status.Players), serverCfg.Name, status.Info.Players)
	}

	// Find or create server record
	serverRecord, err := getOrCreateServerFromConfig(app, serverCfg)
	if err != nil {
		log.Printf("[A2S] Failed to get server record for %s: %v", serverCfg.Name, err)
		return
	}

	// Extract serverID from logPath for RCON commands
	serverID, err := util.GetServerIdFromPath(serverCfg.LogPath)
	if err != nil {
		log.Printf("[A2S] Failed to get serverID from path %s: %v", serverCfg.LogPath, err)
		return
	}

	// Find active match for this server
	activeMatch, err := getActiveMatchForServer(app, serverRecord.Id)
	if err != nil {
		log.Printf("[A2S] Failed to get active match for server %s: %v", serverCfg.Name, err)
		return
	}

	if activeMatch == nil {
		log.Printf("[A2S] No active match found for server %s, skipping player update", serverCfg.Name)
		return
	}

	// A2S player queries don't work for Insurgency: Sandstorm, use RCON instead
	// Only use RCON if server info indicates there are actually players (to avoid log pollution)
	if len(status.Players) == 0 && status.Info != nil && status.Info.Players > 0 {
		players, err := queryPlayersViaRcon(app, serverID)
		if err != nil {
			log.Printf("[RCON] Failed to query players via RCON for %s: %v", serverCfg.Name, err)
			return
		}

		// Only log when players are actually found
		if len(players) > 0 {
			log.Printf("[RCON] Found %d players via RCON for %s", len(players), serverCfg.Name)
			updatePlayersFromRcon(app, activeMatch.Id, players)
		}
		return
	}

	// Fallback: Update from A2S data (though this rarely works for Insurgency)
	for _, player := range status.Players {
		if player.Name == "" {
			continue // Skip unnamed players
		}

		// Find or create player record by name (A2S doesn't provide Steam ID)
		playerRecord, err := findOrCreatePlayerByName(app, player.Name)
		if err != nil {
			log.Printf("[A2S] Failed to find/create player %s: %v", player.Name, err)
			continue
		}

		// Update match score
		err = updatePlayerMatchScore(app, activeMatch.Id, playerRecord.Id, int32(player.Score))
		if err != nil {
			log.Printf("[A2S] Failed to update score for player %s: %v", player.Name, err)
		}
	}
}

// updatePlayerScoresFromA2S queries all configured servers via A2S and updates current player scores
func updatePlayerScoresFromA2S(ctx context.Context, app AppInterface, cfg *config.Config) {
	// Use the app's A2S pool to query all servers
	pool := app.GetA2SPool()
	if pool == nil {
		log.Println("[A2S] A2S pool not available")
		return
	}

	// Query all servers at once
	results := pool.QueryAll(ctx)

	for address, status := range results {
		if !status.Online || status.Error != nil {
			log.Printf("[A2S] Server %s offline or error: %v", address, status.Error)
			continue
		}

		// Find server config by query address
		var serverCfg *config.ServerConfig
		for _, sc := range cfg.Servers {
			queryAddr := sc.RconAddress
			if sc.QueryAddress != "" {
				queryAddr = sc.QueryAddress
			}
			if queryAddr == address {
				serverCfg = &sc
				break
			}
		}

		if serverCfg == nil {
			continue
		}

		log.Printf("[A2S] Queried %d players from %s", len(status.Players), serverCfg.Name)

		// Find or create server record
		serverRecord, err := getOrCreateServerFromConfig(app, *serverCfg)
		if err != nil {
			log.Printf("[A2S] Failed to get server record for %s: %v", serverCfg.Name, err)
			continue
		}

		// Find active match for this server
		activeMatch, err := getActiveMatchForServer(app, serverRecord.Id)
		if err != nil {
			log.Printf("[A2S] Failed to get active match for server %s: %v", serverCfg.Name, err)
			continue
		}

		if activeMatch == nil {
			log.Printf("[A2S] No active match found for server %s, skipping player update", serverCfg.Name)
			continue
		}

		// Update each player's score in the active match
		for _, player := range status.Players {
			if player.Name == "" {
				continue // Skip unnamed players
			}

			// Find or create player record by name (A2S doesn't provide Steam ID)
			playerRecord, err := findOrCreatePlayerByName(app, player.Name)
			if err != nil {
				log.Printf("[A2S] Failed to find/create player %s: %v", player.Name, err)
				continue
			}

			// Update match_player_stats with current score
			// Note: A2S provides total score from the game server
			err = updatePlayerMatchScore(app, activeMatch.Id, playerRecord.Id, int32(player.Score))
			if err != nil {
				log.Printf("[A2S] Failed to update score for player %s: %v", player.Name, err)
			}
		}
	}
}

// getOrCreateServerFromConfig finds or creates a server record based on config
func getOrCreateServerFromConfig(pbApp core.App, cfg config.ServerConfig) (*core.Record, error) {
	collection, err := pbApp.FindCollectionByNameOrId("servers")
	if err != nil {
		return nil, err
	}

	// Try to find existing server by external_id (server name)
	record, err := pbApp.FindFirstRecordByFilter(
		"servers",
		"external_id = {:name}",
		dbx.Params{"name": cfg.Name},
	)

	if err == nil {
		return record, nil
	}

	// Create new server record
	record = core.NewRecord(collection)
	record.Set("external_id", cfg.Name)
	record.Set("name", cfg.Name)
	record.Set("path", cfg.LogPath)

	if err := pbApp.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

// getActiveMatchForServer finds the currently active match for a server
func getActiveMatchForServer(pbApp core.App, serverID string) (*core.Record, error) {
	record, err := pbApp.FindFirstRecordByFilter(
		"matches",
		"server = {:serverId} && end_time = ''",
		dbx.Params{"serverId": serverID},
	)

	if err != nil {
		return nil, nil // No active match is not an error
	}

	return record, nil
}

// findOrCreatePlayerByName finds or creates a player by their display name
func findOrCreatePlayerByName(pbApp core.App, name string) (*core.Record, error) {
	collection, err := pbApp.FindCollectionByNameOrId("players")
	if err != nil {
		return nil, err
	}

	// Try to find existing player by name
	record, err := pbApp.FindFirstRecordByFilter(
		"players",
		"name = {:name}",
		dbx.Params{"name": name},
	)

	if err == nil {
		return record, nil
	}

	// Create new player record (use name as external_id since we don't have Steam ID from A2S)
	record = core.NewRecord(collection)
	record.Set("external_id", "a2s_"+name) // Prefix to distinguish from Steam IDs
	record.Set("name", name)

	if err := pbApp.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

// updatePlayerMatchScore updates a player's score in the match_player_stats table
func updatePlayerMatchScore(pbApp core.App, matchID, playerID string, score int32) error {
	collection, err := pbApp.FindCollectionByNameOrId("match_player_stats")
	if err != nil {
		return err
	}

	// Try to find existing stat record for this player in this match
	record, err := pbApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:matchId} && player = {:playerId}",
		dbx.Params{
			"matchId":  matchID,
			"playerId": playerID,
		},
	)

	if err != nil {
		// Create new record if not found
		record = core.NewRecord(collection)
		record.Set("match", matchID)
		record.Set("player", playerID)
		record.Set("score", int(score))
	} else {
		// Update existing record
		record.Set("score", int(score))
	}

	return pbApp.Save(record)
}

// RconPlayer represents a player from RCON listplayers response
type RconPlayer struct {
	Name  string
	Score int32
	NetID string
}

// queryPlayersViaRcon queries players using RCON listplayers command
func queryPlayersViaRcon(app AppInterface, serverID string) ([]RconPlayer, error) {
	response, err := app.SendRconCommand(serverID, "listplayers")
	if err != nil {
		return nil, err
	}

	return parseRconListPlayers(response), nil
}

// parseRconListPlayers parses the RCON listplayers response
// Format: ID | Name | NetID | IP | Score | ...
func parseRconListPlayers(response string) []RconPlayer {
	players := []RconPlayer{}

	lines := splitLines(response)
	for _, line := range lines {
		// Skip header and separator lines
		if len(line) < 10 || line[0] == '=' || containsString(line, "ID") && containsString(line, "Name") {
			continue
		}

		// Parse line by splitting on |
		parts := splitOnPipe(line)
		if len(parts) < 5 {
			continue
		}

		name := trimSpace(parts[1])
		scoreStr := trimSpace(parts[4])
		netID := trimSpace(parts[2])

		// Skip if name is empty, matches default classes, or is a number
		if name == "" || name == "Observer" || name == "Commander" || name == "Marksman" || name == "0" || name == "1" || name == "2" {
			continue
		}

		// Skip if NetID doesn't contain "SteamNWI" (only real players have Steam IDs)
		if !containsString(netID, "SteamNWI") {
			continue
		}

		// Parse score
		score := int32(0)
		fmt.Sscanf(scoreStr, "%d", &score)

		players = append(players, RconPlayer{
			Name:  name,
			Score: score,
			NetID: netID,
		})
	}

	return players
}

// updatePlayersFromRcon updates player scores from RCON data
func updatePlayersFromRcon(app AppInterface, matchID string, players []RconPlayer) {
	successCount := 0
	for _, player := range players {
		// Find or create player record by name
		playerRecord, err := findOrCreatePlayerByName(app, player.Name)
		if err != nil {
			log.Printf("[RCON] Failed to find/create player %s: %v", player.Name, err)
			continue
		}

		// Update match score
		err = updatePlayerMatchScore(app, matchID, playerRecord.Id, player.Score)
		if err != nil {
			log.Printf("[RCON] Failed to update score for player %s: %v", player.Name, err)
		} else {
			successCount++
		}
	}

	if successCount > 0 {
		log.Printf("[RCON] Updated scores for %d players", successCount)
	}
}

// Helper string functions
func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, c := range s {
		if c == '\n' || c == '\r' {
			if len(current) > 0 {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}
	return lines
}

func splitOnPipe(s string) []string {
	parts := []string{}
	current := ""
	for _, c := range s {
		if c == '|' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func trimSpace(s string) string {
	// Trim leading/trailing whitespace
	start := 0
	end := len(s)

	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	if start >= end {
		return ""
	}
	return s[start:end]
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
