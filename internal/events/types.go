package events

// Event type constants for the events collection
const (
	// Player events
	TypePlayerLogin = "player_login"
	TypePlayerKill  = "player_kill"
	TypePlayerJoin  = "player_join"
	TypePlayerLeave = "player_leave"

	// Match events
	TypeMatchStart = "match_start"
	TypeMatchEnd   = "match_end"

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
