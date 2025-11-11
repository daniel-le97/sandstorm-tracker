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
	pbApp              core.App
	patterns           *logPatterns
	chatHandler        *ChatCommandHandler
	lastMapTravelTimes map[string]time.Time // Track last map travel time per server to ignore reconnects
}

// logPatterns contains compiled regex patterns for log parsing
type logPatterns struct {
	LogFileOpen      *regexp.Regexp // Log file open timestamp (first line of log)
	CommandLine      *regexp.Regexp
	PlayerKill       *regexp.Regexp
	PlayerLogin      *regexp.Regexp // Login request (earliest connection event)
	PlayerRegister   *regexp.Regexp // ServerRegisterClient (pre-match)
	PlayerJoin       *regexp.Regexp // Join succeeded (in-match)
	PlayerDisconnect *regexp.Regexp
	RoundStart       *regexp.Regexp
	RoundEnd         *regexp.Regexp
	GameOver         *regexp.Regexp
	MapLoad          *regexp.Regexp
	MapTravel        *regexp.Regexp
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
		// Log file open timestamp (first line of every log file)
		// Format: "Log file open, 11/10/25 20:58:31"
		LogFileOpen: regexp.MustCompile(`^Log file open, (\d{2}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`),

		CommandLine: regexp.MustCompile(`LogInit: Command Line:\s+(\w+)\?Scenario=([^?]+)\?MaxPlayers=(\d+)\?Game=([^?]+)\?Lighting=(\w+).*?-Hostname="([^"]+)"`),
		// Kill events - always provide consistent capture groups for killer/victim/weapon fields
		// PlayerKill: timestamp, killerSection, victimName, victimSteam, victimTeam, weapon
		PlayerKill: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: (.+?) killed ([^\[]+)\[([^,\]]*), team (\d+)\] with (.+)$`),

		// Player connection events - three stages:
		// 1. PlayerLogin: [timestamp][id]LogNet: Login request (earliest connection event with name & Steam ID)
		// 2. PlayerRegister: [timestamp][id]LogEOSAntiCheat: ServerRegisterClient (happens after login)
		// 3. PlayerJoin: [timestamp][id]LogNet: Join succeeded (happens when player actually in match)
		PlayerLogin:    regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogNet: Login request:.*\?Name=(.+?) userId: \w+:(\d+) platform: (\w+)`),
		PlayerRegister: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogEOSAntiCheat: Display: ServerRegisterClient: Client: \((\d+)\) Result: \(EOS_Success\)`),
		PlayerJoin:     regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogNet: Join succeeded: ([^\r\n]+)`),

		PlayerDisconnect: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId \((\d+)\), Result: \(EOS_Success\)`),

		// Game state events
		RoundStart: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: (?:Pre-)?round (\d+) started`),

		RoundEnd: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: Round (\d+) O\s*ver: Team (\d+) won \(win reason: (.+)\)`),

		GameOver: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogSession: Display: AINSGameSession::HandleMatchHasEnded`),

		// Map and server events
		MapLoad: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogLoad: LoadMap: /Game/Maps/([^/]+)/[^?]+\?.*Scenario=([^?&]+).*MaxPlayers=(\d+).*Lighting=([^?&]+)`),
		// ProcessServerTravel events when the map changes during runtime
		// Example: [2025.10.21-20.12.42:785][454]LogGameMode: ProcessServerTravel: Town?Scenario=Scenario_Hideout_Skirmish?Game=?
		MapTravel: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameMode: ProcessServerTravel: ([^?]+)\?Scenario=([^?]+)\?Game=([^\s]*)`),

		// DifficultyChange: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogAI: Warning: AI difficulty set to ([0-9.]+)`), // Not currently used

		MapVote: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogMapVoteManager: Display: New Vote Options:`),

		// Chat and RCON events
		ChatCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogChat: Display: ([^(]+)\((\d+)\) Global Chat: (!.+)`),

		RconCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogRcon: ([^<]+)<< (.+)`),

		// Objective events
		// ObjectiveDestroyed: timestamp, objectiveNum, owningTeam, destroyingTeam, playerSection
		ObjectiveDestroyed: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: Objective (\d+) owned by team (\d+) was destroyed for team (\d+) by (.+)\.`),

		// ObjectiveCaptured: timestamp, objectiveNum, capturingTeam, losingTeam, playerSection
		ObjectiveCaptured: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: Objective (\d+) was captured for team (\d+) from team (\d+) by (.+)\.`),

		// Utility pattern for timestamp extraction
		Timestamp: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]`),
	}
}

