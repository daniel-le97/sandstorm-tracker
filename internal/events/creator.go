package events

import (
	"encoding/json"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// Creator provides methods for creating event records
type Creator struct {
	app core.App
}

// NewCreator creates a new event creator
func NewCreator(app core.App) *Creator {
	return &Creator{app: app}
}

// CreateEvent creates a new event record in the events collection
// This is a low-level method - prefer using specific typed methods below
// serverExternalID is the server's external_id (UUID), not the PocketBase record ID
// data can include "is_catchup" boolean to mark events created during catchup mode
// All player data (Steam IDs, names) should be stored in the data JSON
func (c *Creator) CreateEvent(eventType string, serverExternalID string, data interface{}) error {
	collection, err := c.app.FindCollectionByNameOrId("events")
	if err != nil {
		return fmt.Errorf("events collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("type", eventType)

	// Set server relation (optional - can be empty for system events)
	// Need to look up server record ID from external_id
	if serverExternalID != "" {
		serverRecord, err := c.app.FindFirstRecordByFilter(
			"servers",
			"external_id = {:external_id}",
			map[string]any{"external_id": serverExternalID},
		)
		if err != nil {
			return fmt.Errorf("server not found with external_id %s: %w", serverExternalID, err)
		}
		record.Set("server", serverRecord.Id)
	}

	// Set event-specific data as JSON
	if data != nil {
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}
		record.Set("data", string(dataJSON))
	}

	// Save triggers OnRecordCreate hooks automatically
	if err := c.app.Save(record); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// CreatePlayerLoginEvent creates a player login event
// This is the earliest connection event - creates/updates player record
func (c *Creator) CreatePlayerLoginEvent(serverID, playerName, steamID, platform string, isCatchup bool) error {
	data := PlayerLoginData{
		PlayerName: playerName,
		SteamID:    steamID,
		Platform:   platform,
		IsCatchup:  isCatchup,
	}
	return c.CreateEvent(TypePlayerLogin, serverID, data)
}

// CreatePlayerKillEvent creates a player kill event
func (c *Creator) CreatePlayerKillEvent(serverID string, killers []Killer, victim Victim, weapon string, isCatchup bool) error {
	data := PlayerKillData{
		Killers:   killers,
		Victim:    victim,
		Weapon:    weapon,
		IsCatchup: isCatchup,
	}
	return c.CreateEvent(TypePlayerKill, serverID, data)
}

// CreatePlayerJoinEvent creates a player join event
func (c *Creator) CreatePlayerJoinEvent(serverID, playerName string, isCatchup bool) error {
	data := PlayerJoinData{
		PlayerName: playerName,
		IsCatchup:  isCatchup,
	}
	return c.CreateEvent(TypePlayerJoin, serverID, data)
}

// CreatePlayerLeaveEvent creates a player leave event
func (c *Creator) CreatePlayerLeaveEvent(serverID, steamID, playerName string) error {
	data := PlayerLeaveData{
		SteamID:    steamID,
		PlayerName: playerName,
	}
	return c.CreateEvent(TypePlayerLeave, serverID, data)
}

// CreateMatchStartEvent creates a match start event
func (c *Creator) CreateMatchStartEvent(serverID, matchID, mapName, scenario string, isCatchup bool) error {
	data := MatchStartData{
		MatchID:   matchID,
		Map:       mapName,
		Scenario:  scenario,
		IsCatchup: isCatchup,
	}
	return c.CreateEvent(TypeMatchStart, serverID, data)
}

// CreateMatchEndEvent creates a match end event
func (c *Creator) CreateMatchEndEvent(serverID, matchID string) error {
	data := MatchEndData{
		MatchID: matchID,
	}
	return c.CreateEvent(TypeMatchEnd, serverID, data)
}

// CreateRoundStartEvent creates a round start event
func (c *Creator) CreateRoundStartEvent(serverID, matchID string, roundNumber int) error {
	data := RoundStartData{
		MatchID:     matchID,
		RoundNumber: roundNumber,
	}
	return c.CreateEvent(TypeRoundStart, serverID, data)
}

// CreateRoundEndEvent creates a round end event
func (c *Creator) CreateRoundEndEvent(serverID, matchID string, roundNumber int, winningTeam int, isCatchup bool) error {
	data := RoundEndData{
		MatchID:     matchID,
		RoundNumber: roundNumber,
		WinningTeam: winningTeam,
		IsCatchup:   isCatchup,
	}
	return c.CreateEvent(TypeRoundEnd, serverID, data)
}

// CreateObjectiveCapturedEvent creates an objective captured event with multiple players
func (c *Creator) CreateObjectiveCapturedEvent(serverID, matchID, objectiveNum string, players []ObjectivePlayer, capturingTeam int, isCatchup bool) error {
	data := ObjectiveCapturedData{
		MatchID:       matchID,
		Players:       players,
		Objective:     objectiveNum,
		CapturingTeam: capturingTeam,
		IsCatchup:     isCatchup,
	}
	return c.CreateEvent(TypeObjectiveCaptured, serverID, data)
}

// CreateObjectiveDestroyedEvent creates an objective destroyed event with multiple players
func (c *Creator) CreateObjectiveDestroyedEvent(serverID, matchID, objectiveNum string, players []ObjectivePlayer, destroyingTeam int, isCatchup bool) error {
	data := ObjectiveDestroyedData{
		MatchID:        matchID,
		Players:        players,
		Objective:      objectiveNum,
		DestroyingTeam: destroyingTeam,
		IsCatchup:      isCatchup,
	}
	return c.CreateEvent(TypeObjectiveDestroyed, serverID, data)
}

// CreateChatCommandEvent creates a chat command event
func (c *Creator) CreateChatCommandEvent(serverID, steamID, playerName, command string, args []string, isCatchup bool) error {
	data := ChatCommandData{
		SteamID:    steamID,
		PlayerName: playerName,
		Command:    command,
		Args:       args,
		IsCatchup:  isCatchup,
	}
	return c.CreateEvent(TypeChatCommand, serverID, data)
}

// CreateAppStartedEvent creates an app started event (no server)
func (c *Creator) CreateAppStartedEvent(version string) error {
	data := AppStartedData{
		Version: version,
	}
	return c.CreateEvent(TypeAppStarted, "", data)
}

// CreateAppShutdownEvent creates an app shutdown event (no server)
func (c *Creator) CreateAppShutdownEvent() error {
	data := AppShutdownData{}
	return c.CreateEvent(TypeAppShutdown, "", data)
}
