// GetMultiplier returns the multiplier for this kill event
package events

import (
	"time"
)

// KillType represents different types of kills
type KillType string

const (
	KillTypeRegular       KillType = "player_kill" // Regular player kill
	KillTypeAI            KillType = "player_kill" // AI kill (also counted as player kill)
	KillTypeFriendlyFire  KillType = "team_kill"   // Team kill / friendly fire
	KillTypeSuicide       KillType = "suicide"     // Suicide
	KillTypeEnvironmental KillType = "player_kill" // Environmental kill
)

// Killer represents a player who participated in a kill
type Killer struct {
	Name    string `json:"name"`
	SteamID string `json:"steam_id"`
	Team    int    `json:"team"`
}

// GameEvent represents a parsed game event
type GameEvent struct {
	ID         string         `json:"id"`
	Type       EventType      `json:"type"`
	Timestamp  time.Time      `json:"timestamp"`
	ServerID   string         `json:"server_id"`
	Data       map[string]any `json:"data"`
	RawLogLine string         `json:"raw_log_line"`
}

// PlayerStats represents aggregated player statistics
type PlayerStats struct {
	SteamID           string    `json:"steam_id"`
	PlayerName        string    `json:"player_name"`
	TotalKills        int       `json:"total_kills"`
	TotalDeaths       int       `json:"total_deaths"`
	FriendlyFireKills int       `json:"friendly_fire_kills"`
	Suicides          int       `json:"suicides"`
	SessionStart      time.Time `json:"session_start"`
	LastSeen          time.Time `json:"last_seen"`
	KDR               float64   `json:"kdr"`
}

// WeaponStats represents weapon usage statistics
type WeaponStats struct {
	PlayerID   string    `json:"player_id"`
	WeaponName string    `json:"weapon_name"`
	KillCount  int       `json:"kill_count"`
	LastUsed   time.Time `json:"last_used"`
}

// GameSession represents a game session
type GameSession struct {
	SessionID    string    `json:"session_id"`
	ServerID     string    `json:"server_id"`
	MapName      string    `json:"map_name"`
	Scenario     string    `json:"scenario"`
	StartTime    time.Time `json:"start_time"`
	CurrentRound int       `json:"current_round"`
	Difficulty   float64   `json:"difficulty"`
	Active       bool      `json:"active"`
}

// KillEvent represents a kill event with full details
type KillEvent struct {
	KillID        string    `json:"kill_id"`
	Timestamp     time.Time `json:"timestamp"`
	ServerID      string    `json:"server_id"`
	KillerSteamID string    `json:"killer_steam_id"`
	KillerName    string    `json:"killer_name"`
	KillerTeam    int       `json:"killer_team"`
	VictimSteamID string    `json:"victim_steam_id"`
	VictimName    string    `json:"victim_name"`
	VictimTeam    int       `json:"victim_team"`
	Weapon        string    `json:"weapon"`
	KillType      KillType  `json:"kill_type"`
	Distance      float64   `json:"distance,omitempty"`
	MapName       string    `json:"map_name"`
	RoundNumber   int       `json:"round_number"`
	Multiplier    float64   `json:"multiplier"`
}

func (k *KillEvent) GetMultiplier() float64 {
	return k.Multiplier
}

// ChatCommand represents a chat command from a player
type ChatCommand struct {
	CommandID  string    `json:"command_id"`
	Timestamp  time.Time `json:"timestamp"`
	PlayerName string    `json:"player_name"`
	SteamID    string    `json:"steam_id"`
	Command    string    `json:"command"`
	Arguments  []string  `json:"arguments"`
	RawMessage string    `json:"raw_message"`
}

// CalculateKDR calculates the kill-death ratio
func (p *PlayerStats) CalculateKDR() float64 {
	if p.TotalDeaths == 0 {
		if p.TotalKills == 0 {
			return 0.0
		}
		return float64(p.TotalKills)
	}
	return float64(p.TotalKills) / float64(p.TotalDeaths)
}

// PlayerInfo represents current player information from RCON
type PlayerInfo struct {
	Name    string `json:"name"`
	SteamID string `json:"steam_id"`
	Team    int    `json:"team"`
}

// IsActive checks if the session is currently active
func (s *GameSession) IsActive() bool {
	return s.Active && time.Since(s.StartTime) < 24*time.Hour
}