// endActiveMatchAndCreateNew ends any active match on the server and creates a new match
// This is used when a map changes (either initial load or travel)
func (p *LogParser) endActiveMatchAndCreateNew(ctx context.Context, serverID string, mapName, scenario string, timestamp time.Time, playerTeam *string) error {
	// If there's an active match, force-end it first
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
	if err == nil && activeMatch != nil {
		log.Printf("Found active match %s on server %s, force-ending it before new map",
			activeMatch.ID, serverID)

		// Determine end time (use last player activity if available)
		endTime := activeMatch.StartTime
		if activeMatch.UpdatedAt != nil {
			endTime = activeMatch.UpdatedAt
		}

		playersInMatch, err := database.GetAllPlayersInMatch(ctx, p.pbApp, activeMatch.ID)
		if err == nil && len(playersInMatch) > 0 {
			for _, pl := range playersInMatch {
				if pl.UpdatedAt != nil && pl.UpdatedAt.After(*endTime) {
					endTime = pl.UpdatedAt
				}
			}
			log.Printf("Using last player activity as end time: %v", endTime)
		}

		if err := database.EndMatch(ctx, p.pbApp, activeMatch.ID, endTime, nil); err != nil {
			log.Printf("Failed to force-end match %s: %v", activeMatch.ID, err)
		}

		if err := database.DisconnectAllPlayersInMatch(ctx, p.pbApp, activeMatch.ID, endTime); err != nil {
			log.Printf("Failed to disconnect players from match %s: %v", activeMatch.ID, err)
		}

		log.Printf("Successfully closed match %s", activeMatch.ID)

		if err := database.DeleteMatchIfEmpty(ctx, p.pbApp, activeMatch.ID); err != nil {
			log.Printf("Failed to check/delete empty match %s: %v", activeMatch.ID, err)
		}
	}

	// Create the new match
	_, err = database.CreateMatch(ctx, p.pbApp, serverID, &mapName, &scenario, &timestamp, playerTeam)
	if err != nil {
		return fmt.Errorf("failed to create match for map %s on server %s: %w", mapName, serverID, err)
	}

	log.Printf("Created new match for map %s on server %s", mapName, serverID)
	return nil
}

// tryProcessMapTravel handles in-game map transitions (ProcessServerTravel)
// This occurs when the server changes maps during runtime (after the initial map load)
func (p *LogParser) tryProcessMapTravel(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.MapTravel.FindStringSubmatch(line)
	if len(matches) < 4 {
		return false
	}

	mapName := strings.TrimSpace(matches[2])
	scenario := strings.TrimSpace(matches[3])
	// gameParam := strings.TrimSpace(matches[4]) // currently unused

	log.Printf("Map travel detected: %s (scenario: %s) on server %s", mapName, scenario, serverID)

	// Track this map travel time so we can ignore immediate disconnects/reconnects
	p.lastMapTravelTimes[serverID] = timestamp

	// End active match and create new one
	if err := p.endActiveMatchAndCreateNew(ctx, serverID, mapName, scenario, timestamp, nil); err != nil {
		log.Printf("Failed to transition to new map: %v", err)
	}

	return true
}

// NewLogParser creates a new log parser with PocketBase app
func NewLogParser(pbApp core.App) *LogParser {
	return &LogParser{
		patterns:           newLogPatterns(),
		pbApp:              pbApp,
		chatHandler:        nil, // Set later via SetChatHandler
		lastMapTravelTimes: make(map[string]time.Time),
	}
}

