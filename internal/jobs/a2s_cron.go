package jobs

import (
	"context"
	"log"
	"sandstorm-tracker/internal/a2s"
	"sandstorm-tracker/internal/config"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// AppInterface defines the methods jobs need from the app
type AppInterface interface {
	core.App
	GetA2SPool() *a2s.ServerPool
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

			// Update match_weapon_stats with current score
			// Note: A2S provides total score, not kills, so we use it as-is
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

// updatePlayerMatchScore updates a player's score in the match_weapon_stats table
func updatePlayerMatchScore(pbApp core.App, matchID, playerID string, score int32) error {
	collection, err := pbApp.FindCollectionByNameOrId("match_weapon_stats")
	if err != nil {
		return err
	}

	// Try to find existing stat record for this player in this match
	record, err := pbApp.FindFirstRecordByFilter(
		"match_weapon_stats",
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
		record.Set("weapon", "a2s_score") // Use special weapon name to indicate A2S data
		record.Set("kills", int(score))
	} else {
		// Update existing record
		record.Set("kills", int(score))
	}

	return pbApp.Save(record)
}
