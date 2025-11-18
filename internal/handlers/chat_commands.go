package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/events"

	"github.com/pocketbase/pocketbase/core"
)

// HandleChatCommand processes a chat command event with all functionality inline
func HandleChatCommand(rconSender func(string, string) (string, error)) func(e *core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		ctx := context.Background()

		// Extract typed data from event
		var data events.ChatCommandData
		if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
			log.Printf("[HANDLER] Failed to parse chat command event data: %v", err)
			return e.Next()
		}

		// Skip command processing during catchup mode (avoid RCON spam)
		if data.IsCatchup {
			log.Printf("[CHAT] Skipping command processing (catchup mode): %s", data.Command)
			return e.Next()
		}

		// Get server external ID
		serverRecordID := e.Record.GetString("server")
		serverRecord, err := e.App.FindRecordById("servers", serverRecordID)
		if err != nil {
			log.Printf("[CHAT] Failed to get server record: %v", err)
			return e.Next()
		}
		serverID := serverRecord.GetString("external_id")

		log.Printf("[CHAT] %s (%s): %s", data.PlayerName, data.SteamID, data.Command)

		if rconSender == nil {
			log.Printf("[CHAT] No RCON sender configured, skipping command")
			return e.Next()
		}

		// Get or create player for !kdr, !stats, !guns commands
		var player *database.Player
		if command := strings.ToLower(data.Command); command == "!kdr" || command == "!stats" || command == "!guns" {
			var err error
			player, err = database.GetOrCreatePlayerBySteamID(ctx, e.App, data.SteamID, data.PlayerName)
			if err != nil {
				log.Printf("[CHAT] Failed to get player: %v", err)
				return e.Next()
			}
		}

		// Handle different commands
		switch strings.ToLower(data.Command) {
		case "!kdr":
			// Show K/D ratio
			kills, deaths, err := database.GetPlayerTotalKD(ctx, e.App, player.ID)
			if err != nil {
				log.Printf("[CHAT] Failed to get player K/D: %v", err)
				return e.Next()
			}

			kdr := 0.0
			if deaths > 0 {
				kdr = float64(kills) / float64(deaths)
			} else if kills > 0 {
				kdr = float64(kills)
			}

			message := fmt.Sprintf("%s: %d kills, %d deaths, K/D: %.2f", data.PlayerName, kills, deaths, kdr)
			sendRconSay(rconSender, serverID, message)

		case "!stats":
			// Show total stats and ranking
			stats, rank, totalPlayers, err := database.GetPlayerStatsAndRank(ctx, e.App, player.ID)
			if err != nil {
				log.Printf("[CHAT] Failed to get player stats/rank: %v", err)
				return e.Next()
			}

			scorePerMin := 0.0
			if stats != nil && stats.TotalDurationSeconds > 0 {
				scorePerMin = float64(stats.TotalScore) / (float64(stats.TotalDurationSeconds) / 60.0)
			}

			durationHours := stats.TotalDurationSeconds / 3600
			durationMins := (stats.TotalDurationSeconds % 3600) / 60

			message := fmt.Sprintf("%s: Score: %d, Time: %dh%dm, Score/Min: %.1f, Rank: #%d/%d",
				player.Name, stats.TotalScore, durationHours, durationMins, scorePerMin, rank, totalPlayers)
			sendRconSay(rconSender, serverID, message)

		case "!top":
			// Show top 3 players by score/min
			topPlayers, err := database.GetTopPlayersByScorePerMin(ctx, e.App, 3)
			if err != nil {
				log.Printf("[CHAT] Failed to get top players: %v", err)
				return e.Next()
			}

			if len(topPlayers) == 0 {
				sendRconSay(rconSender, serverID, "No stats available yet!")
				return e.Next()
			}

			message := "Top 3 Players by Score/Min:"
			for i, topPlayer := range topPlayers {
				message += fmt.Sprintf(" | #%d: %s - %.1f score/min", i+1, topPlayer.Name, topPlayer.ScorePerMin)
			}
			sendRconSay(rconSender, serverID, message)

		case "!guns", "!weapons":
			// Show top 3 weapons
			topWeapons, err := database.GetTopWeapons(ctx, e.App, player.ID, 3)
			if err != nil {
				log.Printf("[CHAT] Failed to get top weapons: %v", err)
				return e.Next()
			}

			if len(topWeapons) == 0 {
				sendRconSay(rconSender, serverID, fmt.Sprintf("%s: No weapon stats available yet!", data.PlayerName))
				return e.Next()
			}

			weaponList := ""
			for i, weapon := range topWeapons {
				if i > 0 {
					weaponList += ", "
				}
				weaponList += fmt.Sprintf("#%d: %s (%d)", i+1, weapon.Name, weapon.Kills)
			}
			message := fmt.Sprintf("%s's Top Weapons: %s", data.PlayerName, weaponList)
			sendRconSay(rconSender, serverID, message)

		default:
			// Unknown/unsupported command - log it but don't respond
			log.Printf("[CHAT] Unsupported command: %s", data.Command)
		}

		return e.Next()
	}
}

// sendRconSay sends a message via RCON say command
func sendRconSay(rconSender func(string, string) (string, error), serverID, message string) {
	command := fmt.Sprintf("say %s", message)
	_, err := rconSender(serverID, command)
	if err != nil {
		log.Printf("[CHAT] Failed to send RCON message: %v", err)
	} else {
		log.Printf("[CHAT] Sent: %s", message)
	}
}
