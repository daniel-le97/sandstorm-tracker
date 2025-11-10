package parser

import (
	"context"
	"fmt"
	"log"
	"sandstorm-tracker/internal/database"
	"strings"
	"time"
)

// ChatCommandHandler handles chat commands like !stats, !kdr, !top, !guns
type ChatCommandHandler struct {
	parser     *LogParser
	rconSender RconSender
}

// RconSender interface for sending RCON commands
type RconSender interface {
	SendRconCommand(serverID string, command string) (string, error)
}

// NewChatCommandHandler creates a new chat command handler
func NewChatCommandHandler(parser *LogParser, rconSender RconSender) *ChatCommandHandler {
	return &ChatCommandHandler{
		parser:     parser,
		rconSender: rconSender,
	}
}

// tryProcessChatCommand parses and processes chat commands
func (p *LogParser) tryProcessChatCommand(ctx context.Context, line string, timestamp time.Time, serverID string, handler *ChatCommandHandler) bool {
	matches := p.patterns.ChatCommand.FindStringSubmatch(line)
	if len(matches) < 5 {
		return false
	}

	playerName := strings.TrimSpace(matches[2])
	steamID := strings.TrimSpace(matches[3])
	command := strings.TrimSpace(matches[4])

	log.Printf("[CHAT] %s (%s): %s", playerName, steamID, command)

	if handler == nil || handler.rconSender == nil {
		log.Printf("[CHAT] No handler configured, skipping command")
		return true
	}

	// Handle different commands
	// Note: Only specific commands are handled. All other commands (like !map, !maplist, etc.)
	// are detected and logged but silently ignored (no RCON response sent).
	switch strings.ToLower(command) {
	case "!kdr":
		handler.handleKDR(ctx, serverID, steamID, playerName)
	case "!stats":
		handler.handleStats(ctx, serverID, steamID, playerName)
	case "!top":
		handler.handleTop(ctx, serverID)
	case "!guns", "!weapons":
		handler.handleGuns(ctx, serverID, steamID, playerName)
	default:
		// Unknown/unsupported command - log it but don't respond
		log.Printf("[CHAT] Unsupported command: %s", command)
	}

	return true
}

// handleKDR shows player's K/D ratio
func (h *ChatCommandHandler) handleKDR(ctx context.Context, serverID, steamID, playerName string) {
	// Get player stats
	player, err := database.GetOrCreatePlayerBySteamID(ctx, h.parser.pbApp, steamID, playerName)
	if err != nil {
		log.Printf("[CHAT] Failed to get player for KDR: %v", err)
		return
	}

	// Get total kills and deaths across all matches
	kills, deaths, err := database.GetPlayerTotalKD(ctx, h.parser.pbApp, player.ID)
	if err != nil {
		log.Printf("[CHAT] Failed to get player K/D: %v", err)
		return
	}

	kdr := 0.0
	if deaths > 0 {
		kdr = float64(kills) / float64(deaths)
	} else if kills > 0 {
		kdr = float64(kills)
	}

	message := fmt.Sprintf("%s: %d kills, %d deaths, K/D: %.2f", playerName, kills, deaths, kdr)
	h.sendRconSay(serverID, message)
}

// handleStats shows player's total stats and ranking
func (h *ChatCommandHandler) handleStats(ctx context.Context, serverID, steamID, playerName string) {
	// Get or create player and fetch aggregated stats + rank in one helper call
	player, stats, rank, totalPlayers, err := database.GetOrCreatePlayerWithStatsAndRank(ctx, h.parser.pbApp, steamID, playerName)
	if err != nil {
		log.Printf("[CHAT] Failed to get player stats/rank: %v", err)
		return
	}

	// Calculate score/min (stats may already contain zero values)
	scorePerMin := 0.0
	if stats != nil && stats.TotalDurationSeconds > 0 {
		scorePerMin = float64(stats.TotalScore) / (float64(stats.TotalDurationSeconds) / 60.0)
	}

	durationHours := stats.TotalDurationSeconds / 3600
	durationMins := (stats.TotalDurationSeconds % 3600) / 60

	// Use the canonical name from the DB record when available
	displayName := playerName
	if player != nil && player.Name != "" {
		displayName = player.Name
	}

	message := fmt.Sprintf("%s: Score: %d, Time: %dh%dm, Score/Min: %.1f, Rank: #%d/%d",
		displayName, stats.TotalScore, durationHours, durationMins, scorePerMin, rank, totalPlayers)
	h.sendRconSay(serverID, message)
}

// handleTop shows top 3 players by score/min
func (h *ChatCommandHandler) handleTop(ctx context.Context, serverID string) {
	topPlayers, err := database.GetTopPlayersByScorePerMin(ctx, h.parser.pbApp, 3)
	if err != nil {
		log.Printf("[CHAT] Failed to get top players: %v", err)
		return
	}

	if len(topPlayers) == 0 {
		h.sendRconSay(serverID, "No stats available yet!")
		return
	}

	h.sendRconSay(serverID, "Top 3 Players by Score/Min:")
	for i, player := range topPlayers {
		message := fmt.Sprintf("#%d: %s - %.1f score/min", i+1, player.Name, player.ScorePerMin)
		h.sendRconSay(serverID, message)
	}
}

// handleGuns shows player's top 3 most used weapons
func (h *ChatCommandHandler) handleGuns(ctx context.Context, serverID, steamID, playerName string) {
	// Get player record by Steam ID
	player, err := database.GetPlayerByExternalID(ctx, h.parser.pbApp, steamID)
	if err != nil {
		log.Printf("[CHAT] Failed to get player %s: %v", steamID, err)
		h.sendRconSay(serverID, fmt.Sprintf("%s: No stats available yet!", playerName))
		return
	}

	topWeapons, err := database.GetTopWeapons(ctx, h.parser.pbApp, player.ID, 3)
	if err != nil {
		log.Printf("[CHAT] Failed to get top weapons for %s: %v", playerName, err)
		return
	}

	if len(topWeapons) == 0 {
		h.sendRconSay(serverID, fmt.Sprintf("%s: No weapon stats available yet!", playerName))
		return
	}

	// Build a single message with all weapons
	weaponList := ""
	for i, weapon := range topWeapons {
		if i > 0 {
			weaponList += ", "
		}
		weaponList += fmt.Sprintf("#%d: %s (%d)", i+1, weapon.Name, weapon.Kills)
	}
	message := fmt.Sprintf("%s's Top Weapons: %s", playerName, weaponList)
	h.sendRconSay(serverID, message)
}

// sendRconSay sends a message via RCON say command
func (h *ChatCommandHandler) sendRconSay(serverID, message string) {
	command := fmt.Sprintf("say %s", message)
	_, err := h.rconSender.SendRconCommand(serverID, command)
	if err != nil {
		log.Printf("[CHAT] Failed to send RCON message: %v", err)
	} else {
		log.Printf("[CHAT] Sent: %s", message)
	}
}
