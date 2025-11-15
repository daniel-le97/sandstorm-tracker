package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/events"

	"github.com/pocketbase/pocketbase/core"
)

// GameEventHandlers handles game event processing via PocketBase hooks
type GameEventHandlers struct {
	app            core.App
	scoreDebouncer ScoreDebouncer
}

// ScoreDebouncer interface for triggering score updates
type ScoreDebouncer interface {
	TriggerScoreUpdate(serverID string)
	TriggerScoreUpdateFixed(serverID string, delay time.Duration)
	ExecuteImmediately(serverID string)
}

// NewGameEventHandlers creates a new game event handler
func NewGameEventHandlers(app core.App, scoreDebouncer ScoreDebouncer) *GameEventHandlers {
	return &GameEventHandlers{
		app:            app,
		scoreDebouncer: scoreDebouncer,
	}
}

// RegisterHooks registers all event handlers with PocketBase hooks
func (h *GameEventHandlers) RegisterHooks() {
	// Register handler for all event types
	h.app.OnRecordCreate("events").BindFunc(h.handleEvent)
}

// handleEvent routes events to specific handlers based on type
func (h *GameEventHandlers) handleEvent(e *core.RecordEvent) error {
	eventType := e.Record.GetString("type")

	switch eventType {
	case events.TypePlayerLogin:
		return h.handlePlayerLogin(e)
	case events.TypePlayerKill:
		return h.handlePlayerKill(e)
	case events.TypePlayerJoin:
		return h.handlePlayerJoin(e)
	case events.TypePlayerLeave:
		return h.handlePlayerLeave(e)
	case events.TypeRoundEnd:
		return h.handleRoundEnd(e)
	case events.TypeMatchStart:
		return h.handleMatchStart(e)
	case events.TypeMatchEnd:
		return h.handleMatchEnd(e)
	case events.TypeObjectiveCaptured:
		return h.handleObjectiveCaptured(e)
	case events.TypeObjectiveDestroyed:
		return h.handleObjectiveDestroyed(e)
	}

	// Not a game event we handle, continue
	return e.Next()
}

// getServerExternalID converts a server record ID to external_id
func (h *GameEventHandlers) getServerExternalID(ctx context.Context, serverRecordID string) (string, error) {
	if serverRecordID == "" {
		return "", fmt.Errorf("empty server record ID")
	}
	serverRecord, err := h.app.FindRecordById("servers", serverRecordID)
	if err != nil {
		return "", fmt.Errorf("server record not found: %w", err)
	}
	return serverRecord.GetString("external_id"), nil
}

// handlePlayerLogin processes player login events
// Creates or updates player record when they connect to server
func (h *GameEventHandlers) handlePlayerLogin(e *core.RecordEvent) error {
	ctx := context.Background()

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse player login event data: %v", err)
		return e.Next()
	}

	playerName, _ := data["player_name"].(string)
	steamID, _ := data["steam_id"].(string)
	platform, _ := data["platform"].(string)

	log.Printf("[HANDLER] Processing player login: %s (Steam: %s, Platform: %s)", playerName, steamID, platform)

	// Create or update player record
	_, err := database.GetOrCreatePlayerBySteamID(ctx, h.app, steamID, playerName)
	if err != nil {
		log.Printf("[HANDLER] Failed to create/update player: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Player record created/updated for %s", playerName)
	return e.Next()
}

