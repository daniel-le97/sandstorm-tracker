package parser

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

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/ml"

	"github.com/pocketbase/pocketbase/core"
)

// LogParser handles parsing log lines and writing directly to database
type LogParser struct {
	patterns    *logPatterns
	pbApp       core.App
	chatHandler *ChatCommandHandler
}

// logPatterns contains compiled regex patterns for log parsing
type logPatterns struct {
	CommandLine      *regexp.Regexp
	PlayerKill       *regexp.Regexp
	PlayerJoin       *regexp.Regexp
	PlayerDisconnect *regexp.Regexp
	RoundStart       *regexp.Regexp
	RoundEnd         *regexp.Regexp
	GameOver         *regexp.Regexp
	MapLoad          *regexp.Regexp
	// DifficultyChange   *regexp.Regexp // Not currently used
	MapVote            *regexp.Regexp
	ChatCommand        *regexp.Regexp
	RconCommand        *regexp.Regexp
	ObjectiveDestroyed *regexp.Regexp
	ObjectiveCaptured  *regexp.Regexp
	Timestamp          *regexp.Regexp
}

func newLogPatterns() *logPatterns {
	return &logPatterns{
		CommandLine: regexp.MustCompile(`LogInit: Command Line:\s+(\w+)\?Scenario=([^?]+)\?MaxPlayers=(\d+)\?Game=([^?]+)\?Lighting=(\w+).*?-Hostname="([^"]+)"`),
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

		// DifficultyChange: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogAI: Warning: AI difficulty set to ([0-9.]+)`), // Not currently used

		MapVote: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogMapVoteManager: Display: .*Vote Options:`),

		// Chat and RCON events
		ChatCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogChat: Display: ([^(]+)\((\d+)\) Global Chat: (!.+)`),

		RconCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogRcon: ([^<]+)<< (.+)`),

		// Objective events
		// ObjectiveDestroyed: timestamp, objectiveNum, owningTeam, destroyingTeam, playerSection
		ObjectiveDestroyed: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogGameplayEvents: Display: Objective (\d+) owned by team (\d+) was destroyed for team (\d+) by (.+)\.`),

		// ObjectiveCaptured: timestamp, objectiveNum, capturingTeam, losingTeam, playerSection
		ObjectiveCaptured: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogGameplayEvents: Display: Objective (\d+) was captured for team (\d+) from team (\d+) by (.+)\.`),

		// Utility pattern for timestamp extraction
		Timestamp: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]`),
	}
}

// NewLogParser creates a new log parser with PocketBase app
func NewLogParser(pbApp core.App) *LogParser {
	return &LogParser{
		patterns:    newLogPatterns(),
		pbApp:       pbApp,
		chatHandler: nil, // Set later via SetChatHandler
	}
}

