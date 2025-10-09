package events

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EventType represents different types of game events
type EventType int

const (
	EventPlayerKill EventType = iota
	EventPlayerJoin
	EventPlayerLeave
	EventRoundStart
	EventRoundEnd
	EventGameOver
	EventMapLoad
	EventDifficultyChange
	EventMapVote
	EventChatCommand
	EventRconCommand
	EventFriendlyFire
	EventSuicide
	EventFallDamage
)

// ParseEventType converts a string to EventType
func ParseEventType(s string) EventType {
	switch s {
	case "player_kill":
		return EventPlayerKill
	case "player_join":
		return EventPlayerJoin
	case "player_leave":
		return EventPlayerLeave
	case "round_start":
		return EventRoundStart
	case "round_end":
		return EventRoundEnd
	case "game_over":
		return EventGameOver
	case "map_load":
		return EventMapLoad
	case "difficulty_change":
		return EventDifficultyChange
	case "map_vote":
		return EventMapVote
	case "chat_command":
		return EventChatCommand
	case "rcon_command":
		return EventRconCommand
	case "friendly_fire":
		return EventFriendlyFire
	case "suicide":
		return EventSuicide
	case "fall_damage":
		return EventFallDamage
	default:
		return EventPlayerKill // Default fallback
	}
}

// String returns the string representation of the event type
func (e EventType) String() string {
	switch e {
	case EventPlayerKill:
		return "player_kill"
	case EventPlayerJoin:
		return "player_join"
	case EventPlayerLeave:
		return "player_leave"
	case EventRoundStart:
		return "round_start"
	case EventRoundEnd:
		return "round_end"
	case EventGameOver:
		return "game_over"
	case EventMapLoad:
		return "map_load"
	case EventDifficultyChange:
		return "difficulty_change"
	case EventMapVote:
		return "map_vote"
	case EventChatCommand:
		return "chat_command"
	case EventRconCommand:
		return "rcon_command"
	case EventFriendlyFire:
		return "friendly_fire"
	case EventSuicide:
		return "suicide"
	case EventFallDamage:
		return "fall_damage"
	default:
		return "unknown"
	}
}

// LogPatterns contains all regex patterns for parsing log entries
type LogPatterns struct {
	PlayerKill       *regexp.Regexp
	PlayerJoin       *regexp.Regexp
	PlayerDisconnect *regexp.Regexp
	PlayerRconLeave  *regexp.Regexp
	RoundStart       *regexp.Regexp
	RoundEnd         *regexp.Regexp
	GameOver         *regexp.Regexp
	MapLoad          *regexp.Regexp
	DifficultyChange *regexp.Regexp
	MapVote          *regexp.Regexp
	ChatCommand      *regexp.Regexp
	RconCommand      *regexp.Regexp
	FallDamage       *regexp.Regexp
	Timestamp        *regexp.Regexp
}

// NewLogPatterns creates and compiles all regex patterns
func NewLogPatterns() *LogPatterns {
	return &LogPatterns{
		// Kill events - always provide consistent capture groups for killer/victim/weapon fields
		// PlayerKill: timestamp, killerSection, victimName, victimSteam, victimTeam, weapon
		PlayerKill: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogGameplayEvents: Display: (.+?) killed ([^\[]+)\[([^,\]]*), team (\d+)\] with (.+)$`),

		// Player connection events - handles both LogNet and LogGameMode formats
		PlayerJoin: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]Log(?:Net: Join succeeded: (.+)|GameMode: Display: Player \d+ '([^']+)' joined team (\d+))`),

		PlayerDisconnect: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId \((\d+)\), Result: \(EOS_Success\)`),

		PlayerRconLeave: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogRcon: .* << say See you later, (.+)!`),

		// Game state events
		RoundStart: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogGameplayEvents: Display: (?:Pre-)?round (\d+) started`),

		RoundEnd: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]Log(?:GameMode|GameplayEvents): Display: Round (?:(\d+) )?O\s*ver: Team (\d+) won \(win reason: (.+)\)`),

		GameOver: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogGameplayEvents: Display: Game over`),

		// Map and server events
		MapLoad: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogLoad: LoadMap: /Game/Maps/([^/]+)/[^?]+\?.*Scenario=([^?&]+).*MaxPlayers=(\d+).*Lighting=([^?&]+)`),

		DifficultyChange: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogAI: Warning: AI difficulty set to ([0-9.]+)`),

		MapVote: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogMapVoteManager: Display: .*Vote Options:`),

		// Chat and RCON events
		ChatCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\d+\]LogChat: Display: ([^(]+)\((\d+)\) Global Chat: (!.+)`),

		RconCommand: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogRcon: ([^<]+)<< (.+)`),

		// Other events
		FallDamage: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]\[\s*\d+\]LogSoldier: Applying ([0-9.]+) fall damage`),

		// Utility pattern for timestamp extraction
		Timestamp: regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]`),
	}
}

