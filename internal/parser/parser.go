package parser

import (
	"bufio"
	"context"
	"fmt"

	// "log"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sandstorm-tracker/internal/events"

	"github.com/pocketbase/pocketbase/core"
)

// Context key for marking catchup mode
type contextKey string

const isCatchupModeKey contextKey = "isCatchupMode"

// isCatchupMode checks if the context indicates catchup mode
func isCatchupMode(ctx context.Context) bool {
	isCatchup, _ := ctx.Value(isCatchupModeKey).(bool)
	return isCatchup
}

// LogParser handles parsing log lines and writing directly to database
type LogParser struct {
	pbApp              core.App
	logger             *slog.Logger
	patterns           *logPatterns
	lastMapTravelTimes map[string]time.Time // Track last map travel time per server to ignore reconnects
	eventCreator       *events.Creator      // Creates event records for hook-based processing
}

// logPatterns contains compiled regex patterns for log parsing
type logPatterns struct {
	LogFileOpen      *regexp.Regexp // Log file open timestamp (first line of log)
	CommandLine      *regexp.Regexp
	PlayerKill       *regexp.Regexp
	PlayerLogin      *regexp.Regexp // Login request (earliest connection event)
	PlayerRegister   *regexp.Regexp // ServerRegisterClient (pre-match)
	PlayerJoin       *regexp.Regexp // Join succeeded (in-match)
	PlayerConnection *regexp.Regexp // Connection event (accepts post-challenge connection)
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

func NewLogPatterns() *logPatterns {
	return &logPatterns{
		// Log file open timestamp (first line of every log file)
		// Format: "Log file open, 11/10/25 20:58:31"
		// Note: May have UTF-8 BOM (U+FEFF) at the start
		LogFileOpen: regexp.MustCompile(`^(?:` + "\uFEFF" + `)?Log file open, (\d{2}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s*$`),

		CommandLine: regexp.MustCompile(`LogInit: Command Line:\s+(\w+)\?Scenario=([^?]+)\?MaxPlayers=(\d+)\?Game=([^?]+)\?Lighting=(\w+).*?-Hostname="([^"]+)"`),
		// Kill events - always provide consistent capture groups for killer/victim/weapon fields
		// PlayerKill: timestamp, killerSection, victimName, victimSteam, victimTeam, weapon
		PlayerKill: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: (.+?) killed (.+?) with (.+)$`),

		// Player connection events - three stages:
		// 1. PlayerLogin: [timestamp][id]LogNet: Login request (earliest connection event with name & Steam ID)
		// 2. PlayerRegister: [timestamp][id]LogEOSAntiCheat: ServerRegisterClient (happens after login)
		// 3. PlayerJoin: [timestamp][id]LogNet: Join succeeded (happens when player actually in match)
		PlayerLogin:    regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogNet: Login request:.*\?Name=(.+?) userId: \w+:(\d+) platform: (\w+)`),
		PlayerRegister: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogEOSAntiCheat: Display: ServerRegisterClient: Client: \((\d+)\) Result: \(EOS_Success\)`),
		PlayerJoin:     regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogNet: Join succeeded: ([^\r\n]+)`),
		// Player connection event - accepts post-challenge connection (IP capture)
		// Format: [timestamp][id]LogNet: Server accepting post-challenge connection from: IP:PORT
		PlayerConnection: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogNet: Server accepting post-challenge connection from: ([0-9.]+):`),

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

// extractPlayerTeam extracts the player team from a scenario string
// Returns "Security" or "Insurgents" based on scenario suffix, or empty string if undetermined
func extractPlayerTeam(scenario string) string {
	if strings.HasSuffix(scenario, "_Security") {
		return "Security"
	}
	if strings.HasSuffix(scenario, "_Insurgents") {
		return "Insurgents"
	}
	return ""
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

	p.logger.Debug("Map travel detected", "map", mapName, "scenario", scenario, "serverID", serverID)

	// Track this map travel time so we can ignore immediate disconnects/reconnects
	p.lastMapTravelTimes[serverID] = timestamp

	// Emit map travel event for handler to process match transition
	if p.eventCreator != nil {
		err := p.eventCreator.CreateEvent(events.TypeMapTravel, serverID, map[string]interface{}{
			"map":        mapName,
			"scenario":   scenario,
			"timestamp":  timestamp,
			"is_catchup": isCatchupMode(ctx),
		})
		if err != nil {
			p.logger.Error("Failed to create map travel event",
				"map", mapName, "scenario", scenario, "error", err.Error())
		}
	}

	return true
}

// NewLogParser creates a new log parser with PocketBase app
func NewLogParser(pbApp core.App, logger *slog.Logger) *LogParser {
	return &LogParser{
		patterns:           NewLogPatterns(),
		pbApp:              pbApp,
		logger:             logger,
		lastMapTravelTimes: make(map[string]time.Time),
		eventCreator:       events.NewCreator(pbApp), // Initialize event creator for dual-write phase
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

	if p.tryProcessPlayerConnection(ctx, line, timestamp, serverID) {
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

	if p.tryProcessGameOver(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessLogFileOpen(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessChatCommand(ctx, line, timestamp, serverID) {
		return nil
	}

	// Add other event types as needed (round start/end, etc.)
	// For now, we're focusing on the core stat-tracking events

	return nil
}

// tryProcessLogFileOpen handles the log file creation event (first line of log)
// This occurs when a new log file is created (server restart or log rotation)
func (p *LogParser) tryProcessLogFileOpen(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.LogFileOpen.FindStringSubmatch(line)
	if len(matches) < 2 {
		return false
	}

	p.logger.Debug("Log file created", "serverID", serverID, "timestamp", timestamp)

	// Emit log file created event for handler to process
	if p.eventCreator != nil {
		err := p.eventCreator.CreateEvent(events.TypeLogFileCreated, serverID, map[string]interface{}{
			"timestamp": timestamp,
		})
		if err != nil {
			p.logger.Error("Failed to create log file created event",
				"server", serverID,
				"timestamp", timestamp,
				"error", err,
			)
			return true
		}
	}

	return true
}

// tryProcessPlayerConnection handles player connection events (accepts post-challenge connection)
// This captures the IP address of connecting players before they fully login
func (p *LogParser) tryProcessPlayerConnection(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.PlayerConnection.FindStringSubmatch(line)
	if len(matches) < 3 {
		return false
	}

	ip := matches[2]

	p.logger.Debug("Player connection from IP", "ip", ip, "serverID", serverID)

	// Store IP in app store with key format: "serverID:lastIP"
	// This will be used when the next player_login event occurs
	storeKey := fmt.Sprintf("%s:lastIP", serverID)

	// Check if there's already a pending IP (multiple connections before login)
	existingIP := p.pbApp.Store().Get(storeKey)
	if existingIP != nil {
		// Another connection came in before the previous one logged in
		// Discard both to avoid confusion about which IP connects to which player
		p.logger.Debug("Discarding duplicate connection IP - unable to match with player", "serverID", serverID)
		p.pbApp.Store().Set(storeKey, nil)
		return true
	}

	// Store the IP for use on the next player_login event
	p.pbApp.Store().Set(storeKey, ip)

	return true
}

// tryProcessKillEvent parses and processes kill events
func (p *LogParser) tryProcessKillEvent(ctx context.Context, line string, timestamp time.Time, serverID string, logFilePath string) bool {
	matches := p.patterns.PlayerKill.FindStringSubmatch(line)
	if len(matches) < 5 {
		return false
	}

	killerSection := strings.TrimSpace(matches[2])
	victimSection := strings.TrimSpace(matches[3])
	weapon := matches[4]

	// Parse killer section
	killers := ParseKillerSection(killerSection)
	if len(killers) == 0 {
		return true // Parsed but no valid killers (AI suicide)
	}

	// Parse victim section as a single Killer struct
	victimParsed := ParseKillerSection(victimSection)
	var victim killer
	if len(victimParsed) > 0 {
		victim = victimParsed[0]
	} else {
		victim = killer{Name: victimSection, SteamID: "", Team: -1}
	}
	// Determine if this is PvP (player victim) or PvE (bot victim)
	isPvP := victim.SteamID != "INVALID"

	// EVENT-DRIVEN ARCHITECTURE: Create event records for hook-based processing
	// Hooks handle all database updates, player creation, scoring, and ML classification
	if p.eventCreator != nil && len(killers) > 0 {
		// Build killers array for event data as array of killer structs
		killersArr := make([]killer, 0, len(killers))
		for _, k := range killers {
			if k.SteamID == "INVALID" {
				continue
			}
			killersArr = append(killersArr, k)
		}

		// Emit a single event with all killers in the array (as []killer)
		// Note: weapon is passed raw (unsanitized) - UpsertMatchWeaponStats will handle
		// cleaning the name and extracting the weapon type internally
		err := p.eventCreator.CreateEvent(events.TypePlayerKill, serverID, map[string]interface{}{
			"killers":    killersArr,
			"victim":     victim,
			"weapon":     weapon,
			"is_catchup": isCatchupMode(ctx),
		})
		if err != nil {
			p.logger.Error("Failed to create kill event",
				"killers", killersArr,
				"victim", victim,
				"weapon", weapon,
				"error", err.Error())
		} else {
			p.logger.Debug("Created kill event",
				"killers", killersArr,
				"victim", victim,
				"weapon", weapon,
				"isPvP", isPvP)
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
	steamID := strings.TrimSpace(matches[3])
	platform := strings.TrimSpace(matches[4])

	p.logger.Debug("Player login request", "playerName", playerName, "steamID", steamID, "platform", platform, "serverID", serverID)

	// Create player_login event (handler will create/update player record)
	if p.eventCreator != nil {
		err := p.eventCreator.CreatePlayerLoginEvent(serverID, playerName, steamID, platform, isCatchupMode(ctx))
		if err != nil {
			p.logger.Debug("Failed to create player_login event", "error", err)
		}
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
	p.logger.Debug("Player registered (pre-match)", "steamID", steamID, "serverID", serverID)

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

	// Create player join event (handler will ensure player exists, add to match, and send RCON message)
	if p.eventCreator != nil {
		err := p.eventCreator.CreatePlayerJoinEvent(serverID, playerName, isCatchupMode(ctx))
		if err != nil {
			p.logger.Debug("Failed to create player_join event", "error", err)
		}
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
	isMapTravelDisconnect := false
	if lastTravelTime, exists := p.lastMapTravelTimes[serverID]; exists {
		timeSinceTravel := timestamp.Sub(lastTravelTime)
		if timeSinceTravel >= 0 && timeSinceTravel < 30*time.Second {
			// This is a temporary disconnect during map travel, ignore it for database updates
			// but don't return early - we still need to handle it below
			p.logger.Debug("Ignoring disconnect for player during map travel", "steamID", steamID, "secondsAfterTravel", timeSinceTravel.Seconds())
			isMapTravelDisconnect = true
		}
	}

	// Only create leave event if this is a real disconnect (not map travel)
	if !isMapTravelDisconnect && p.eventCreator != nil {
		// Create player_leave event with raw Steam ID (handler will do player lookup)
		err := p.eventCreator.CreatePlayerLeaveEvent(serverID, steamID, "")
		if err != nil {
			p.logger.Debug("Failed to create player_leave event", "error", err)
		}
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
		p.logger.Debug("Failed to parse round number", "error", err)
		return true
	}

	p.logger.Debug("Round started on server", "roundNum", roundNum, "serverID", serverID)

	// Emit round start event - handler will reset round objectives
	if p.eventCreator != nil {
		err := p.eventCreator.CreateRoundStartEvent(serverID, "", roundNum)
		if err != nil {
			p.logger.Error("Failed to create round start event",
				"round", roundNum, "error", err.Error())
		}
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

	winningTeamStr := strings.TrimSpace(matches[3])

	winningTeam, err := strconv.Atoi(winningTeamStr)
	if err != nil {
		p.logger.Debug("Failed to parse winning team", "error", err)
		return true
	}

	if p.eventCreator != nil {
		err := p.eventCreator.CreateRoundEndEvent(serverID, "", 0, winningTeam, isCatchupMode(ctx))
		if err != nil {
			p.logger.Error("Failed to create round end event",
				"winningTeam", winningTeam, "serverID", serverID, "error", err.Error())
		}
	}

	return true
}

// tryProcessGameOver parses and processes game over events
// This occurs when a match ends, and we want to immediately capture final scores
// Example: [2025.11.10-21.12.25:385][831]LogSession: Display: AINSGameSession::HandleMatchHasEnded
func (p *LogParser) tryProcessGameOver(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	if !p.patterns.GameOver.MatchString(line) {
		return false
	}

	p.logger.Debug("Game over detected", "serverID", serverID)

	// Emit game over event - handler will finalize match
	if p.eventCreator != nil {
		err := p.eventCreator.CreateEvent(events.TypeGameOver, serverID, map[string]interface{}{
			"is_catchup": isCatchupMode(ctx),
		})
		if err != nil {
			p.logger.Error("Failed to create game over event",
				"serverID", serverID, "error", err.Error())
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

	// Extract player team from scenario
	playerTeam := extractPlayerTeam(scenario)
	var playerTeamPtr *string
	if playerTeam != "" {
		playerTeamPtr = &playerTeam
	}

	if p.eventCreator != nil {
		// Emit map load event - handler will create new match and start event
		err := p.eventCreator.CreateEvent(events.TypeMapLoad, serverID, map[string]interface{}{
			"map":         mapName,
			"scenario":    scenario,
			"timestamp":   timestamp,
			"player_team": playerTeamPtr,
			"is_catchup":  isCatchupMode(ctx),
		})
		if err != nil {
			p.logger.Error("Failed to create map load event",
				"map", mapName, "scenario", scenario, "error", err.Error())
		}
	}

	return true
}

// Helper types and functions

type killer struct {
	Name    string
	SteamID string
	Team    int
}

func ParseKillerSection(killerSection string) []killer {
	var killers []killer

	if strings.TrimSpace(killerSection) == "?" {
		return killers
	}

	// Determine separator based on content:
	// Kills use " + " as separator: Name[SteamID, team X] + Name[SteamID, team X]
	// Objectives use ", " as separator: Name[SteamID], Name[SteamID]
	var playerParts []string
	if strings.Contains(killerSection, " + ") {
		// Kill events with " + " separator
		playerParts = strings.Split(killerSection, " + ")
	} else if strings.Contains(killerSection, ", ") && !strings.Contains(killerSection, "team") {
		// Objective events with ", " separator (no "team" keyword means it's objective format)
		playerParts = strings.Split(killerSection, ", ")
	} else {
		playerParts = []string{killerSection}
	}

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

func CleanWeaponName(weapon string) string {
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

	// Parse the date/time part in local timezone (log timestamps are in server's local time)
	dt, err := time.ParseInLocation("2006.01.02-15.04.05", dateTimePart, time.Local)
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
// FindLastMapEvent finds the most recent map event in a log file before the given time.
// Prioritizes MapTravel (runtime map change) over MapLoad (initial server start) since the server may not be on default map
// Returns map name, scenario, timestamp, and line number where the event was found, or error if not found
func (p *LogParser) FindLastMapEvent(logFilePath string, beforeTime time.Time) (mapName, scenario string, timestamp time.Time, lineNumber int, err error) {
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
	destroyingTeam := strings.TrimSpace(matches[4])
	playerSection := strings.TrimSpace(matches[5])

	// Parse player section (can have multiple players)
	killers := ParseKillerSection(playerSection)
	if len(killers) == 0 {
		return true // Parsed but no valid players
	}

	if p.eventCreator != nil {
		team, _ := strconv.Atoi(destroyingTeam)

		// Convert killers to ObjectivePlayer array
		objectivePlayers := make([]events.ObjectivePlayer, 0)
		for _, killer := range killers {
			if killer.SteamID == "INVALID" || killer.SteamID == "" {
				continue
			}
			objectivePlayers = append(objectivePlayers, events.ObjectivePlayer{
				SteamID:    killer.SteamID,
				PlayerName: killer.Name,
			})
		}

		if len(objectivePlayers) > 0 {
			// Create single event with all players
			err := p.eventCreator.CreateObjectiveDestroyedEvent(
				serverID,
				"",
				objectiveNum,
				objectivePlayers,
				team,
				isCatchupMode(ctx),
			)
			if err != nil {
				p.logger.Error("Failed to create objective destroyed event",
					"objective", objectiveNum, "players", len(objectivePlayers), "error", err.Error())
			}
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
	playerSection := strings.TrimSpace(matches[5])

	// Parse player section (can have multiple players)
	killers := ParseKillerSection(playerSection)
	if len(killers) == 0 {
		return true // Parsed but no valid players
	}

	if p.eventCreator != nil {
		team, _ := strconv.Atoi(capturingTeam)

		// Convert killers to ObjectivePlayer array
		objectivePlayers := make([]events.ObjectivePlayer, 0)
		for _, killer := range killers {
			if killer.SteamID == "INVALID" || killer.SteamID == "" {
				continue
			}
			objectivePlayers = append(objectivePlayers, events.ObjectivePlayer{
				SteamID:    killer.SteamID,
				PlayerName: killer.Name,
			})
		}

		if len(objectivePlayers) > 0 {
			// Create single event with all players
			err := p.eventCreator.CreateObjectiveCapturedEvent(
				serverID,
				"",
				objectiveNum,
				objectivePlayers,
				team,
				isCatchupMode(ctx),
			)
			if err != nil {
				p.logger.Error("Failed to create objective captured event",
					"objective", objectiveNum, "players", len(objectivePlayers), "error", err.Error())
			}
		}
	}

	return true
}

// tryProcessChatCommand parses chat commands and emits events for handling
func (p *LogParser) tryProcessChatCommand(ctx context.Context, line string, timestamp time.Time, serverID string) bool {
	matches := p.patterns.ChatCommand.FindStringSubmatch(line)
	if len(matches) < 5 {
		return false
	}

	playerName := strings.TrimSpace(matches[2])
	steamID := strings.TrimSpace(matches[3])
	command := strings.TrimSpace(matches[4])

	p.logger.Debug("Chat command", "playerName", playerName, "steamID", steamID, "command", command)

	// Emit chat command event for handler to process
	if p.eventCreator != nil {
		err := p.eventCreator.CreateChatCommandEvent(
			serverID,
			steamID,
			playerName,
			command,
			[]string{}, // Args can be parsed from command if needed in future
			isCatchupMode(ctx),
		)
		if err != nil {
			p.logger.Error("Failed to create chat command event",
				"player", playerName,
				"steam_id", steamID,
				"command", command,
				"error", err.Error(),
			)
		}
	}

	return true
}
