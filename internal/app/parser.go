package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// LogParser handles parsing log lines and writing directly to database
type LogParser struct {
	patterns *logPatterns
	pbApp    core.App
}

// logPatterns contains compiled regex patterns for log parsing
type logPatterns struct {
	PlayerKill       *regexp.Regexp
	PlayerJoin       *regexp.Regexp
	PlayerDisconnect *regexp.Regexp
	RoundStart       *regexp.Regexp
	RoundEnd         *regexp.Regexp
	GameOver         *regexp.Regexp
	MapLoad          *regexp.Regexp
	DifficultyChange *regexp.Regexp
	MapVote          *regexp.Regexp
	ChatCommand      *regexp.Regexp
	RconCommand      *regexp.Regexp
	Timestamp        *regexp.Regexp
}

func newLogPatterns() *logPatterns {
	return &logPatterns{
		// Kill events - always provide consistent capture groups for killer/victim/weapon fields
		// PlayerKill: timestamp, killerSection, victimName, victimSteam, victimTeam, weapon
		PlayerKill: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: (.+?) killed ([^\[]+)\[([^,\]]*), team (\d+)\] with (.+)$`),

		// Player connection events - handles both LogNet and LogEOSAntiCheat formats
		// 1. [timestamp][id]LogNet: Join succeeded: PlayerName
		// 2. [timestamp][id]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (STEAMID) Result: (EOS_Success)
		PlayerJoin: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\](?:LogNet: Join succeeded: ([^\r\n]+)|LogEOSAntiCheat: Display: ServerRegisterClient: Client: \((\d+)\) Result: \(EOS_Success\))`),

		PlayerDisconnect: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId \((\d+)\), Result: \(EOS_Success\)`),

		// Game state events
		RoundStart: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogGameplayEvents: Display: (?:Pre-)?round (\d+) started`),

		RoundEnd: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]Log(?:GameMode|GameplayEvents): Display: Round (?:(\d+) )?O\s*ver: Team (\d+) won \(win reason: (.+)\)`),

		GameOver: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogSession: Display: AINSGameSession::HandleMatchHasEnded`),

		// Map and server events
		MapLoad: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogLoad: LoadMap: /Game/Maps/([^/]+)/[^?]+\?.*Scenario=([^?&]+).*MaxPlayers=(\d+).*Lighting=([^?&]+)`),

		DifficultyChange: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogAI: Warning: AI difficulty set to ([0-9.]+)`),

		MapVote: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogMapVoteManager: Display: .*Vote Options:`),

		// Chat and RCON events
		ChatCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogChat: Display: ([^(]+)\((\d+)\) Global Chat: (!.+)`),

		RconCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogRcon: ([^<]+)<< (.+)`),

		// Utility pattern for timestamp extraction
		Timestamp: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]`),
	}
}

// NewLogParser creates a new log parser with PocketBase app
func NewLogParser(pbApp core.App) *LogParser {
	return &LogParser{
		patterns: newLogPatterns(),
		pbApp:    pbApp,
	}
}