// EventParser handles parsing of log lines into game events
type EventParser struct {
	patterns *LogPatterns
}

// NewEventParser creates a new event parser
func NewEventParser() *EventParser {
	return &EventParser{
		patterns: NewLogPatterns(),
	}
}

// ParseLine parses a single log line and returns a GameEvent if recognized
func (p *EventParser) ParseLine(line, serverID string) (*GameEvent, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	// Extract timestamp first
	timestampMatches := p.patterns.Timestamp.FindStringSubmatch(line)
	if len(timestampMatches) < 2 {
		return nil, nil // Skip lines without proper timestamp
	}

	timestamp, err := parseTimestamp(timestampMatches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Try to match against each pattern
	if event := p.parseKillEvent(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parsePlayerJoin(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parsePlayerDisconnect(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parsePlayerRconLeave(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseRoundStart(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseRoundEnd(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseGameOver(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseMapLoad(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseDifficultyChange(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseMapVote(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseChatCommand(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseRconCommand(line, timestamp, serverID); event != nil {
		return event, nil
	}

	if event := p.parseFallDamage(line, timestamp, serverID); event != nil {
		return event, nil
	}

	return nil, nil // No pattern matched
}

// parseKillEvent parses kill events and determines kill type (handles single and multi-player kills)
func (p *EventParser) parseKillEvent(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.PlayerKill.FindStringSubmatch(line)
	if len(matches) < 7 {
		return nil
	}

	killerSection := strings.TrimSpace(matches[2])
	victimName := strings.TrimSpace(matches[3])
	victimSteamID := strings.TrimSpace(matches[4])
	victimTeam, _ := strconv.Atoi(matches[5])
	weapon := cleanWeaponName(matches[6])

	// Parse killer section - could be single player, multiple players, or "?" for AI suicide
	killers := parseKillerSection(killerSection)

	// If no valid killers found (like AI suicide with "?"), skip this event
	if len(killers) == 0 {
		return nil
	}

	// Determine event type and kill type
	var eventType EventType
	var killType KillType

	// Check if it's a suicide (killer is victim)
	if len(killers) == 1 && killers[0].Name == victimName && killers[0].SteamID == victimSteamID {
		eventType = EventSuicide
		killType = KillTypeSuicide
	} else if len(killers) == 1 && killers[0].Team == victimTeam {
		// Friendly fire (same team)
		eventType = EventFriendlyFire
		killType = KillTypeFriendlyFire
	} else if victimSteamID == "INVALID" {
		// AI victim - this is what we want to track
		eventType = EventPlayerKill
		killType = KillTypeRegular
	} else {
		// Player vs player
		eventType = EventPlayerKill
		killType = KillTypeRegular
	}

	// For multi-player kills, we'll create separate events for each killer
	// but return the primary event with all killer data
	data := map[string]interface{}{
		"killers":         killers,
		"victim_name":     victimName,
		"victim_steam_id": victimSteamID,
		"victim_team":     victimTeam,
		"weapon":          weapon,
		"kill_type":       string(killType),
		"multi_kill":      len(killers) > 1,
	}

	return &GameEvent{
		Type:       eventType,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parsePlayerJoin parses player join events
func (p *EventParser) parsePlayerJoin(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.PlayerJoin.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	playerName := strings.TrimSpace(matches[2])

	data := map[string]interface{}{
		"player_name": playerName,
	}

	return &GameEvent{
		Type:       EventPlayerJoin,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parsePlayerDisconnect parses player disconnect events
func (p *EventParser) parsePlayerDisconnect(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.PlayerDisconnect.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	steamID := strings.TrimSpace(matches[2])

	data := map[string]interface{}{
		"steam_id": steamID,
	}

	return &GameEvent{
		Type:       EventPlayerLeave,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parsePlayerRconLeave parses RCON leave messages
func (p *EventParser) parsePlayerRconLeave(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.PlayerRconLeave.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	playerName := strings.TrimSpace(matches[2])

	data := map[string]interface{}{
		"player_name": playerName,
	}

	return &GameEvent{
		Type:       EventPlayerLeave,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseRoundStart parses round start events
func (p *EventParser) parseRoundStart(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.RoundStart.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	roundNumber, _ := strconv.Atoi(matches[2])

	data := map[string]interface{}{
		"round_number": roundNumber,
	}

	return &GameEvent{
		Type:       EventRoundStart,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseRoundEnd parses round end events
func (p *EventParser) parseRoundEnd(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.RoundEnd.FindStringSubmatch(line)
	if len(matches) < 5 {
		return nil
	}

	var roundNumber int
	if matches[2] != "" {
		roundNumber, _ = strconv.Atoi(matches[2])
	}
	winningTeam, _ := strconv.Atoi(matches[3])
	winReason := strings.TrimSpace(matches[4])

	data := map[string]interface{}{
		"round_number": roundNumber,
		"winning_team": winningTeam,
		"win_reason":   winReason,
	}

	return &GameEvent{
		Type:       EventRoundEnd,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseGameOver parses game over events
func (p *EventParser) parseGameOver(line string, timestamp time.Time, serverID string) *GameEvent {
	if !p.patterns.GameOver.MatchString(line) {
		return nil
	}

	return &GameEvent{
		Type:       EventGameOver,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       map[string]interface{}{},
		RawLogLine: line,
	}
}

// parseMapLoad parses map loading events
func (p *EventParser) parseMapLoad(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.MapLoad.FindStringSubmatch(line)
	if len(matches) < 5 {
		return nil
	}

	mapName := strings.TrimSpace(matches[2])
	scenario := strings.TrimSpace(matches[3])
	maxPlayers, _ := strconv.Atoi(matches[4])
	lighting := strings.TrimSpace(matches[5])

	// Clean up scenario name
	scenario = strings.ReplaceAll(scenario, "Scenario_", "")
	scenario = strings.ReplaceAll(scenario, "_", " ")

	data := map[string]interface{}{
		"map_name":    mapName,
		"scenario":    scenario,
		"max_players": maxPlayers,
		"lighting":    lighting,
	}

	return &GameEvent{
		Type:       EventMapLoad,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseDifficultyChange parses AI difficulty change events
func (p *EventParser) parseDifficultyChange(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.DifficultyChange.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	difficulty, _ := strconv.ParseFloat(matches[2], 64)

	data := map[string]interface{}{
		"difficulty": difficulty,
	}

	return &GameEvent{
		Type:       EventDifficultyChange,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseMapVote parses map voting events
func (p *EventParser) parseMapVote(line string, timestamp time.Time, serverID string) *GameEvent {
	if !p.patterns.MapVote.MatchString(line) {
		return nil
	}

	return &GameEvent{
		Type:       EventMapVote,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       map[string]interface{}{},
		RawLogLine: line,
	}
}

// parseChatCommand parses chat command events
func (p *EventParser) parseChatCommand(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.ChatCommand.FindStringSubmatch(line)
	if len(matches) < 5 {
		return nil
	}

	playerName := strings.TrimSpace(matches[2])
	steamID := strings.TrimSpace(matches[3])
	message := strings.TrimSpace(matches[4])

	// Parse command and arguments
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	var arguments []string
	if len(parts) > 1 {
		arguments = parts[1:]
	}

	data := map[string]interface{}{
		"player_name": playerName,
		"steam_id":    steamID,
		"command":     command,
		"arguments":   arguments,
		"raw_message": message,
	}

	return &GameEvent{
		Type:       EventChatCommand,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseRconCommand parses RCON command events
func (p *EventParser) parseRconCommand(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.RconCommand.FindStringSubmatch(line)
	if len(matches) < 4 {
		return nil
	}

	client := strings.TrimSpace(matches[2])
	command := strings.TrimSpace(matches[3])

	data := map[string]interface{}{
		"client":  client,
		"command": command,
	}

	return &GameEvent{
		Type:       EventRconCommand,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseFallDamage parses fall damage events
func (p *EventParser) parseFallDamage(line string, timestamp time.Time, serverID string) *GameEvent {
	matches := p.patterns.FallDamage.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	damage, _ := strconv.ParseFloat(matches[2], 64)

	data := map[string]interface{}{
		"damage": damage,
	}

	return &GameEvent{
		Type:       EventFallDamage,
		Timestamp:  timestamp,
		ServerID:   serverID,
		Data:       data,
		RawLogLine: line,
	}
}

// parseTimestamp parses UE4 timestamp format
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

// cleanWeaponName cleans up weapon names from BP format
func cleanWeaponName(weapon string) string {
	// Remove BP_ prefix and _C_ suffix with ID
	weapon = strings.TrimSpace(weapon)

	// Handle BP_ prefix
	weapon = strings.TrimPrefix(weapon, "BP_")

	// Remove _C_ and everything after it
	if idx := strings.Index(weapon, "_C_"); idx != -1 {
		weapon = weapon[:idx]
	}

	// Replace underscores with spaces for readability
	weapon = strings.ReplaceAll(weapon, "_", " ")

	// Handle specific weapon types
	if strings.HasPrefix(weapon, "Firearm ") {
		weapon = strings.TrimPrefix(weapon, "Firearm ")
	} else if strings.HasPrefix(weapon, "Projectile ") {
		weapon = strings.TrimPrefix(weapon, "Projectile ")
	} else if strings.HasPrefix(weapon, "Character ") {
		weapon = "Fall Damage"
	}

	return weapon
}

// parseKillerSection parses the killer section which can contain single or multiple players
// Examples:
// "ArmoredBear[76561198995742987, team 0]" - single player
// "-=312th=- Rabbit[76561198262186571, team 0] + ArmoredBear[76561198995742987, team 0]" - multiple players
// "?" - AI suicide (should return empty slice)
func parseKillerSection(killerSection string) []Killer {
	var killers []Killer

	// Handle AI suicide case
	if strings.TrimSpace(killerSection) == "?" {
		return killers
	}

	// Split by " + " for multi-player kills
	playerParts := strings.Split(killerSection, " + ")

	// Regex to parse individual player: "PlayerName[SteamID, team X]"
	playerRegex := regexp.MustCompile(`^(.+?)\[([^,\]]*), team (\d+)\]$`)

	for _, playerPart := range playerParts {
		playerPart = strings.TrimSpace(playerPart)
		matches := playerRegex.FindStringSubmatch(playerPart)

		if len(matches) == 4 {
			team, _ := strconv.Atoi(matches[3])
			killer := Killer{
				Name:    strings.TrimSpace(matches[1]),
				SteamID: strings.TrimSpace(matches[2]),
				Team:    team,
			}
			killers = append(killers, killer)
		}
	}

	return killers
}