// ExtractLogFileCreationTime reads the first line of a log file and extracts the creation timestamp
// Returns the timestamp or error if not found
func (p *LogParser) ExtractLogFileCreationTime(logFilePath string) (time.Time, error) {
	file, err := os.Open(logFilePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return time.Time{}, fmt.Errorf("log file is empty")
	}

	firstLine := scanner.Text()
	matches := p.patterns.LogFileOpen.FindStringSubmatch(firstLine)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("first line does not match log file open pattern: %s", firstLine)
	}

	// Parse timestamp: "11/10/25 20:58:31"
	timestamp, err := time.Parse("01/02/06 15:04:05", matches[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse log file timestamp: %w", err)
	}

	return timestamp, nil
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

	if p.tryProcessMapTravel(ctx, line, timestamp, serverID) {
		return nil
	}

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

	if p.tryProcessPlayerLogin(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessPlayerRegister(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessPlayerJoin(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessPlayerDisconnect(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessRoundStart(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessRoundEnd(ctx, line, timestamp, serverID) {
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

// tryProcessPlayerLogin parses Login request events (earliest connection event)
// This event happens first when a player connects to the server, before ServerRegisterClient
// Example: [2025.11.10-20.58.50:166][881]LogNet: Login request: ?InitialConnectTimeout=30?Name=ArmoredBear userId: SteamNWI:76561198995742987 platform: SteamNWI
func (p *LogParser) tryProcessPlayerLogin(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerLogin.FindStringSubmatch(line)
	if len(matches) < 5 {
		return false
	}

	// Group 1: timestamp
	// Group 2: player name
	// Group 3: Steam ID (or other platform ID)
	// Group 4: platform (SteamNWI, Epic, etc.)
	playerName := strings.TrimSpace(matches[2])
	playerID := strings.TrimSpace(matches[3])
	platform := strings.TrimSpace(matches[4])

	log.Printf("Player login request: %s (ID: %s, Platform: %s) on server %s", playerName, playerID, platform, serverID)

	// Create or update player record early in the connection process
	_, err := database.GetOrCreatePlayerBySteamID(ctx, p.pbApp, playerID, playerName)
	if err != nil {
		log.Printf("Error creating player from login request: %v", err)
		return false
	}

	return true
}

// tryProcessPlayerRegister parses ServerRegisterClient events (pre-match)
// This event happens before the player actually joins the match
// Example: [timestamp][ 23]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (76561198995742987) Result: (EOS_Success)
func (p *LogParser) tryProcessPlayerRegister(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerRegister.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	// Group 1: timestamp
	// Group 2: Steam ID
	steamID := strings.TrimSpace(matches[2])
	log.Printf("Player registered (pre-match): Steam ID %s on server %s", steamID, serverID)

	// We don't create the player here - wait for the LogNet "Join succeeded" event
	// which will have the player's name
	return true
}

// tryProcessPlayerJoin parses Join succeeded events (in-match)
// This event happens when the player actually joins the match
// Example: [timestamp][ 23]LogNet: Join succeeded: PlayerName
func (p *LogParser) tryProcessPlayerJoin(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerJoin.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	// Group 1: timestamp
	// Group 2: player name
	playerName := strings.TrimSpace(matches[2])

	// Check if player already exists by name
	player, err := database.GetPlayerByName(ctx, p.pbApp, playerName)
	if err != nil {
		// Player doesn't exist - create with name only (external_id will be empty)
		player, err = database.CreatePlayer(ctx, p.pbApp, "", playerName)
		if err != nil {
			log.Printf("Failed to create player on join: %v", err)
			return true
		}
	}

	// Send RCON join message with player stats (only for existing players with stats)
	if p.chatHandler != nil && p.chatHandler.rconSender != nil {
		go func() {
			// Get player stats and rank
			stats, rank, totalPlayers, err := database.GetPlayerStatsAndRank(ctx, p.pbApp, player.ID)
			if err != nil {
				log.Printf("Failed to get stats for player %s: %v", playerName, err)
				return
			}

			// Skip message for new players (no play time)
			if stats.TotalDurationSeconds == 0 {
				log.Printf("[JOIN] Skipping message for new player: %s", playerName)
				return
			}

			// Calculate score per minute
			scorePerMin := float64(stats.TotalScore) / (float64(stats.TotalDurationSeconds) / 60.0)

			// Format total play time
			hours := stats.TotalDurationSeconds / 3600
			minutes := (stats.TotalDurationSeconds % 3600) / 60

			var playTimeStr string
			if hours > 0 {
				playTimeStr = fmt.Sprintf("%dh %dm", hours, minutes)
			} else {
				playTimeStr = fmt.Sprintf("%dm", minutes)
			}

			// Build join message
			message := fmt.Sprintf("%s joined - #%d/%d with %.1f score/min over %s play time",
				playerName, rank, totalPlayers, scorePerMin, playTimeStr)

			// Send RCON message
			_, err = p.chatHandler.rconSender.SendRconCommand(serverID, fmt.Sprintf("say %s", message))
			if err != nil {
				log.Printf("Failed to send join message for %s: %v", playerName, err)
			} else {
				log.Printf("[JOIN] Sent RCON message: %s", message)
			}
		}()
	}

	return true
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

	// Check if this disconnect occurred shortly after a map travel
	// If so, it's just the server reconnecting players during map change, not a real disconnect
	if lastTravelTime, exists := p.lastMapTravelTimes[serverID]; exists {
		timeSinceTravel := timestamp.Sub(lastTravelTime)
		if timeSinceTravel >= 0 && timeSinceTravel < 30*time.Second {
			// This is a temporary disconnect during map travel, ignore it
			log.Printf("Ignoring disconnect for player %s during map travel (%.1fs after travel)", steamID, timeSinceTravel.Seconds())
			return true
		}
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

// tryProcessRoundStart parses and processes round start events
// Example: [2025.11.10-21.00.01:452][131]LogGameplayEvents: Display: Pre-round 2 started
// Resets the round_objective counter to 0 at the start of each round
func (p *LogParser) tryProcessRoundStart(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.RoundStart.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	roundNumStr := strings.TrimSpace(matches[2])
	roundNum, err := strconv.Atoi(roundNumStr)
	if err != nil {
		log.Printf("Failed to parse round number: %v", err)
		return true
	}

	log.Printf("Round %d started on server %s", roundNum, serverID)

	// Get active match to reset round_objective counter
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("No active match found for round start event on server %s", serverID)
		return true
	}

	// Reset the round_objective counter to 0 at the start of each round
	if err := database.ResetMatchRoundObjective(ctx, p.pbApp, activeMatch.ID); err != nil {
		log.Printf("Failed to reset round_objective for match %s: %v", activeMatch.ID, err)
	}

	return true
}

// tryProcessRoundEnd parses and processes round end events
// Example: [2025.11.10-20.59.41:417][930]LogGameplayEvents: Display: Round 1 Over: Team 1 won (win reason: Elimination)
func (p *LogParser) tryProcessRoundEnd(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.RoundEnd.FindStringSubmatch(line)
	if len(matches) < 5 {
		return false
	}

	// Round number might be empty in some log formats
	// roundNumStr := strings.TrimSpace(matches[2])
	winningTeamStr := strings.TrimSpace(matches[3])
	winReason := strings.TrimSpace(matches[4])

	winningTeam, err := strconv.Atoi(winningTeamStr)
	if err != nil {
		log.Printf("Failed to parse winning team: %v", err)
		return true
	}

	log.Printf("Round ended: Team %d won (reason: %s) on server %s", winningTeam, winReason, serverID)

	// Get active match to increment round counter
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("No active match found for round end event on server %s", serverID)
		return true
	}

	// Increment the round counter
	if err := database.IncrementMatchRound(ctx, p.pbApp, activeMatch.ID); err != nil {
		log.Printf("Failed to increment round for match %s: %v", activeMatch.ID, err)
	}

	// Determine if player team won based on match configuration
	// Team 0 = Security, Team 1 = Insurgents
	// If player_team is "Security" and team 0 won, or player_team is "Insurgents" and team 1 won,
	// then the player team won
	if activeMatch.PlayerTeam != nil {
		playerTeamWon := false
		if *activeMatch.PlayerTeam == "Security" && winningTeam == 0 {
			playerTeamWon = true
		} else if *activeMatch.PlayerTeam == "Insurgents" && winningTeam == 1 {
			playerTeamWon = true
		}

		if playerTeamWon {
			log.Printf("Player team (%s) won round %d in match %s", *activeMatch.PlayerTeam, activeMatch.Round+1, activeMatch.ID)
		} else {
			log.Printf("Player team (%s) lost round %d in match %s", *activeMatch.PlayerTeam, activeMatch.Round+1, activeMatch.ID)
		}
	}

	return true
}

// tryProcessMapLoad parses and processes map load events
// This occurs when the server first starts and loads the initial map
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

	// End active match (if any from previous server session) and create new one
	if err := p.endActiveMatchAndCreateNew(ctx, serverID, mapName, scenario, timestamp, playerTeam); err != nil {
		log.Printf("Failed to load new map: %v", err)
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

// findLastMapEvent searches backwards through a log file to find the most recent map event (MapTravel or MapLoad)
// Prioritizes MapTravel (runtime map change) over MapLoad (initial server start) since the server may not be on default map
// Returns map name, scenario, timestamp, and line number where the event was found, or error if not found
func (p *LogParser) findLastMapEvent(logFilePath string, beforeTime time.Time) (mapName, scenario string, timestamp time.Time, lineNumber int, err error) {
	file, err := os.Open(logFilePath)
	if err != nil {
		return "", "", time.Time{}, 0, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read file in reverse to find the last map event before the given time
	// For simplicity, we'll read all lines and process from end to start
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", "", time.Time{}, 0, fmt.Errorf("failed to read log file: %w", err)
	}

	// Search backwards through lines, checking MapTravel first, then MapLoad
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Try MapTravel pattern first (preferred as it means server is not on default map)
		matches := p.patterns.MapTravel.FindStringSubmatch(line)
		if len(matches) >= 4 {
			// Parse timestamp
			ts, err := parseTimestamp(matches[1])
			if err != nil {
				continue
			}

			// Only consider events before the target time
			if ts.After(beforeTime) {
				continue
			}

			// Found MapTravel event!
			mapName = strings.TrimSpace(matches[2])
			scenario = strings.TrimSpace(matches[3])
			timestamp = ts
			return mapName, scenario, timestamp, i, nil
		}

		// Try MapLoad pattern as fallback
		matches = p.patterns.MapLoad.FindStringSubmatch(line)
		if len(matches) >= 5 {
			// Parse timestamp
			ts, err := parseTimestamp(matches[1])
			if err != nil {
				continue
			}

			// Only consider events before the target time
			if ts.After(beforeTime) {
				continue
			}

			// Found MapLoad event!
			mapName = strings.TrimSpace(matches[2])
			scenario = strings.TrimSpace(matches[3])
			timestamp = ts
			return mapName, scenario, timestamp, i, nil
		}
	}

	return "", "", time.Time{}, 0, fmt.Errorf("no map event found before %v", beforeTime)
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

	// Increment the round_objective counter for the match
	if err := database.IncrementMatchRoundObjective(ctx, p.pbApp, activeMatch.ID); err != nil {
		log.Printf("Failed to increment round_objective for match %s: %v", activeMatch.ID, err)
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

	// Increment the round_objective counter for the match
	if err := database.IncrementMatchRoundObjective(ctx, p.pbApp, activeMatch.ID); err != nil {
		log.Printf("Failed to increment round_objective for match %s: %v", activeMatch.ID, err)
	}

	return true
}

// getOrCreateMatchForEvent retrieves the active match for a server, or creates one if none exists
// by searching backwards in the log file for the last map event (MapTravel or MapLoad).
// NOTE: Catch-up processing is NOT done here to avoid duplicate event processing.
// The watcher will naturally process all events as it reads through the file.
func (p *LogParser) getOrCreateMatchForEvent(ctx context.Context, serverID, logFilePath string, timestamp time.Time, eventName string) (*database.Match, error) {
	// Try to get active match first
	activeMatch, err := database.GetActiveMatch(ctx, p.pbApp, serverID)
	if err == nil && activeMatch != nil {
		return activeMatch, nil
	}

	// No active match found - search backwards in log for last map event
	log.Printf("No active match found for %s event on server %s, searching for last map event...", eventName, serverID)

	mapName, scenario, mapLoadTime, startLineNum, err := p.findLastMapEvent(logFilePath, timestamp)
	if err != nil {
		log.Printf("Failed to find last map event: %v, creating match with unknown map", err)
		// Create match with unknown map info
		mode := "Unknown"
		activeMatch, err = database.CreateMatch(ctx, p.pbApp, serverID, nil, &mode, &timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to create match for %s event: %w", eventName, err)
		}
		log.Printf("Created new match %s for server %s (unknown map)", activeMatch.ID, serverID)
		return activeMatch, nil
	}

	log.Printf("Found last map event: %s (%s) at %v (line %d)", mapName, scenario, mapLoadTime, startLineNum)

	// Extract player team from scenario
	var playerTeam *string
	if strings.Contains(scenario, "_Security") {
		team := "Security"
		playerTeam = &team
	} else if strings.Contains(scenario, "_Insurgents") {
		team := "Insurgents"
		playerTeam = &team
	}

	// Create match with proper map info
	activeMatch, err = database.CreateMatch(ctx, p.pbApp, serverID, &mapName, &scenario, &mapLoadTime, playerTeam)
	if err != nil {
		return nil, fmt.Errorf("failed to create match for %s event: %w", eventName, err)
	}

	log.Printf("Created new match %s for server %s (watcher will process events naturally)", activeMatch.ID, serverID)

	// NOTE: We do NOT call catchUpOnMissedEvents here because the watcher is already
	// processing the file sequentially. Doing catch-up would cause duplicate event processing:
	// 1. Event at line 500 triggers this function
	// 2. Match created based on map event at line 100
	// 3. If we catch up from line 100-500, the event at line 500 gets processed twice
	//
	// Instead, the watcher continues processing from where it left off, and subsequent events
	// will now find this active match.

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
