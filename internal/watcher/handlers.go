package watcher

import (
	"context"
	"fmt"
	"log"
	"net"

	// "strconv"

	generated "sandstorm-tracker/internal/db/generated"

	"sandstorm-tracker/internal/events"
	"sandstorm-tracker/internal/rcon"
)

func (fw *FileWatcher) handleKillEvent(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	killersData, ok := event.Data["killers"].([]events.Killer)
	if !ok {
		return fmt.Errorf("invalid killers data in event")
	}
	victimName, _ := event.Data["victim_name"].(string)
	weapon, _ := event.Data["weapon"].(string)
	// For each killer, upsert player, upsert player_stats, and upsert weapon_stats
	for _, killer := range killersData {
		if killer.SteamID == "INVALID" {
			continue
		}
		// Upsert player
		player, err := fw.db.GetQueries().GetPlayerByExternalID(ctx, killer.SteamID)
		if err != nil {
			player, err = fw.db.GetQueries().CreatePlayer(ctx, generated.CreatePlayerParams{
				ExternalID: killer.SteamID,
				Name:       killer.Name,
			})
			if err != nil {
				return fmt.Errorf("failed to create player: %w", err)
			}
		}
		// Upsert player_stats
		stats, err := fw.db.GetQueries().GetPlayerStatsByPlayerID(ctx, player.ID)
		if err != nil {
			stats, err = fw.db.GetQueries().CreatePlayerStats(ctx, generated.CreatePlayerStatsParams{
				ID:                killer.SteamID,
				PlayerID:          player.ID,
				ServerID:          serverDBID,
				GamesPlayed:       nil,
				Wins:              nil,
				Losses:            nil,
				TotalScore:        nil,
				TotalPlayTime:     nil,
				LastLogin:         nil,
				TotalKills:        nil,
				TotalDeaths:       nil,
				FriendlyFireKills: nil,
				HighestScore:      nil,
			})
			if err != nil {
				return fmt.Errorf("failed to create player_stats: %w", err)
			}
		}
		// Upsert weapon_stats (increment kills)
		one := int64(1)
		zero := int64(0)
		_, _ = fw.db.GetQueries().UpsertWeaponStats(ctx, generated.UpsertWeaponStatsParams{
			PlayerStatsID: stats.ID,
			WeaponName:    weapon,
			Kills:         &one,
			Assists:       &zero,
		})
		log.Printf("Kill recorded: %s killed %s with %s", killer.Name, victimName, weapon)
	}
	return nil
}

func (fw *FileWatcher) handlePlayerJoin(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	if steamID == "INVALID" || steamID == "" {
		log.Printf("Skipping player join for %s - no Steam ID provided", playerName)
		return nil
	}
	_, err := fw.db.GetQueries().GetPlayerByExternalID(ctx, steamID)
	if err != nil {
		_, err = fw.db.GetQueries().CreatePlayer(ctx, generated.CreatePlayerParams{
			ExternalID: steamID,
			Name:       playerName,
		})
	}
	return err
}

func (fw *FileWatcher) handlePlayerLeave(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	if steamID == "INVALID" || steamID == "" {
		log.Printf("Player %s left - no Steam ID to track", playerName)
		return nil
	}
	log.Printf("Player %s [%s] left the server", playerName, steamID)
	return nil
}