// ParseAndProcess parses a log line and writes to database if it's a recognized event
func (p *LogParser) ParseAndProcess(ctx context.Context, line string, serverID string, logFilePath string) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Extract timestamp first
	timestampMatches := p.patterns.Timestamp.FindStringSubmatch(line)
	if len(timestampMatches) < 2 {
		return nil // Skip lines without proper timestamp
	}

	timestamp, err := parseTimestamp(timestampMatches[1])
	if err != nil {
		return nil // Skip lines with invalid timestamp
	}

	// Try each event type and process immediately
	if p.tryProcessMapLoad(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessKillEvent(ctx, line, timestamp, serverID, logFilePath) {
		return nil
	}

	if p.tryProcessPlayerJoin(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessPlayerDisconnect(ctx, line, timestamp, serverID) {
		return nil
	}

	// Add other event types as needed (round start/end, etc.)
	// For now, we're focusing on the core stat-tracking events

	return nil
}

// tryProcessKillEvent parses and processes kill events
func (p *LogParser) tryProcessKillEvent(ctx context.Context, line string, timestamp time.Time, serverID string, logFilePath string) bool {
	matches := p.patterns.PlayerKill.FindStringSubmatch(line)
	if len(matches) < 7 {
		return false
	}

	killerSection := strings.TrimSpace(matches[2])
	_ = strings.TrimSpace(matches[3]) // victimName - not currently used
	victimSteamID := strings.TrimSpace(matches[4])
	weapon := cleanWeaponName(matches[6])

	// Parse killer section
	killers := parseKillerSection(killerSection)
	if len(killers) == 0 {
		return true // Parsed but no valid killers (AI suicide)
	}

	// Skip if victim is not a bot (we only track bot kills for PvE)
	if victimSteamID != "INVALID" {
		return true // Parsed but ignored (PvP)
	}

	// Get or create the current active match for this server
	activeMatch, err := GetActiveMatch(ctx, p.pbApp, serverID)
	if err != nil {
		// No active match found - search backwards in log for last map load event
		log.Printf("No active match found for kill event on server %s, searching for last map load...", serverID)

		mapName, scenario, mapLoadTime, err := p.findLastMapLoadEvent(logFilePath, timestamp)
		if err != nil {
			log.Printf("Failed to find last map load event: %v, creating match with unknown map", err)
			// Create match with unknown map info
			mode := "Unknown"
			activeMatch, err = CreateMatch(ctx, p.pbApp, serverID, nil, &mode, &timestamp)
		} else {
			log.Printf("Found last map load: %s (%s) at %v", mapName, scenario, mapLoadTime)
			// Create match with proper map info
			activeMatch, err = CreateMatch(ctx, p.pbApp, serverID, &mapName, &scenario, &mapLoadTime)
		}

		if err != nil {
			log.Printf("Failed to create match for kill event: %v", err)
			return true // Can't track without a match
		}
		log.Printf("Created new match %s for server %s", activeMatch.ID, serverID)
	}

	// Process each killer (first gets kill, rest get assists)
	for i, killer := range killers {
		if killer.SteamID == "INVALID" {
			continue
		}

		// Upsert player
		player, err := GetPlayerByExternalID(ctx, p.pbApp, killer.SteamID)
		if err != nil {
			player, err = CreatePlayer(ctx, p.pbApp, killer.SteamID, killer.Name)
			if err != nil {
				log.Printf("Failed to create player %s: %v", killer.Name, err)
				continue
			}
		}

		// Ensure player is in the match (upsert creates row if needed)
		team := int64(killer.Team)
		err = UpsertMatchPlayerStats(ctx, p.pbApp, activeMatch.ID, player.ID, &team, &timestamp)
		if err != nil {
			log.Printf("Failed to upsert player %s into match: %v", killer.Name, err)
			continue
		}

		// Determine if this is a kill or assist
		if i == 0 {
			// Primary killer - increment kills
			err = IncrementMatchPlayerKills(ctx, p.pbApp, activeMatch.ID, player.ID)
			if err != nil {
				log.Printf("Failed to increment kills for %s: %v", killer.Name, err)
			}
		} else {
			// Assisting player - increment assists
			err = IncrementMatchPlayerAssists(ctx, p.pbApp, activeMatch.ID, player.ID)
			if err != nil {
				log.Printf("Failed to increment assists for %s: %v", killer.Name, err)
			}
		}

		// Update weapon stats for this match
		killCount := int64(0)
		assistCount := int64(0)
		if i == 0 {
			killCount = 1
		} else {
			assistCount = 1
		}

		err = UpsertMatchWeaponStats(ctx, p.pbApp, activeMatch.ID, player.ID, weapon, &killCount, &assistCount)
		if err != nil {
			log.Printf("Failed to update weapon stats for %s: %v", weapon, err)
		}
	}

	return true
}

// tryProcessPlayerJoin parses and processes a player join event
func (p *LogParser) tryProcessPlayerJoin(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerJoin.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	playerName := strings.TrimSpace(matches[1])
	steamID := strings.TrimSpace(matches[2])

	if steamID == "INVALID" || steamID == "" {
		return true // Parsed but invalid
	}

	// Upsert player
	_, err := GetPlayerByExternalID(ctx, p.pbApp, steamID)
	if err != nil {
		_, err := CreatePlayer(ctx, p.pbApp, steamID, playerName)
		if err != nil {
			log.Printf("Failed to create player on join: %v", err)
		}
	}

	return true
}

// tryProcessPlayerDisconnect parses and processes a player disconnect event
func (p *LogParser) tryProcessPlayerDisconnect(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerDisconnect.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	steamID := strings.TrimSpace(matches[2])
	if steamID == "INVALID" || steamID == "" {
		return true // Parsed but invalid
	}

	// Note: Player disconnect tracking could be added here if needed
	return true
}

// tryProcessMapLoad parses and processes map load events
// When a new map loads, we check for any unfinished match on this server and force-end it
func (p *LogParser) tryProcessMapLoad(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.MapLoad.FindStringSubmatch(line)
	if len(matches) < 5 {
		return false
	}

	mapName := strings.TrimSpace(matches[2])
	scenario := strings.TrimSpace(matches[3])
	// maxPlayers := strings.TrimSpace(matches[4])
	// lighting := strings.TrimSpace(matches[5])

	log.Printf("Map load detected: %s (scenario: %s) on server %s", mapName, scenario, serverID)

	// Check if there's an active match that never ended (server crash or restart)
	activeMatch, err := GetActiveMatch(ctx, p.pbApp, serverID)
	if err == nil {
		log.Printf("Found unfinished match %s on server %s, force-ending it before new match",
			activeMatch.ID, serverID)

		// Determine the best end time: use last player activity, or fall back to match start time
		endTime := activeMatch.StartTime // Default fallback
		if activeMatch.UpdatedAt != nil {
			// Use the last time the match record was updated (typically from player activity)
			endTime = activeMatch.UpdatedAt
		}

		// Get all players in the match to find the last activity time
		playersInMatch, err := GetAllPlayersInMatch(ctx, p.pbApp, activeMatch.ID)
		if err == nil && len(playersInMatch) > 0 {
			// Find the most recent player activity
			for _, player := range playersInMatch {
				if player.UpdatedAt != nil && player.UpdatedAt.After(*endTime) {
					endTime = player.UpdatedAt
				}
			}
			log.Printf("Using last player activity as end time: %v", endTime)
		}

		// Force-end the old match using the last known activity time
		err = EndMatch(ctx, p.pbApp, activeMatch.ID, endTime, nil)
		if err != nil {
			log.Printf("Failed to force-end match %s: %v", activeMatch.ID, err)
		}

		// Disconnect all players still marked as connected in that match
		err = DisconnectAllPlayersInMatch(ctx, p.pbApp, activeMatch.ID, endTime)
		if err != nil {
			log.Printf("Failed to disconnect players from match %s: %v", activeMatch.ID, err)
		} else {
			log.Printf("Successfully closed stale match %s", activeMatch.ID)
		}
	}

	// TODO: Create new match record here when match tracking is fully implemented
	// For now, we're just cleaning up stale matches

	return true
}

// Helper types and functions

type killer struct {
	Name    string
	SteamID string
	Team    int
}

func parseKillerSection(killerSection string) []killer {
	var killers []killer

	if strings.TrimSpace(killerSection) == "?" {
		return killers
	}

	playerParts := strings.Split(killerSection, " + ")
	playerRegex := regexp.MustCompile(`^(.+?)\[([^,\]]*), team (\d+)\]$`)

	for _, playerPart := range playerParts {
		playerPart = strings.TrimSpace(playerPart)
		matches := playerRegex.FindStringSubmatch(playerPart)

		if len(matches) == 4 {
			team, _ := strconv.Atoi(matches[3])
			killers = append(killers, killer{
				Name:    strings.TrimSpace(matches[1]),
				SteamID: strings.TrimSpace(matches[2]),
				Team:    team,
			})
		}
	}

	return killers
}

func cleanWeaponName(weapon string) string {
	weapon = strings.TrimSpace(weapon)
	weapon = strings.TrimPrefix(weapon, "BP_")
	weapon = strings.TrimSuffix(weapon, "_C")
	weapon = strings.ReplaceAll(weapon, "_", " ")
	return weapon
}

func parseTimestamp(ts string) (time.Time, error) {
	// Format: 2025.10.04-15.23.38:790
	// Handle variable length milliseconds by using a custom parsing approach

	// Split on the colon to separate the milliseconds
	colonIdx := strings.LastIndex(ts, ":")
	if colonIdx == -1 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
	}

	dateTimePart := ts[:colonIdx]
	msPart := ts[colonIdx+1:]

	// Parse the date/time part
	dt, err := time.Parse("2006.01.02-15.04.05", dateTimePart)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime part: %w", err)
	}

	// Parse milliseconds and add to the time
	ms, err := strconv.Atoi(msPart)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse milliseconds: %w", err)
	}

	// Add milliseconds to the parsed time
	return dt.Add(time.Duration(ms) * time.Millisecond), nil
}

// findLastMapLoadEvent searches backwards through a log file to find the most recent map load event
// Returns map name, scenario, and timestamp, or error if not found
func (p *LogParser) findLastMapLoadEvent(logFilePath string, beforeTime time.Time) (mapName, scenario string, timestamp time.Time, err error) {
	file, err := os.Open(logFilePath)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read file in reverse to find the last map load event before the given time
	// For simplicity, we'll read all lines and process from end to start
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to read log file: %w", err)
	}

	// Search backwards through lines
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Try to match map load pattern
		matches := p.patterns.MapLoad.FindStringSubmatch(line)
		if len(matches) < 5 {
			continue
		}

		// Parse timestamp
		ts, err := parseTimestamp(matches[1])
		if err != nil {
			continue
		}

		// Only consider events before the target time
		if ts.After(beforeTime) {
			continue
		}

		// Found it!
		mapName = strings.TrimSpace(matches[2])
		scenario = strings.TrimSpace(matches[3])
		timestamp = ts
		return mapName, scenario, timestamp, nil
	}

	return "", "", time.Time{}, fmt.Errorf("no map load event found before %v", beforeTime)
}