// SetChatHandler sets the chat command handler (must be called after parser creation)
func (p *LogParser) SetChatHandler(rconSender RconSender) {
	p.chatHandler = NewChatCommandHandler(p, rconSender)
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
	// NOTE: Check objectives BEFORE kills to prevent objectives from being counted as kills
	if p.tryProcessMapLoad(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessObjectiveDestroyed(ctx, line, timestamp, serverID, logFilePath) {
		return nil
	}

	if p.tryProcessObjectiveCaptured(ctx, line, timestamp, serverID, logFilePath) {
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

	if p.tryProcessChatCommand(ctx, line, timestamp, serverID, p.chatHandler) {
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
	victimName := strings.TrimSpace(matches[3])
	victimSteamID := strings.TrimSpace(matches[4])
	victimTeamStr := strings.TrimSpace(matches[5])
	weapon := cleanWeaponName(matches[6])

	victimTeam := -1
	if t, err := strconv.Atoi(victimTeamStr); err == nil {
		victimTeam = t
	}

	// Parse killer section
	killers := parseKillerSection(killerSection)
	if len(killers) == 0 {
		return true // Parsed but no valid killers (AI suicide)
	}

	// Determine if this is PvP (player victim) or PvE (bot victim)
	isPvP := victimSteamID != "INVALID"

	// Get or create the current active match for this server
	activeMatch, err := p.getOrCreateMatchForEvent(ctx, serverID, logFilePath, timestamp, "kill")
	if err != nil {
		log.Printf("Failed to get/create match for kill event: %v", err)
		return true // Can't track without a match
	}

	// Process each killer (first gets kill, rest get assists)
	for i, killer := range killers {
		if killer.SteamID == "INVALID" {
			continue
		}

		// Get or create player with up-to-date Steam ID and name
		player, err := p.getOrCreatePlayerForEvent(ctx, killer.SteamID, killer.Name)
		if err != nil {
			log.Printf("Failed to get/create player %s: %v", killer.Name, err)
			continue
		}

		// Ensure player is in the match (upsert creates row if needed)
		team := int64(killer.Team)
		err = database.UpsertMatchPlayerStats(ctx, p.pbApp, activeMatch.ID, player.ID, &team, &timestamp)
		if err != nil {
			log.Printf("Failed to upsert player %s into match: %v", killer.Name, err)
			continue
		}

		// Only process for primary killer (index 0)
		if i == 0 {
			// Check if this is friendly fire (same team) or suicide
			isSuicide := killer.SteamID == victimSteamID
			isFriendlyFire := !isSuicide && isPvP && killer.Team == victimTeam

			if isFriendlyFire {
				// Friendly fire - increment friendly_fire_kills for killer
				err = database.IncrementMatchPlayerStat(ctx, p.pbApp, activeMatch.ID, player.ID, "friendly_fire_kills")
				if err != nil {
					log.Printf("Failed to increment friendly fire for %s: %v", killer.Name, err)
				}

				// Classify player behavior after recording friendly fire incident
				classifier := ml.NewDefaultClassifier()
				prediction, err := classifier.ClassifyPlayer(ctx, p.pbApp, player.ID)
				if err != nil {
					log.Printf("Failed to classify player %s: %v", killer.Name, err)
				} else {
					// Log the classification result
					log.Printf("[FF ANALYSIS] Player: %s | Classification: %s | Confidence: %.0f%%",
						killer.Name, prediction.Classification, prediction.Confidence*100)
					log.Printf("[FF ANALYSIS] Reasoning:")
					for _, reason := range prediction.Reasoning {
						log.Printf("[FF ANALYSIS]   - %s", reason)
					}

					// Flag high-risk players (but don't take action)
					if prediction.Classification == "likely_intentional" && prediction.Confidence > 0.80 {
						log.Printf("⚠️  [HIGH RISK] Player %s flagged as likely intentional team killer (%.0f%% confidence)",
							killer.Name, prediction.Confidence*100)
					} else if prediction.Classification == "possibly_intentional" {
						log.Printf("⚠️  [MONITOR] Player %s may be intentionally team killing (%.0f%% confidence)",
							killer.Name, prediction.Confidence*100)
					}
				}
			} else if !isSuicide {
				// Normal kill (PvP enemy or PvE bot) - increment kills
				err = database.IncrementMatchPlayerStat(ctx, p.pbApp, activeMatch.ID, player.ID, "kills")
				if err != nil {
					log.Printf("Failed to increment kills for %s: %v", killer.Name, err)
				}
			}
			// Note: Suicides don't increment kills or friendly fire for the killer
		} else {
			// Assisting player - increment assists
			err = database.IncrementMatchPlayerStat(ctx, p.pbApp, activeMatch.ID, player.ID, "assists")
			if err != nil {
				log.Printf("Failed to increment assists for %s: %v", killer.Name, err)
			}
		}

		// Update weapon stats for this match (only for non-friendly-fire, non-suicide)
		if i == 0 {
			isSuicide := killer.SteamID == victimSteamID
			isFriendlyFire := !isSuicide && isPvP && killer.Team == victimTeam

			if !isFriendlyFire && !isSuicide {
				killCount := int64(1)
				assistCount := int64(0)
				err = database.UpsertMatchWeaponStats(ctx, p.pbApp, activeMatch.ID, player.ID, weapon, &killCount, &assistCount)
				if err != nil {
					log.Printf("Failed to update weapon stats for %s: %v", weapon, err)
				}
			}
		} else {
			killCount := int64(0)
			assistCount := int64(1)
			err = database.UpsertMatchWeaponStats(ctx, p.pbApp, activeMatch.ID, player.ID, weapon, &killCount, &assistCount)
			if err != nil {
				log.Printf("Failed to update weapon stats for %s: %v", weapon, err)
			}
		}
	}

	// Process victim deaths (if player, not bot)
	if isPvP {
		// Try to find victim player by Steam ID first, then by name
		victimPlayer, err := database.GetPlayerByExternalID(ctx, p.pbApp, victimSteamID)
		if err != nil {
			// Not found by Steam ID, try by name
			victimPlayer, err = database.GetPlayerByName(ctx, p.pbApp, victimName)
			if err != nil {
				// Player doesn't exist at all - create with both name and Steam ID
				victimPlayer, err = database.CreatePlayer(ctx, p.pbApp, victimSteamID, victimName)
				if err != nil {
					log.Printf("Failed to create victim player %s: %v", victimName, err)
					return true
				}
			} else {
				// Found by name but missing Steam ID - update it
				err = database.UpdatePlayerExternalID(ctx, p.pbApp, victimPlayer, victimSteamID)
				if err != nil {
					log.Printf("Failed to update victim player external_id for %s: %v", victimName, err)
				}
			}
		} else {
			// Found by Steam ID - update name if it has changed
			err = database.UpdatePlayerName(ctx, p.pbApp, victimPlayer, victimName)
			if err != nil {
				log.Printf("Failed to update victim player name for %s: %v", victimName, err)
			}
		}

		// Ensure victim is in the match
		victimTeam64 := int64(victimTeam)
		err = database.UpsertMatchPlayerStats(ctx, p.pbApp, activeMatch.ID, victimPlayer.ID, &victimTeam64, &timestamp)
		if err != nil {
			log.Printf("Failed to upsert victim player %s into match: %v", victimName, err)
			return true
		}

		// Increment victim's deaths
		err = database.IncrementMatchPlayerStat(ctx, p.pbApp, activeMatch.ID, victimPlayer.ID, "deaths")
		if err != nil {
			log.Printf("Failed to increment deaths for victim %s: %v", victimName, err)
		}
	}

	return true
}

// tryProcessPlayerJoin parses and processes a player join event
// Note: Player joins generate TWO log lines:
// 1. LogNet: Join succeeded: PlayerName (comes first)
// 2. LogEOSAntiCheat: ServerRegisterClient: Client: (STEAMID) (comes second)
// We create the player from the LogNet event with just the name,
// and update the external_id when we see kill events with their Steam ID
func (p *LogParser) tryProcessPlayerJoin(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerJoin.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	// PlayerJoin regex has alternation:
	// Group 1: timestamp (always)
	// Group 2: player name (LogNet branch) OR empty
	// Group 3: Steam ID (EOS branch) OR empty

	if matches[2] != "" {
		// LogNet branch: "Join succeeded: PlayerName"
		playerName := strings.TrimSpace(matches[2])

		// Check if player already exists by name
		_, err := database.GetPlayerByName(ctx, p.pbApp, playerName)
		if err != nil {
			// Player doesn't exist - create with name only (external_id will be empty)
			_, err := database.CreatePlayer(ctx, p.pbApp, "", playerName)
			if err != nil {
				log.Printf("Failed to create player on join: %v", err)
			}
		}

		return true
	} else if len(matches) > 3 && matches[3] != "" {
		// EOS branch: "ServerRegisterClient: Client: (STEAMID)"
		// We don't need to do anything here - the player was already created from LogNet event
		// Their external_id will be populated when we see them in a kill event
		return true
	}

	return false
} // tryProcessPlayerDisconnect parses and processes a player disconnect event
func (p *LogParser) tryProcessPlayerDisconnect(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerDisconnect.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	steamID := strings.TrimSpace(matches[2])
	if steamID == "INVALID" || steamID == "" {
		return true // Parsed but invalid
	}

	// Find the active match for this server
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
	if err != nil || activeMatch == nil {
		// No active match, nothing to do
		return true
	}

	// Find the player by Steam ID
	player, err := database.GetPlayerByExternalID(ctx, p.pbApp, steamID)
	if err != nil || player == nil {
		// Player not found in database
		return true
	}

	// Mark player as disconnected from the match
	err = database.DisconnectPlayerFromMatch(ctx, p.pbApp, activeMatch.ID, player.ID, &timestamp)
	if err != nil {
		log.Printf("Failed to disconnect player %s from match %s: %v", player.Name, activeMatch.ID, err)
	} else {
		log.Printf("Player %s disconnected from match %s", player.Name, activeMatch.ID)
	}

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

	// Extract player team from scenario (e.g., "Scenario_Ministry_Checkpoint_Security" -> "Security")
	var playerTeam *string
	if strings.Contains(scenario, "_Security") {
		team := "Security"
		playerTeam = &team
	} else if strings.Contains(scenario, "_Insurgents") {
		team := "Insurgents"
		playerTeam = &team
	}

	log.Printf("Map load detected: %s (scenario: %s) on server %s", mapName, scenario, serverID)

	// Check if there's an active match that never ended (server crash or restart)
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
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
		playersInMatch, err := database.GetAllPlayersInMatch(ctx, p.pbApp, activeMatch.ID)
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
		err = database.EndMatch(ctx, p.pbApp, activeMatch.ID, endTime, nil)
		if err != nil {
			log.Printf("Failed to force-end match %s: %v", activeMatch.ID, err)
		}

		// Disconnect all players still marked as connected in that match
		err = database.DisconnectAllPlayersInMatch(ctx, p.pbApp, activeMatch.ID, endTime)
		if err != nil {
			log.Printf("Failed to disconnect players from match %s: %v", activeMatch.ID, err)
		}

		log.Printf("Successfully closed stale match %s", activeMatch.ID)

		// Delete the match if it has no stats (was likely never used)
		err = database.DeleteMatchIfEmpty(ctx, p.pbApp, activeMatch.ID)
		if err != nil {
			log.Printf("Failed to check/delete empty match %s: %v", activeMatch.ID, err)
		}
	}

	// Create new match record
	_, err = database.CreateMatch(ctx, p.pbApp, serverID, &mapName, &scenario, &timestamp, playerTeam)
	if err != nil {
		log.Printf("Failed to create match for map %s on server %s: %v", mapName, serverID, err)
	}

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
	// Try full format first: Name[SteamID, team X]
	playerRegexFull := regexp.MustCompile(`^(.+?)\[([^,\]]*), team (\d+)\]$`)
	// Fallback format for objective events: Name[SteamID]
	playerRegexSimple := regexp.MustCompile(`^(.+?)\[([^\]]+)\]$`)

	for _, playerPart := range playerParts {
		playerPart = strings.TrimSpace(playerPart)
		matches := playerRegexFull.FindStringSubmatch(playerPart)

		if len(matches) == 4 {
			team, _ := strconv.Atoi(matches[3])
			killers = append(killers, killer{
				Name:    strings.TrimSpace(matches[1]),
				SteamID: strings.TrimSpace(matches[2]),
				Team:    team,
			})
		} else {
			// Try simple format
			matches = playerRegexSimple.FindStringSubmatch(playerPart)
			if len(matches) == 3 {
				killers = append(killers, killer{
					Name:    strings.TrimSpace(matches[1]),
					SteamID: strings.TrimSpace(matches[2]),
					Team:    -1, // Team is provided elsewhere in objective events
				})
			}
		}
	}

	return killers
}

func cleanWeaponName(weapon string) string {
	weapon = strings.TrimSpace(weapon)
	weapon = strings.TrimPrefix(weapon, "BP_")

	// Remove numeric ID suffix first (e.g., "_2147480339")
	// Find the last underscore followed by only digits
	lastUnderscore := strings.LastIndex(weapon, "_")
	if lastUnderscore != -1 {
		// Check if everything after the last underscore is digits
		potentialID := weapon[lastUnderscore+1:]
		isNumeric := true
		for _, ch := range potentialID {
			if ch < '0' || ch > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric && len(potentialID) > 0 {
			weapon = weapon[:lastUnderscore]
		}
	}

	// Now remove _C suffix
	weapon = strings.TrimSuffix(weapon, "_C")

	// Remove common prefixes like "Firearm_", "Weapon_", "Melee_", "Projectile_"
	weapon = strings.TrimPrefix(weapon, "Firearm_")
	weapon = strings.TrimPrefix(weapon, "Weapon_")
	weapon = strings.TrimPrefix(weapon, "Melee_")
	weapon = strings.TrimPrefix(weapon, "Projectile_")

	weapon = strings.ReplaceAll(weapon, "_", " ")

	// Standardize ODCheckpoint variants (ODCheckpoint A, ODCheckpoint B -> ODCheckpoint)
	if strings.HasPrefix(weapon, "ODCheckpoint ") {
		weapon = "ODCheckpoint"
	}

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

// tryProcessObjectiveDestroyed parses and processes objective destroyed events
func (p *LogParser) tryProcessObjectiveDestroyed(ctx context.Context, line string, timestamp time.Time, serverID string, logFilePath string) bool {
	matches := p.patterns.ObjectiveDestroyed.FindStringSubmatch(line)
	if len(matches) < 6 {
		return false
	}

	objectiveNum := strings.TrimSpace(matches[2])
	owningTeam := strings.TrimSpace(matches[3])
	destroyingTeam := strings.TrimSpace(matches[4])
	playerSection := strings.TrimSpace(matches[5])

	log.Printf("Objective %s destroyed: owned by team %s, destroyed for team %s by %s",
		objectiveNum, owningTeam, destroyingTeam, playerSection)

	// Parse player section (can have multiple players)
	players := parseKillerSection(playerSection)
	if len(players) == 0 {
		return true // Parsed but no valid players
	}

	// Get or create the current active match for this server
	activeMatch, err := p.getOrCreateMatchForEvent(ctx, serverID, logFilePath, timestamp, "objective destroyed")
	if err != nil {
		log.Printf("Failed to get/create match for objective destroyed event: %v", err)
		return true // Can't track without a match
	}

	// Process each player (first gets credit, rest get assists)
	for i, player := range players {
		if player.SteamID == "INVALID" || player.SteamID == "" {
			continue
		}

		// Get or create player with up-to-date Steam ID and name
		playerRecord, err := p.getOrCreatePlayerForEvent(ctx, player.SteamID, player.Name)
		if err != nil {
			log.Printf("Failed to get/create player %s: %v", player.Name, err)
			continue
		}

		// Ensure player is in the match (upsert creates row if needed)
		team, _ := strconv.ParseInt(destroyingTeam, 10, 64)
		err = database.UpsertMatchPlayerStats(ctx, p.pbApp, activeMatch.ID, playerRecord.ID, &team, &timestamp)
		if err != nil {
			log.Printf("Failed to upsert player %s into match: %v", player.Name, err)
			continue
		}

		// First player gets the objective destroy credit, others get assists
		if i == 0 {
			// Track as objective destroyed in match_player_stats
			err = database.IncrementMatchPlayerStat(ctx, p.pbApp, activeMatch.ID, playerRecord.ID, "objectives_destroyed")
			if err != nil {
				log.Printf("Failed to increment objective destroyed for player %s: %v", player.Name, err)
			}
		} else {
			// Could track assists here if desired
			log.Printf("Player %s assisted in destroying objective %s", player.Name, objectiveNum)
		}
	}

	return true
}

// tryProcessObjectiveCaptured parses and processes objective captured events
func (p *LogParser) tryProcessObjectiveCaptured(ctx context.Context, line string, timestamp time.Time, serverID string, logFilePath string) bool {
	matches := p.patterns.ObjectiveCaptured.FindStringSubmatch(line)
	if len(matches) < 6 {
		return false
	}

	objectiveNum := strings.TrimSpace(matches[2])
	capturingTeam := strings.TrimSpace(matches[3])
	losingTeam := strings.TrimSpace(matches[4])
	playerSection := strings.TrimSpace(matches[5])

	log.Printf("Objective %s captured: team %s captured from team %s by %s",
		objectiveNum, capturingTeam, losingTeam, playerSection)

	// Parse player section (can have multiple players)
	players := parseKillerSection(playerSection)
	if len(players) == 0 {
		return true // Parsed but no valid players
	}

	// Get or create the current active match for this server
	activeMatch, err := p.getOrCreateMatchForEvent(ctx, serverID, logFilePath, timestamp, "objective captured")
	if err != nil {
		log.Printf("Failed to get/create match for objective captured event: %v", err)
		return true // Can't track without a match
	}

	// Process each player (first gets credit, rest get assists)
	for i, player := range players {
		if player.SteamID == "INVALID" || player.SteamID == "" {
			continue
		}

		// Get or create player with up-to-date Steam ID and name
		playerRecord, err := p.getOrCreatePlayerForEvent(ctx, player.SteamID, player.Name)
		if err != nil {
			log.Printf("Failed to get/create player %s: %v", player.Name, err)
			continue
		}

		// Ensure player is in the match (upsert creates row if needed)
		team, _ := strconv.ParseInt(capturingTeam, 10, 64)
		err = database.UpsertMatchPlayerStats(ctx, p.pbApp, activeMatch.ID, playerRecord.ID, &team, &timestamp)
		if err != nil {
			log.Printf("Failed to upsert player %s into match: %v", player.Name, err)
			continue
		}

		// First player gets the objective capture credit, others get assists
		if i == 0 {
			// Track as objective captured in match_player_stats
			err = database.IncrementMatchPlayerStat(ctx, p.pbApp, activeMatch.ID, playerRecord.ID, "objectives_captured")
			if err != nil {
				log.Printf("Failed to increment objective captured for player %s: %v", player.Name, err)
			}
		} else {
			// Could track assists here if desired
			log.Printf("Player %s assisted in capturing objective %s", player.Name, objectiveNum)
		}
	}

	return true
}

// getOrCreateMatchForEvent retrieves the active match for a server, or creates one if none exists
// by searching backwards in the log file for the last map load event.
func (p *LogParser) getOrCreateMatchForEvent(ctx context.Context, serverID, logFilePath string, timestamp time.Time, eventName string) (*database.Match, error) {
	// Try to get active match first
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
	if err == nil && activeMatch != nil {
		return activeMatch, nil
	}

	// No active match found - search backwards in log for last map load event
	log.Printf("No active match found for %s event on server %s, searching for last map load...", eventName, serverID)

	mapName, scenario, mapLoadTime, err := p.findLastMapLoadEvent(logFilePath, timestamp)
	if err != nil {
		log.Printf("Failed to find last map load event: %v, creating match with unknown map", err)
		// Create match with unknown map info
		mode := "Unknown"
		activeMatch, err = database.CreateMatch(ctx, p.pbApp, serverID, nil, &mode, &timestamp)
	} else {
		log.Printf("Found last map load: %s (%s) at %v", mapName, scenario, mapLoadTime)
		// Create match with proper map info
		activeMatch, err = database.CreateMatch(ctx, p.pbApp, serverID, &mapName, &scenario, &mapLoadTime)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create match for %s event: %w", eventName, err)
	}

	log.Printf("Created new match %s for server %s", activeMatch.ID, serverID)
	return activeMatch, nil
}

// getOrCreatePlayerForEvent retrieves a player by Steam ID or name, creating or updating as needed.
// Returns the player record, ensuring both Steam ID and name are set correctly.
func (p *LogParser) getOrCreatePlayerForEvent(ctx context.Context, steamID, name string) (*database.Player, error) {
	// Try to find player by Steam ID first
	player, err := database.GetPlayerByExternalID(ctx, p.pbApp, steamID)
	if err != nil {
		// Not found by Steam ID, try by name
		player, err = database.GetPlayerByName(ctx, p.pbApp, name)
		if err != nil {
			// Player doesn't exist at all - create with both name and Steam ID
			player, err = database.CreatePlayer(ctx, p.pbApp, steamID, name)
			if err != nil {
				return nil, fmt.Errorf("failed to create player %s: %w", name, err)
			}
			return player, nil
		}

		// Found by name but missing Steam ID - update it
		err = database.UpdatePlayerExternalID(ctx, p.pbApp, player, steamID)
		if err != nil {
			log.Printf("Failed to update player external_id for %s: %v", name, err)
		}
		return player, nil
	}

	// Found by Steam ID - update name if it has changed
	err = database.UpdatePlayerName(ctx, p.pbApp, player, name)
	if err != nil {
		log.Printf("Failed to update player name for %s: %v", name, err)
	}

	return player, nil
}