func (fw *FileWatcher) handleChatCommand(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	command, _ := event.Data["command"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	log.Printf("Chat command from %s: %s", playerName, command)

	// Find server config by serverDBID or event.ServerID (adjust as needed)
	serverID := event.ServerID
	sc, ok := fw.serverConfigs[event.ServerID]
	if !ok {
		log.Printf("No server config found for serverID: %s", serverID)
		return nil
	}

	fw.rconMu.Lock()
	client, exists := fw.rconClients[serverID]
	fw.rconMu.Unlock()
	if !exists {
		// Create new RCON client
		conn, err := net.Dial("tcp", sc.RconAddress)
		if err != nil {
			log.Printf("Failed to dial RCON for %s: %v", sc.RconAddress, err)
			return err
		}
		client = rcon.NewRconClient(conn, nil)
		if !client.Auth(sc.RconPassword) {
			log.Printf("RCON auth failed for server %s", serverID)
			conn.Close()
			return fmt.Errorf("RCON auth failed")
		}
		fw.rconMu.Lock()
		fw.rconClients[serverID] = client
		fw.rconMu.Unlock()
	}

	player, err := fw.db.GetQueries().GetPlayerByExternalID(ctx, steamID)
	if err != nil {
		log.Printf("Could not find player for SteamID %s: %v", steamID, err)
		return err
	}

	stats, err := fw.db.GetQueries().GetPlayerStatsByPlayerID(ctx, player.ID)
	if err != nil {
		log.Printf("No player_stats for player %s: %v", player.Name, err)
		return err
	}

	// !kdr command
	if command == "!kdr" {
		kills := int64(0)
		deaths := int64(0)
		if stats.TotalKills != nil {
			kills = *stats.TotalKills
		}
		if stats.TotalDeaths != nil {
			deaths = *stats.TotalDeaths
		}
		kdr := 0.0
		if deaths == 0 {
			if kills == 0 {
				kdr = 0.0
			} else {
				kdr = float64(kills)
			}
		} else {
			kdr = float64(kills) / float64(deaths)
		}
		msg := fmt.Sprintf("%s: kills:%d, deaths:%d, KDR: %.2f", playerName, kills, deaths, kdr)
		_, err = client.Send("say " + msg)
		if err != nil {
			log.Printf("Failed to send RCON say: %v", err)
		}
	}

	// !guns command
	if command == "!guns" {
		weaponStats, err := fw.db.GetQueries().GetWeaponStatsForPlayerStats(ctx, stats.ID)
		if err != nil || len(weaponStats) == 0 {
			msg := fmt.Sprintf("%s: No weapon stats found.", playerName)
			_, err = client.Send("say " + msg)
			if err != nil {
				log.Printf("Failed to send RCON say: %v", err)
			}
		} else {
			// Sort by kills descending
			type ws struct {
				Name  string
				Kills int64
			}
			var wsList []ws
			for _, w := range weaponStats {
				kills := int64(0)
				if w.Kills != nil {
					kills = *w.Kills
				}
				wsList = append(wsList, ws{Name: w.WeaponName, Kills: kills})
			}
			// Simple bubble sort (small list)
			for i := 0; i < len(wsList); i++ {
				for j := i + 1; j < len(wsList); j++ {
					if wsList[j].Kills > wsList[i].Kills {
						wsList[i], wsList[j] = wsList[j], wsList[i]
					}
				}
			}
			// Top 3
			weaponLines := ""
			for i := 0; i < len(wsList) && i < 3; i++ {
				weaponLines += fmt.Sprintf("%d. %s (%d kills)", i+1, wsList[i].Name, wsList[i].Kills)
				if i < 2 && i < len(wsList)-1 {
					weaponLines += "| "
				}
			}
			msg := fmt.Sprintf("%s: %s", playerName, weaponLines)
			_, err = client.Send("say " + msg)
			if err != nil {
				log.Printf("Failed to send RCON say: %v", err)
			}
		}
	}

	// !stats command
	if command == "!stats" {
		kills := int64(0)
		deaths := int64(0)
		ff := int64(0)
		if stats.TotalKills != nil {
			kills = *stats.TotalKills
		}
		if stats.TotalDeaths != nil {
			deaths = *stats.TotalDeaths
		}
		if stats.FriendlyFireKills != nil {
			ff = *stats.FriendlyFireKills
		}
		msg := fmt.Sprintf("%s: Kills: %d, Deaths: %d, FF Kills: %d", playerName, kills, deaths, ff)
		_, err := client.Send("say " + msg)
		if err != nil {
			log.Printf("Failed to send RCON say: %v", err)
		}
	}

	// !top command: top 3 players by score/min for this server
	if command == "!top" {
		// Query top 3 player_stats for this server, ordered by score/minute
		topStats, err := fw.db.GetQueries().GetTopPlayersByScorePerMin(ctx, serverDBID)
		if err != nil || len(topStats) == 0 {
			msg := fmt.Sprintf("%s: No stats found for this server.", playerName)
			_, err := client.Send("say " + msg)
			if err != nil {
				log.Printf("Failed to send RCON say: %v", err)
			}
		} else {
			msg := "Top players by score/min: "
			for i, stat := range topStats {
				pname := stat.PlayerName
				score := int64(0)
				mins := float64(0)
				if stat.TotalScore != nil {
					score = *stat.TotalScore
				}
				if stat.TotalPlayTime != nil && *stat.TotalPlayTime > 0 {
					mins = float64(*stat.TotalPlayTime) / 60.0
				}
				spm := 0.0
				if mins > 0 {
					spm = float64(score) / mins
				}
				msg += fmt.Sprintf("%d. %s (%.1f spm)", i+1, pname, spm)
				if i < 2 && i < len(topStats)-1 {
					msg += " | "
				}
			}
			_, err := client.Send("say " + msg)
			if err != nil {
				log.Printf("Failed to send RCON say: %v", err)
			}
		}
	}

	// Add more command handling as needed
	return nil
}