// handlePlayerKill processes player kill events
// Replicates the logic from parser.tryProcessKillEvent
func (h *GameEventHandlers) handlePlayerKill(e *core.RecordEvent) error {
	ctx := context.Background()

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse kill event data: %v", err)
		return e.Next()
	}

	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}
	killerSteamID, _ := data["killer_steam_id"].(string)
	killerName, _ := data["killer_name"].(string)
	victimSteamID, _ := data["victim_steam_id"].(string)
	victimName, _ := data["victim_name"].(string)
	weapon, _ := data["weapon"].(string)
	// isHeadshot, _ := data["is_headshot"].(bool)

	log.Printf("[HANDLER] Processing kill event: killer=%s victim=%s weapon=%s server=%s",
		killerName, victimName, weapon, serverID)

	// Get active match for this server
	activeMatch, err := database.GetActiveMatch(ctx, h.app, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for server %s", serverID)
		return e.Next()
	}

	// Update killer stats (if not empty - could be AI kill)
	if killerSteamID != "" && killerSteamID != "INVALID" {
		// Get or create killer player by Steam ID
		killerPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, h.app, killerSteamID, killerName)
		if err != nil {
			log.Printf("[HANDLER] Failed to get/create killer player %s: %v", killerName, err)
			return e.Next()
		}

		// Upsert player into match (in case join event was missed)
		err = database.UpsertMatchPlayerStats(ctx, h.app, activeMatch.ID, killerPlayer.ID, nil, nil)
		if err != nil {
			log.Printf("[HANDLER] Failed to upsert killer into match: %v", err)
		}

		// Increment kills
		err = database.IncrementMatchPlayerStat(ctx, h.app, activeMatch.ID, killerPlayer.ID, "kills")
		if err != nil {
			log.Printf("[HANDLER] Failed to increment kills for killer: %v", err)
		}

		// Update weapon stats
		killCount := int64(1)
		assistCount := int64(0)
		err = database.UpsertMatchWeaponStats(ctx, h.app, activeMatch.ID, killerPlayer.ID, weapon, &killCount, &assistCount)
		if err != nil {
			log.Printf("[HANDLER] Failed to update weapon stats: %v", err)
		}
	}

	// Update victim stats (if not empty - could be bot death)
	if victimSteamID != "" && victimSteamID != "INVALID" {
		// Get or create victim player by Steam ID
		victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, h.app, victimSteamID, victimName)
		if err != nil {
			log.Printf("[HANDLER] Failed to get/create victim player %s: %v", victimName, err)
			return e.Next()
		}

		// Upsert player into match (in case join event was missed)
		err = database.UpsertMatchPlayerStats(ctx, h.app, activeMatch.ID, victimPlayer.ID, nil, nil)
		if err != nil {
			log.Printf("[HANDLER] Failed to upsert victim into match: %v", err)
		}

		// Increment deaths
		err = database.IncrementMatchPlayerStat(ctx, h.app, activeMatch.ID, victimPlayer.ID, "deaths")
		if err != nil {
			log.Printf("[HANDLER] Failed to increment deaths for victim: %v", err)
		}
	}

	// Trigger score update (debounced) - skip during catchup
	if h.scoreDebouncer != nil {
		isCatchup, _ := data["is_catchup"].(bool)
		if !isCatchup {
			h.scoreDebouncer.TriggerScoreUpdate(serverID)
		}
	}

	return e.Next()
}

// handlePlayerJoin processes player join events
// Creates match_player_stats record so player appears in match
func (h *GameEventHandlers) handlePlayerJoin(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse player join event data: %v", err)
		return e.Next()
	}

	playerName, _ := data["player_name"].(string)

	// Get or create player by name (may not have Steam ID yet)
	player, err := database.GetPlayerByName(ctx, h.app, playerName)
	if err != nil {
		// Create player with name only if doesn't exist
		player, err = database.CreatePlayer(ctx, h.app, "", playerName)
		if err != nil {
			log.Printf("[HANDLER] Failed to create player: %v", err)
			return e.Next()
		}
	}

	playerID := player.ID

	log.Printf("[HANDLER] Processing player join: player=%s server=%s", playerID, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, h.app, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for player join on server %s", serverID)
		return e.Next()
	}

	// Add player to match (upsert creates row if needed)
	timestamp := e.Record.GetDateTime("created").Time()
	err = database.UpsertMatchPlayerStats(ctx, h.app, activeMatch.ID, playerID, nil, &timestamp)
	if err != nil {
		log.Printf("[HANDLER] Failed to add player to match: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Player %s added to match %s", playerID, activeMatch.ID)
	return e.Next()
}

// handlePlayerLeave processes player leave events
func (h *GameEventHandlers) handlePlayerLeave(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse player leave event data: %v", err)
		return e.Next()
	}

	steamID, _ := data["steam_id"].(string)
	if steamID == "" || steamID == "INVALID" {
		log.Printf("[HANDLER] Player leave event has invalid Steam ID")
		return e.Next()
	}

	// Get player by Steam ID
	player, err := database.GetPlayerByExternalID(ctx, h.app, steamID)
	if err != nil || player == nil {
		log.Printf("[HANDLER] Failed to find player with Steam ID %s: %v", steamID, err)
		return e.Next()
	}

	playerID := player.ID
	log.Printf("[HANDLER] Processing player leave: player=%s server=%s", playerID, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, h.app, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for player leave on server %s", serverID)
		return e.Next()
	}

	// Get timestamp from event
	timestamp := e.Record.GetDateTime("created").Time()

	// Mark player as disconnected from the match
	err = database.DisconnectPlayerFromMatch(ctx, h.app, activeMatch.ID, playerID, &timestamp)
	if err != nil {
		log.Printf("[HANDLER] Failed to disconnect player %s from match %s: %v", playerID, activeMatch.ID, err)
	} else {
		log.Printf("[HANDLER] Player %s disconnected from match %s", playerID, activeMatch.ID)
	}

	return e.Next()
}

// handleRoundEnd processes round end events
func (h *GameEventHandlers) handleRoundEnd(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse round end event data: %v", err)
		return e.Next()
	}

	winningTeam, _ := data["winning_team"].(float64)

	log.Printf("[HANDLER] Processing round end: winningTeam=%d server=%s", int(winningTeam), serverID)

	// Get active match to increment round counter
	activeMatch, err := database.GetActiveMatch(ctx, h.app, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for round end event on server %s", serverID)
		return e.Next()
	}

	// Increment the round counter
	if err := database.IncrementMatchRound(ctx, h.app, activeMatch.ID); err != nil {
		log.Printf("[HANDLER] Failed to increment round for match %s: %v", activeMatch.ID, err)
	}

	// Trigger immediate score update after round end - skip during catchup
	if h.scoreDebouncer != nil {
		isCatchup, _ := data["is_catchup"].(bool)
		if !isCatchup {
			h.scoreDebouncer.ExecuteImmediately(serverID)
		}
	}

	return e.Next()
}

