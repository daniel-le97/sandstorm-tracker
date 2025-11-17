package events

import "time"

// Event type constants for the events collection
const (
	// Player events
	TypePlayerLogin = "player_login"
	TypePlayerKill  = "player_kill"
	TypePlayerJoin  = "player_join"
	TypePlayerLeave = "player_leave"

	// Match events
	TypeMatchStart     = "match_start"
	TypeMatchEnd       = "match_end"
	TypeMapLoad        = "map_load"
	TypeMapTravel      = "map_travel"
	TypeGameOver       = "game_over"
	TypeLogFileCreated = "log_file_created"

	// Round events
	TypeRoundStart = "round_start"
	TypeRoundEnd   = "round_end"

	// Objective events
	TypeObjectiveCaptured  = "objective_captured"
	TypeObjectiveDestroyed = "objective_destroyed"

	// Chat events
	TypeChatCommand = "chat_command"

	// System events (no server relation)
	TypeAppStarted  = "app_started"
	TypeAppShutdown = "app_shutdown"
)

// PlayerLoginData represents data for a player_login event
type PlayerLoginData struct {
	PlayerName string `json:"player_name"`
	SteamID    string `json:"steam_id"`
	Platform   string `json:"platform"`
	IsCatchup  bool   `json:"is_catchup"`
}

// PlayerKillData represents data for a player_kill event
type PlayerKillData struct {
	Killers   []Killer `json:"killers"`
	Victim    Victim   `json:"victim"`
	Weapon    string   `json:"weapon"`
	IsCatchup bool     `json:"is_catchup"`
}

// Killer represents a killer in a player_kill event
type Killer struct {
	SteamID    string `json:"steam_id"`
	PlayerName string `json:"player_name"`
	Team       int    `json:"team"`
}

// Victim represents the victim in a player_kill event
type Victim struct {
	SteamID    string `json:"steam_id"`
	PlayerName string `json:"player_name"`
	Team       int    `json:"team"`
}

// PlayerJoinData represents data for a player_join event
type PlayerJoinData struct {
	PlayerName string `json:"player_name"`
	IsCatchup  bool   `json:"is_catchup"`
}

// PlayerLeaveData represents data for a player_leave event
type PlayerLeaveData struct {
	SteamID    string `json:"steam_id"`
	PlayerName string `json:"player_name"`
}

// MatchStartData represents data for a match_start event
type MatchStartData struct {
	MatchID   string    `json:"match_id"`
	Map       string    `json:"map"`
	Scenario  string    `json:"scenario"`
	Timestamp time.Time `json:"timestamp"`
	IsCatchup bool      `json:"is_catchup"`
}

// MatchEndData represents data for a match_end event
type MatchEndData struct {
	MatchID string    `json:"match_id"`
	EndTime time.Time `json:"end_time"`
}

// MapLoadData represents data for a map_load event
type MapLoadData struct {
	Map        string    `json:"map"`
	Scenario   string    `json:"scenario"`
	Timestamp  time.Time `json:"timestamp"`
	PlayerTeam *string   `json:"player_team"`
	IsCatchup  bool      `json:"is_catchup"`
}

// MapTravelData represents data for a map_travel event
type MapTravelData struct {
	Map        string    `json:"map"`
	Scenario   string    `json:"scenario"`
	Timestamp  time.Time `json:"timestamp"`
	PlayerTeam *string   `json:"player_team"`
	IsCatchup  bool      `json:"is_catchup"`
}

// GameOverData represents data for a game_over event
type GameOverData struct {
	Timestamp time.Time `json:"timestamp"`
}

// LogFileCreatedData represents data for a log_file_created event
type LogFileCreatedData struct {
	Timestamp time.Time `json:"timestamp"`
}

// RoundStartData represents data for a round_start event
type RoundStartData struct {
	MatchID     string `json:"match_id"`
	RoundNumber int    `json:"round"`
}

// RoundEndData represents data for a round_end event
type RoundEndData struct {
	MatchID     string `json:"match_id"`
	RoundNumber int    `json:"round"`
	WinningTeam int    `json:"winning_team"`
	IsCatchup   bool   `json:"is_catchup"`
}

// ObjectivePlayer represents a player involved in an objective event
type ObjectivePlayer struct {
	SteamID    string `json:"steam_id"`
	PlayerName string `json:"player_name"`
}

// ObjectiveCapturedData represents data for an objective_captured event with multiple players
type ObjectiveCapturedData struct {
	MatchID       string            `json:"match_id"`
	Players       []ObjectivePlayer `json:"players"`
	Objective     string            `json:"objective"`
	CapturingTeam int               `json:"capturing_team"`
	IsCatchup     bool              `json:"is_catchup"`
}

// ObjectiveDestroyedData represents data for an objective_destroyed event with multiple players
type ObjectiveDestroyedData struct {
	MatchID        string            `json:"match_id"`
	Players        []ObjectivePlayer `json:"players"`
	Objective      string            `json:"objective"`
	DestroyingTeam int               `json:"destroying_team"`
	IsCatchup      bool              `json:"is_catchup"`
}

// ChatCommandData represents data for a chat_command event
type ChatCommandData struct {
	SteamID    string   `json:"steam_id"`
	PlayerName string   `json:"player_name"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	IsCatchup  bool     `json:"is_catchup"`
}

// AppStartedData represents data for an app_started event
type AppStartedData struct {
	Version string `json:"version"`
}

// AppShutdownData represents data for an app_shutdown event (empty)
type AppShutdownData struct {
}
