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
func (c *Creator) CreateEvent(eventType string, serverExternalID string, data map[string]interface{}) error {
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

// CreatePlayerKillEvent creates a player kill event
// serverExternalID is the server's external_id (UUID from log file), not the PocketBase record ID
// killerSteamID, killerName, victimSteamID, victimName are raw data from log parsing
// isCatchup marks events created during startup catchup (skips scoring/RCON side effects)
func (c *Creator) CreatePlayerKillEvent(serverExternalID, killerSteamID, killerName, victimSteamID, victimName, weapon string, isHeadshot, isCatchup bool) error {
	return c.CreateEvent(TypePlayerKill, serverExternalID, map[string]interface{}{
		"killer_steam_id": killerSteamID,
		"killer_name":     killerName,
		"victim_steam_id": victimSteamID,
		"victim_name":     victimName,
		"weapon":          weapon,
		"is_headshot":     isHeadshot,
		"is_catchup":      isCatchup,
	})
}

// CreatePlayerLoginEvent creates a player login event
// This is the earliest connection event - creates/updates player record
func (c *Creator) CreatePlayerLoginEvent(serverID, playerName, steamID, platform string, isCatchup bool) error {
	return c.CreateEvent(TypePlayerLogin, serverID, map[string]interface{}{
		"player_name": playerName,
		"steam_id":    steamID,
		"platform":    platform,
		"is_catchup":  isCatchup,
	})
}

// CreatePlayerJoinEvent creates a player join event
func (c *Creator) CreatePlayerJoinEvent(serverID, playerName string, isCatchup bool) error {
	return c.CreateEvent(TypePlayerJoin, serverID, map[string]interface{}{
		"player_name": playerName,
		"is_catchup":  isCatchup,
	})
}

// CreatePlayerLeaveEvent creates a player leave event
func (c *Creator) CreatePlayerLeaveEvent(serverID, steamID, playerName string) error {
	return c.CreateEvent(TypePlayerLeave, serverID, map[string]interface{}{
		"steam_id":    steamID,
		"player_name": playerName,
	})
}

// CreateMatchStartEvent creates a match start event
func (c *Creator) CreateMatchStartEvent(serverID, matchID, mapName, scenario string, isCatchup bool) error {
	return c.CreateEvent(TypeMatchStart, serverID, map[string]interface{}{
		"match_id":   matchID,
		"map":        mapName,
		"scenario":   scenario,
		"is_catchup": isCatchup,
	})
}

// CreateMatchEndEvent creates a match end event
func (c *Creator) CreateMatchEndEvent(serverID, matchID string) error {
	return c.CreateEvent(TypeMatchEnd, serverID, map[string]interface{}{
		"match_id": matchID,
	})
}

// CreateRoundStartEvent creates a round start event
func (c *Creator) CreateRoundStartEvent(serverID, matchID string, roundNumber int) error {
	return c.CreateEvent(TypeRoundStart, serverID, map[string]interface{}{
		"match_id": matchID,
		"round":    roundNumber,
	})
}

// CreateRoundEndEvent creates a round end event
func (c *Creator) CreateRoundEndEvent(serverID, matchID string, roundNumber int, winningTeam int, isCatchup bool) error {
	return c.CreateEvent(TypeRoundEnd, serverID, map[string]interface{}{
		"match_id":     matchID,
		"round":        roundNumber,
		"winning_team": winningTeam,
		"is_catchup":   isCatchup,
	})
}

// CreateObjectiveCapturedEvent creates an objective captured event
func (c *Creator) CreateObjectiveCapturedEvent(serverID, matchID, steamID, playerName, objectiveNum string, capturingTeam int, isCatchup bool) error {
	return c.CreateEvent(TypeObjectiveCaptured, serverID, map[string]interface{}{
		"match_id":       matchID,
		"steam_id":       steamID,
		"player_name":    playerName,
		"objective":      objectiveNum,
		"capturing_team": capturingTeam,
		"is_catchup":     isCatchup,
	})
}

// CreateObjectiveDestroyedEvent creates an objective destroyed event
func (c *Creator) CreateObjectiveDestroyedEvent(serverID, matchID, steamID, playerName, objectiveNum string, destroyingTeam int, isCatchup bool) error {
	return c.CreateEvent(TypeObjectiveDestroyed, serverID, map[string]interface{}{
		"match_id":        matchID,
		"steam_id":        steamID,
		"player_name":     playerName,
		"objective":       objectiveNum,
		"destroying_team": destroyingTeam,
		"is_catchup":      isCatchup,
	})
}

// CreateChatCommandEvent creates a chat command event
func (c *Creator) CreateChatCommandEvent(serverID, steamID, playerName, command string, args []string, isCatchup bool) error {
	return c.CreateEvent(TypeChatCommand, serverID, map[string]interface{}{
		"steam_id":    steamID,
		"player_name": playerName,
		"command":     command,
		"args":        args,
		"is_catchup":  isCatchup,
	})
}

// CreateAppStartedEvent creates an app started event (no server)
func (c *Creator) CreateAppStartedEvent(version string) error {
	return c.CreateEvent(TypeAppStarted, "", map[string]interface{}{
		"version": version,
	})
}

// CreateAppShutdownEvent creates an app shutdown event (no server)
func (c *Creator) CreateAppShutdownEvent() error {
	return c.CreateEvent(TypeAppShutdown, "", map[string]interface{}{})
}
