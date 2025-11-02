package app

import (
	"context"
	"log"
	"sandstorm-tracker/internal/a2s"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterA2SCron sets up a cron job that queries all configured servers via A2S
// and updates current player scores every minute
func RegisterA2SCron(app core.App, cfg *AppConfig) {
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
func updatePlayerScoresFromA2S(ctx context.Context, pbApp core.App, cfg *AppConfig) {
	a2sClient := a2s.NewClientWithTimeout(5 * time.Second)

	for _, serverCfg := range cfg.Servers {
		if !serverCfg.Enabled {
			continue
		}

		// Use queryAddress if available, otherwise use rconAddress
		queryAddr := serverCfg.RconAddress
		if serverCfg.QueryAddress != "" {
			queryAddr = serverCfg.QueryAddress
		}

		// Query players from this server
		players, err := a2sClient.QueryPlayersContext(ctx, queryAddr)
		if err != nil {
			log.Printf("[A2S] Failed to query players from %s (%s): %v", serverCfg.Name, queryAddr, err)
			continue
		}

		log.Printf("[A2S] Queried %d players from %s", len(players), serverCfg.Name)

		// Find or create server record
		serverRecord, err := getOrCreateServerFromConfig(pbApp, serverCfg)
		if err != nil {
			log.Printf("[A2S] Failed to get server record for %s: %v", serverCfg.Name, err)
			continue
		}

		// Find active match for this server
		activeMatch, err := getActiveMatchForServer(pbApp, serverRecord.Id)
		if err != nil {
			log.Printf("[A2S] Failed to get active match for server %s: %v", serverCfg.Name, err)
			continue
		}

		if activeMatch == nil {
			log.Printf("[A2S] No active match found for server %s, skipping player update", serverCfg.Name)
			continue
		}

		// Update each player's score in the active match
		for _, player := range players {
			if player.Name == "" {
				continue // Skip unnamed players
			}

			// Find or create player record by name (A2S doesn't provide Steam ID)
			playerRecord, err := findOrCreatePlayerByName(pbApp, player.Name)
			if err != nil {
				log.Printf("[A2S] Failed to find/create player %s: %v", player.Name, err)
				continue
			}

			// Update match_weapon_stats with current score
			// Note: A2S provides total score, not kills, so we use it as-is
			err = updatePlayerMatchScore(pbApp, activeMatch.Id, playerRecord.Id, player.Score)
			if err != nil {
				log.Printf("[A2S] Failed to update score for player %s: %v", player.Name, err)
			}
		}
	}
}

// getOrCreateServerFromConfig finds or creates a server record based on config
func getOrCreateServerFromConfig(pbApp core.App, cfg ServerConfig) (*core.Record, error) {
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
