package app

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	db "sandstorm-tracker/internal/db/generated"
)

// LogParser handles parsing log lines and writing directly to database
type LogParser struct {
	patterns *logPatterns
	queries  *db.Queries
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

// NewLogParser creates a new log parser with database queries
func NewLogParser(queries *db.Queries) *LogParser {
	return &LogParser{
		patterns: newLogPatterns(),
		queries:  queries,
	}
}

// ParseAndProcess parses a log line and writes to database if it's a recognized event
func (p *LogParser) ParseAndProcess(ctx context.Context, line string, serverID int64) error {
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
	if p.tryProcessKillEvent(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessPlayerJoin(ctx, line, timestamp, serverID) {
		return nil
	}

	if p.tryProcessPlayerDisconnect(ctx, line, timestamp, serverID) {
		return nil
	}

	// Add other event types as needed (round start/end, map load, etc.)
	// For now, we're focusing on the core stat-tracking events

	return nil
}

// tryProcessKillEvent parses and processes kill events
func (p *LogParser) tryProcessKillEvent(ctx context.Context, line string, timestamp time.Time, serverID int64) bool {
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

	// Get today's date for daily stats
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Process each killer (first gets kill, rest get assists)
	for i, killer := range killers {
		if killer.SteamID == "INVALID" {
			continue
		}

		// Upsert player
		player, err := p.queries.GetPlayerByExternalID(ctx, killer.SteamID)
		if err != nil {
			player, err = p.queries.CreatePlayer(ctx, db.CreatePlayerParams{
				ExternalID: killer.SteamID,
				Name:       killer.Name,
			})
			if err != nil {
				log.Printf("Failed to create player %s: %v", killer.Name, err)
				continue
			}
		}

		// Upsert player_stats
		playerStats, err := p.queries.GetPlayerStatsByPlayerID(ctx, player.ID)
		if err != nil {
			playerStats, err = p.queries.CreatePlayerStats(ctx, db.CreatePlayerStatsParams{
				ID:       killer.SteamID,
				PlayerID: player.ID,
				ServerID: serverID,
			})
			if err != nil {
				log.Printf("Failed to create player_stats for %s: %v", killer.Name, err)
				continue
			}
		}

		// Determine if this is a kill or assist
		one := int64(1)
		zero := int64(0)
		var kills, assists *int64
		if i == 0 {
			kills = &one
			assists = &zero
		} else {
			kills = &zero
			assists = &one
		}

		// Update lifetime weapon stats
		_, err = p.queries.UpsertWeaponStats(ctx, db.UpsertWeaponStatsParams{
			PlayerStatsID: playerStats.ID,
			WeaponName:    weapon,
			Kills:         kills,
			Assists:       assists,
		})
		if err != nil {
			log.Printf("Failed to update weapon stats: %v", err)
		}

		// Update daily weapon stats
		_, err = p.queries.UpsertDailyWeaponStats(ctx, db.UpsertDailyWeaponStatsParams{
			PlayerID:   player.ID,
			ServerID:   serverID,
			Date:       today,
			WeaponName: weapon,
			Kills:      kills,
			Assists:    assists,
		})
		if err != nil {
			log.Printf("Failed to update daily weapon stats: %v", err)
		}

		// Update daily player stats (only once per player, not per weapon)
		if i == 0 {
			_, err = p.queries.UpsertDailyPlayerStats(ctx, db.UpsertDailyPlayerStatsParams{
				PlayerID:    player.ID,
				ServerID:    serverID,
				Date:        today,
				Kills:       kills,
				Assists:     &zero,
				Deaths:      &zero,
				GamesPlayed: &zero,
				TotalScore:  &zero,
			})
		} else {
			_, err = p.queries.UpsertDailyPlayerStats(ctx, db.UpsertDailyPlayerStatsParams{
				PlayerID:    player.ID,
				ServerID:    serverID,
				Date:        today,
				Kills:       &zero,
				Assists:     assists,
				Deaths:      &zero,
				GamesPlayed: &zero,
				TotalScore:  &zero,
			})
		}
		if err != nil {
			log.Printf("Failed to update daily player stats: %v", err)
		}
	}

	return true
}

// tryProcessPlayerJoin parses and processes player join events
func (p *LogParser) tryProcessPlayerJoin(ctx context.Context, line string, timestamp time.Time, serverID int64) bool {
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
	_, err := p.queries.GetPlayerByExternalID(ctx, steamID)
	if err != nil {
		_, err := p.queries.CreatePlayer(ctx, db.CreatePlayerParams{
			ExternalID: steamID,
			Name:       playerName,
		})
		if err != nil {
			log.Printf("Failed to create player on join: %v", err)
		}
	}

	return true
}

// tryProcessPlayerDisconnect parses and processes player disconnect events
func (p *LogParser) tryProcessPlayerDisconnect(ctx context.Context, line string, timestamp time.Time, serverID int64) bool {
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