// handleMatchStart processes match start events
// Match creation is handled in parser (tryProcessMapLoad)
// This handler exists for potential future logic (e.g., notifications)
func (h *GameEventHandlers) handleMatchStart(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse match start event data: %v", err)
		return e.Next()
	}

	mapName, _ := data["map"].(string)
	scenario, _ := data["scenario"].(string)

	log.Printf("[HANDLER] Match started on server %s: %s (%s)", serverID, mapName, scenario)
	return e.Next()
}

// handleMatchEnd processes match end events
func (h *GameEventHandlers) handleMatchEnd(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse match end event data: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Match end event processed for server %s", serverID)

	// Trigger immediate score update when match ends - skip during catchup
	if h.scoreDebouncer != nil {
		isCatchup, _ := data["is_catchup"].(bool)
		if !isCatchup {
			h.scoreDebouncer.ExecuteImmediately(serverID)
		}
	}

	return e.Next()
}

// handleObjectiveCaptured processes objective captured events
func (h *GameEventHandlers) handleObjectiveCaptured(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse objective captured event data: %v", err)
		return e.Next()
	}

	steamID, _ := data["steam_id"].(string)
	playerName, _ := data["player_name"].(string)
	capturingTeam := int64(data["capturing_team"].(float64))

	// Get or create player
	player, err := database.GetOrCreatePlayerBySteamID(ctx, h.app, steamID, playerName)
	if err != nil {
		log.Printf("[HANDLER] Failed to get/create player %s: %v", playerName, err)
		return e.Next()
	}

	playerID := player.ID
	log.Printf("[HANDLER] Processing objective captured: player=%s team=%d server=%s",
		playerName, capturingTeam, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, h.app, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for objective captured on server %s", serverID)
		return e.Next()
	}

	// Ensure player is in match and increment objectives_captured
	timestamp := e.Record.GetDateTime("created").Time()
	if err := database.UpsertMatchPlayerStats(ctx, h.app, activeMatch.ID, playerID, &capturingTeam, &timestamp); err != nil {
		log.Printf("[HANDLER] Failed to upsert player into match: %v", err)
	}

	if err := database.IncrementMatchPlayerStat(ctx, h.app, activeMatch.ID, playerID, "objectives_captured"); err != nil {
		log.Printf("[HANDLER] Failed to increment objectives_captured: %v", err)
	}

	// Increment round_objective counter
	if err := database.IncrementMatchRoundObjective(ctx, h.app, activeMatch.ID); err != nil {
		log.Printf("[HANDLER] Failed to increment round_objective: %v", err)
	}

	// Trigger fixed 10s delay score update for objectives
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.TriggerScoreUpdateFixed(serverID, 10*time.Second)
	}

	return e.Next()
}

// handleObjectiveDestroyed processes objective destroyed events
func (h *GameEventHandlers) handleObjectiveDestroyed(e *core.RecordEvent) error {
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	// Extract data from event
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse objective destroyed event data: %v", err)
		return e.Next()
	}

	steamID, _ := data["steam_id"].(string)
	playerName, _ := data["player_name"].(string)
	destroyingTeam := int64(data["destroying_team"].(float64))

	// Get or create player
	player, err := database.GetOrCreatePlayerBySteamID(ctx, h.app, steamID, playerName)
	if err != nil {
		log.Printf("[HANDLER] Failed to get/create player %s: %v", playerName, err)
		return e.Next()
	}

	playerID := player.ID
	log.Printf("[HANDLER] Processing objective destroyed: player=%s team=%d server=%s",
		playerName, destroyingTeam, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, h.app, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for objective destroyed on server %s", serverID)
		return e.Next()
	}

	// Ensure player is in match and increment objectives_destroyed
	timestamp := e.Record.GetDateTime("created").Time()
	if err := database.UpsertMatchPlayerStats(ctx, h.app, activeMatch.ID, playerID, &destroyingTeam, &timestamp); err != nil {
		log.Printf("[HANDLER] Failed to upsert player into match: %v", err)
	}

	if err := database.IncrementMatchPlayerStat(ctx, h.app, activeMatch.ID, playerID, "objectives_destroyed"); err != nil {
		log.Printf("[HANDLER] Failed to increment objectives_destroyed: %v", err)
	}

	// Increment round_objective counter
	if err := database.IncrementMatchRoundObjective(ctx, h.app, activeMatch.ID); err != nil {
		log.Printf("[HANDLER] Failed to increment round_objective: %v", err)
	}

	// Trigger fixed 10s delay score update for objectives
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.TriggerScoreUpdateFixed(serverID, 10*time.Second)
	}

	return e.Next()
}
