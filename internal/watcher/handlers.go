package watcher
import (
	"context"
	"fmt"
	"log"
	"net"

	// "strconv"

	generated "sandstorm-tracker/db/generated"

	"sandstorm-tracker/internal/events"
	"sandstorm-tracker/internal/rcon"

)


func (fw *FileWatcher) handleKillEvent(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	killersData, ok := event.Data["killers"].([]events.Killer)
	if !ok {
		return fmt.Errorf("invalid killers data in event")
	}
	victimName, _ := event.Data["victim_name"].(string)
	victimSteamID, _ := event.Data["victim_steam_id"].(string)
	weapon, _ := event.Data["weapon"].(string)
	killType, _ := event.Data["kill_type"].(string)
	isMultiKill, _ := event.Data["multi_kill"].(bool)
	if victimSteamID != "INVALID" {
		_, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
			ExternalID: victimSteamID,
			Name:       victimName,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert victim: %w", err)
		}
	}
	if victimSteamID == "INVALID" {
		for _, killer := range killersData {
			if killer.SteamID != "INVALID" {
				killerID, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
					ExternalID: killer.SteamID,
					Name:       killer.Name,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert killer %s: %w", killer.Name, err)
				}
				var killTypeInt int64
				switch killType {
				case "suicide":
					killTypeInt = 1
				case "team_kill":
					killTypeInt = 2
				default:
					killTypeInt = 0
				}
				err = fw.db.GetQueries().InsertKill(ctx, generated.InsertKillParams{
					KillerID:   &killerID,
					VictimName: &victimName,
					ServerID:   serverDBID,
					WeaponName: &weapon,
					KillType:   killTypeInt,
					MatchID:    nil,
					CreatedAt:  &event.Timestamp,
				})
				if err != nil {
					return fmt.Errorf("failed to insert kill for killer %s: %w", killer.Name, err)
				}
				log.Printf("AI kill recorded: %s killed %s with %s%s",
					killer.Name, victimName, weapon,
					func() string {
						if isMultiKill {
							return " (multi-kill)"
						} else {
							return ""
						}
					}())
			}
		}
	} else {
		if len(killersData) == 1 {
			killer := killersData[0]
			if killer.SteamID != "INVALID" {
				killerID, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
					ExternalID: killer.SteamID,
					Name:       killer.Name,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert killer %s: %w", killer.Name, err)
				}
				var killTypeInt int64
				switch killType {
				case "regular":
					killTypeInt = 0
				case "suicide":
					killTypeInt = 1
				case "team_kill":
					killTypeInt = 2
				default:
					killTypeInt = 0
				}
				err = fw.db.GetQueries().InsertKill(ctx, generated.InsertKillParams{
					KillerID:   &killerID,
					VictimName: &victimName,
					ServerID:   serverDBID,
					WeaponName: &weapon,
					KillType:   killTypeInt,
					MatchID:    nil,
					CreatedAt:  &event.Timestamp,
				})
				if err != nil {
					return fmt.Errorf("failed to insert kill: %w", err)
				}
				log.Printf("Player kill recorded: %s killed %s with %s", killer.Name, victimName, weapon)
			}
		}
	}
	return nil
}

func (fw *FileWatcher) handlePlayerJoin(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	_ = serverDBID
	playerName, _ := event.Data["player_name"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	if steamID == "INVALID" {
		return nil
	}
	if steamID == "" {
		log.Printf("Skipping player join for %s - no Steam ID provided", playerName)
		return nil
	}
	_, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
		ExternalID: steamID,
		Name:       playerName,
	})
	return err
}

func (fw *FileWatcher) handlePlayerLeave(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	if steamID == "INVALID" {
		return nil
	}
	if steamID == "" {
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

	// TODO - daniel - implement actual in-game command handling logic
	if command == "!kdr" {
		// Look up player row ID from SteamID (external_id)
		player, err := fw.db.GetQueries().GetPlayer(ctx, steamID)
		if err != nil {
			log.Printf("Could not find player for SteamID %s: %v", steamID, err)
			return err
		}

		kdr, err := fw.db.GetQueries().GetKillsByPlayer(ctx, generated.GetKillsByPlayerParams{
			KillerID: &player.ID,
			ServerID: serverDBID,
		})
		killCount := len(kdr)
		deaths, err := fw.db.GetQueries().GetDeathsByPlayer(ctx, generated.GetDeathsByPlayerParams{
			VictimName: &player.Name,
			ServerID:  serverDBID,
		})
		if err != nil {
			log.Printf("Failed to get deaths for player %s: %v", player.Name, err)
			return err
		}
		deathCount := len(deaths)

	// get kdr
		getKDR := func(kills, deaths int) float64 {
			if deaths == 0 {
				if kills == 0 {
					return 0.0
				}
				return float64(kills)
			}
			return float64(kills) / float64(deaths)
		}

		msg := fmt.Sprintf("%s: kills:%d, deaths:%d, KDR: %.2f", playerName, killCount, deathCount, getKDR(killCount, deathCount)) // Replace with real logic
		_, err = client.Send("say " + msg)
		if err != nil {
			log.Printf("Failed to send RCON say: %v", err)
		}
	}
	if command == "!guns" {
		msg := fmt.Sprintf("%s: Your best guns are ...", playerName) // Replace with real logic
		_, err := client.Send("say " + msg)
		if err != nil {
			log.Printf("Failed to send RCON say: %v", err)
		}
	}
	if command == "!stats" {
		msg := fmt.Sprintf("%s: Your stats are ...", playerName) // Replace with real logic
		_, err := client.Send("say " + msg)
		if err != nil {
			log.Printf("Failed to send RCON say: %v", err)
		}
	}
	if command == "!top" {
		msg := fmt.Sprintf("%s: Top players are ...", playerName) // Replace with real logic
		_, err := client.Send("say " + msg)
		if err != nil {
			log.Printf("Failed to send RCON say: %v", err)
		}
	}
	// Add more command handling as needed
	return nil
}
