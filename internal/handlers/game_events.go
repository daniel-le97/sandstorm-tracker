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

	// Extract typed data from event
	var data events.PlayerLoginData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse player login event data: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Processing player login: %s (Steam: %s, Platform: %s)", data.PlayerName, data.SteamID, data.Platform)

	// Create or update player record
	_, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, data.SteamID, data.PlayerName)
	if err != nil {
		log.Printf("[HANDLER] Failed to create/update player: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Player record created/updated for %s", data.PlayerName)
	return e.Next()
}

// handlePlayerKill processes player kill events
// Handles regular kills, assists, friendly fire, and suicides
func (h *GameEventHandlers) handlePlayerKill(e *core.RecordEvent) error {
	ctx := context.Background()

	// Use Killevent proxy for all data access
	killevent := &Killevent{}
	killevent.SetProxyRecord(e.Record)

	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Printf("[HANDLER] Failed to get server external_id: %v", err)
		return e.Next()
	}

	killers := killevent.Killers()
	victimSteamID := killevent.VictimSteamID()
	victimName := killevent.VictimName()
	victimTeam := killevent.VictimTeam()
	weapon := killevent.Weapon()

	log.Printf("[HANDLER] Processing kill event: %d killers, victim=%s weapon=%s server=%s",
		len(killers), victimName, weapon, serverID)

	// Get active match for this server
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for server %s", serverID)
		return e.Next()
	}

	// Detect suicide (killer == victim)
	isSuicide := false
	if len(killers) > 0 && killers[0].SteamID == victimSteamID && victimSteamID != "" && victimSteamID != "INVALID" {
		isSuicide = true
	}

	// Use transaction to ensure all stats are updated atomically
	if err := e.App.RunInTransaction(func(txApp core.App) error {
		// For suicides: only increment victim deaths
		if isSuicide && killevent.VictimIsPlayer() {
			victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, txApp, victimSteamID, victimName)
			if err != nil {
				log.Printf("[HANDLER] Failed to get/create suicide victim player %s: %v", victimName, err)
				return err
			}

			// Upsert player into match
			if err := database.UpsertMatchPlayerStats(ctx, txApp, activeMatch.ID, victimPlayer.ID, nil, nil); err != nil {
				log.Printf("[HANDLER] Failed to upsert suicide victim into match: %v", err)
				return err
			}

			// Increment deaths (only stat for suicide)
			if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, victimPlayer.ID, "deaths"); err != nil {
				log.Printf("[HANDLER] Failed to increment deaths for suicide: %v", err)
				return err
			}

			return nil
		}

		// For non-suicides: process killer(s) and victim
		for i, killer := range killers {
			if killer.SteamID == "" || killer.SteamID == "INVALID" {
				continue
			}

			killerPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, txApp, killer.SteamID, killer.Name)
			if err != nil {
				log.Printf("[HANDLER] Failed to get/create killer player %s: %v", killer.Name, err)
				return err
			}

			// Upsert player into match
			if err := database.UpsertMatchPlayerStats(ctx, txApp, activeMatch.ID, killerPlayer.ID, nil, nil); err != nil {
				log.Printf("[HANDLER] Failed to upsert killer into match: %v", err)
				return err
			}

			// Check if this is a friendly fire kill
			isTeamKill := killer.Team == victimTeam && victimTeam >= 0 && killer.Team >= 0

			if isTeamKill {
				// Friendly fire: record incident and increment friendly_fire_kills
				if killevent.VictimIsPlayer() {
					if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, killerPlayer.ID, "friendly_fire_kills"); err != nil {
						log.Printf("[HANDLER] Failed to increment friendly_fire_kills: %v", err)
						return err
					}

					// Record friendly fire incident
					killerTeam := killer.Team
					ff := &database.FriendlyFireIncident{
						MatchID:    activeMatch.ID,
						KillerID:   killerPlayer.ID,
						VictimID:   "", // Will be fetched if victim is player
						Weapon:     weapon,
						Timestamp:  e.Record.GetDateTime("created").Time(),
						KillerTeam: &killerTeam,
						VictimTeam: &victimTeam,
					}

					// Get victim player for FF record
					if victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, txApp, victimSteamID, victimName); err == nil {
						ff.VictimID = victimPlayer.ID
						if err := database.RecordFriendlyFireIncident(ctx, txApp, ff); err != nil {
							log.Printf("[HANDLER] Failed to record friendly fire incident: %v", err)
						}
					}
				}
			} else if i == 0 {
				// Regular kill: first killer gets the kill credit
				if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, killerPlayer.ID, "kills"); err != nil {
					log.Printf("[HANDLER] Failed to increment kills for killer: %v", err)
					return err
				}

				// Update weapon stats (only for primary killer)
				killCount := int64(1)
				assistCount := int64(0)
				if err := database.UpsertMatchWeaponStats(ctx, txApp, activeMatch.ID, killerPlayer.ID, weapon, &killCount, &assistCount); err != nil {
					log.Printf("[HANDLER] Failed to update weapon stats: %v", err)
					return err
				}
			} else {
				// Regular assist: non-first killers get assist credit
				if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, killerPlayer.ID, "assists"); err != nil {
					log.Printf("[HANDLER] Failed to increment assists: %v", err)
					return err
				}

				// Update weapon stats with assist
				killCount := int64(0)
				assistCount := int64(1)
				if err := database.UpsertMatchWeaponStats(ctx, txApp, activeMatch.ID, killerPlayer.ID, weapon, &killCount, &assistCount); err != nil {
					log.Printf("[HANDLER] Failed to update weapon stats for assist: %v", err)
					return err
				}
			}
		}

		// Update victim stats (if not empty - could be bot death)
		if killevent.VictimIsPlayer() && !isSuicide {
			victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, txApp, victimSteamID, victimName)
			if err != nil {
				log.Printf("[HANDLER] Failed to get/create victim player %s: %v", victimName, err)
				return err
			}

			// Upsert player into match
			if err := database.UpsertMatchPlayerStats(ctx, txApp, activeMatch.ID, victimPlayer.ID, nil, nil); err != nil {
				log.Printf("[HANDLER] Failed to upsert victim into match: %v", err)
				return err
			}

			// Increment deaths (always incremented except for suicides)
			if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, victimPlayer.ID, "deaths"); err != nil {
				log.Printf("[HANDLER] Failed to increment deaths for victim: %v", err)
				return err
			}
		}

		return nil
	}); err != nil {
		log.Printf("[HANDLER] Transaction failed for kill event: %v", err)
		return e.Next()
	}

	// Trigger score update (debounced) - skip during catchup (outside transaction)
	if h.scoreDebouncer != nil {
		if !killevent.IsCatchup() {
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

	// Extract typed data from event
	var data events.PlayerJoinData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse player join event data: %v", err)
		return e.Next()
	}

	// Get or create player by name (may not have Steam ID yet)
	player, err := database.GetPlayerByName(ctx, e.App, data.PlayerName)
	if err != nil {
		// Create player with name only if doesn't exist
		player, err = database.CreatePlayer(ctx, e.App, "", data.PlayerName)
		if err != nil {
			log.Printf("[HANDLER] Failed to create player: %v", err)
			return e.Next()
		}
	}

	playerID := player.ID

	log.Printf("[HANDLER] Processing player join: player=%s server=%s", playerID, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for player join on server %s", serverID)
		return e.Next()
	}

	// Add player to match (upsert creates row if needed)
	timestamp := e.Record.GetDateTime("created").Time()
	err = database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, playerID, nil, &timestamp)
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

	// Extract typed data from event
	var data events.PlayerLeaveData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse player leave event data: %v", err)
		return e.Next()
	}

	if data.SteamID == "" || data.SteamID == "INVALID" {
		log.Printf("[HANDLER] Player leave event has invalid Steam ID")
		return e.Next()
	}

	// Get player by Steam ID
	player, err := database.GetPlayerByExternalID(ctx, e.App, data.SteamID)
	if err != nil || player == nil {
		log.Printf("[HANDLER] Failed to find player with Steam ID %s: %v", data.SteamID, err)
		return e.Next()
	}

	playerID := player.ID
	log.Printf("[HANDLER] Processing player leave: player=%s server=%s", playerID, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for player leave on server %s", serverID)
		return e.Next()
	}

	// Get timestamp from event
	timestamp := e.Record.GetDateTime("created").Time()

	// Mark player as disconnected from the match
	err = database.DisconnectPlayerFromMatch(ctx, e.App, activeMatch.ID, playerID, &timestamp)
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

	// Extract typed data from event
	var data events.RoundEndData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse round end event data: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Processing round end: winningTeam=%d server=%s", data.WinningTeam, serverID)

	// Get active match to increment round counter
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for round end event on server %s", serverID)
		return e.Next()
	}

	// Increment the round counter
	if err := database.IncrementMatchRound(ctx, e.App, activeMatch.ID); err != nil {
		log.Printf("[HANDLER] Failed to increment round for match %s: %v", activeMatch.ID, err)
	}

	// Trigger immediate score update after round end - skip during catchup
	if h.scoreDebouncer != nil {
		if !data.IsCatchup {
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

	// Extract typed data from event
	var data events.MatchStartData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse match start event data: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Match started on server %s: %s (%s)", serverID, data.Map, data.Scenario)
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

	// Extract typed data from event
	var data events.MatchEndData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse match end event data: %v", err)
		return e.Next()
	}

	log.Printf("[HANDLER] Match end event processed for server %s", serverID)

	// Trigger immediate score update when match ends
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.ExecuteImmediately(serverID)
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

	// Extract typed data from event
	var data events.ObjectiveCapturedData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse objective captured event data: %v", err)
		return e.Next()
	}

	if len(data.Players) == 0 {
		log.Printf("[HANDLER] No players in objective captured event")
		return e.Next()
	}

	log.Printf("[HANDLER] Processing objective captured: %d players, objective=%s, team=%d, server=%s",
		len(data.Players), data.Objective, data.CapturingTeam, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for objective captured on server %s", serverID)
		return e.Next()
	}

	timestamp := e.Record.GetDateTime("created").Time()
	team := int64(data.CapturingTeam)

	// Use transaction to ensure all player stats are updated atomically
	if err := e.App.RunInTransaction(func(txApp core.App) error {
		// Process all players involved in objective
		for _, p := range data.Players {
			if p.SteamID == "" || p.SteamID == "INVALID" {
				continue
			}

			player, err := database.GetOrCreatePlayerBySteamID(ctx, txApp, p.SteamID, p.PlayerName)
			if err != nil {
				log.Printf("[HANDLER] Failed to get/create player %s: %v", p.PlayerName, err)
				return err
			}

			// Ensure player is in match and increment objectives_captured
			if err := database.UpsertMatchPlayerStats(ctx, txApp, activeMatch.ID, player.ID, &team, &timestamp); err != nil {
				log.Printf("[HANDLER] Failed to upsert player into match: %v", err)
				return err
			}

			if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, player.ID, "objectives_captured"); err != nil {
				log.Printf("[HANDLER] Failed to increment objectives_captured for player %s: %v", p.PlayerName, err)
				return err
			}
		}

		// Increment round_objective counter (once per objective event, not per player)
		if err := database.IncrementMatchRoundObjective(ctx, txApp, activeMatch.ID); err != nil {
			log.Printf("[HANDLER] Failed to increment round_objective: %v", err)
			return err
		}

		return nil
	}); err != nil {
		log.Printf("[HANDLER] Transaction failed for objective captured: %v", err)
		return e.Next()
	}

	// Trigger fixed 10s delay score update for objectives (outside transaction)
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

	// Extract typed data from event
	var data events.ObjectiveDestroyedData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Printf("[HANDLER] Failed to parse objective destroyed event data: %v", err)
		return e.Next()
	}

	if len(data.Players) == 0 {
		log.Printf("[HANDLER] No players in objective destroyed event")
		return e.Next()
	}

	log.Printf("[HANDLER] Processing objective destroyed: %d players, objective=%s, team=%d, server=%s",
		len(data.Players), data.Objective, data.DestroyingTeam, serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Printf("[HANDLER] No active match found for objective destroyed on server %s", serverID)
		return e.Next()
	}

	timestamp := e.Record.GetDateTime("created").Time()
	team := int64(data.DestroyingTeam)

	// Use transaction to ensure all player stats are updated atomically
	if err := e.App.RunInTransaction(func(txApp core.App) error {
		// Process all players involved in objective
		for _, p := range data.Players {
			if p.SteamID == "" || p.SteamID == "INVALID" {
				continue
			}

			player, err := database.GetOrCreatePlayerBySteamID(ctx, txApp, p.SteamID, p.PlayerName)
			if err != nil {
				log.Printf("[HANDLER] Failed to get/create player %s: %v", p.PlayerName, err)
				return err
			}

			// Ensure player is in match and increment objectives_destroyed
			if err := database.UpsertMatchPlayerStats(ctx, txApp, activeMatch.ID, player.ID, &team, &timestamp); err != nil {
				log.Printf("[HANDLER] Failed to upsert player into match: %v", err)
				return err
			}

			if err := database.IncrementMatchPlayerStat(ctx, txApp, activeMatch.ID, player.ID, "objectives_destroyed"); err != nil {
				log.Printf("[HANDLER] Failed to increment objectives_destroyed for player %s: %v", p.PlayerName, err)
				return err
			}
		}

		return nil
	}); err != nil {
		log.Printf("[HANDLER] Transaction failed for objective destroyed: %v", err)
		return e.Next()
	}

	// Trigger fixed 10s delay score update for objectives (outside transaction)
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.TriggerScoreUpdateFixed(serverID, 10*time.Second)
	}

	return e.Next()
}
