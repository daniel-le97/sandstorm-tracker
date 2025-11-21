package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	// "log"
	"log/slog"
	"slices"
	"time"

	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/events"

	"github.com/pocketbase/pocketbase/core"
)

// GameEventHandlers handles game event processing via PocketBase hooks
type GameEventHandlers struct {
	app            AppInterface
	scoreDebouncer ScoreDebouncer
}

// ScoreDebouncer interface for triggering score updates
type ScoreDebouncer interface {
	TriggerScoreUpdate(serverID string)
	TriggerScoreUpdateFixed(serverID string, delay time.Duration)
	ExecuteImmediately(serverID string)
}

// NewGameEventHandlers creates a new game event handler
func NewGameEventHandlers(app AppInterface, scoreDebouncer ScoreDebouncer) *GameEventHandlers {
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
	case events.TypeMapLoad:
		return h.handleMapLoad(e)
	case events.TypeMapTravel:
		return h.handleMapTravel(e)
	case events.TypeGameOver:
		return h.handleGameOver(e)
	case events.TypeLogFileCreated:
		return h.handleLogFileCreated(e)
	case events.TypeObjectiveCaptured:
		return h.handleObjectiveCaptured(e)
	case events.TypeObjectiveDestroyed:
		return h.handleObjectiveDestroyed(e)
	case events.TypeChatCommand:
		return h.handleChatCommand(e)
	}

	// Not a game event we handle, continue
	return e.Next()
}

func getLogger(e *core.RecordEvent) *slog.Logger {
	return e.App.Logger().With("component", "GAME_EVENTS")
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
	log := getLogger(e)
	ctx := context.Background()

	// Extract typed data from event
	var data events.PlayerLoginData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse player login event data", "error", err)
		return e.Next()
	}

	// Get server external ID for store key lookup
	serverRecordID := e.Record.GetString("server")
	serverRecord, err := e.App.FindRecordById("servers", serverRecordID)
	if err != nil {
		log.Debug("Failed to get server record", "error", err)
		return e.Next()
	}
	serverID := serverRecord.GetString("external_id")

	log.Debug("Processing player login", "player", data.PlayerName, "steamID", data.SteamID, "platform", data.Platform)

	// Create or update player record
	_, err = database.GetOrCreatePlayerBySteamID(ctx, e.App, data.SteamID, data.PlayerName)
	if err != nil {
		log.Debug("Failed to create/update player", "steamID", data.SteamID, "error", err)
		return e.Next()
	}

	// Try to get stored IP from connection event and update player metadata
	storeKey := fmt.Sprintf("%s:lastIP", serverID)
	storedIP := e.App.Store().Get(storeKey)
	if storedIP != nil {
		ipStr, ok := storedIP.(string)
		if ok && ipStr != "" {
			// Find the player record to update it with IP in metadata
			playerRecord, err := e.App.FindFirstRecordByFilter(
				"players",
				"external_id = {:externalID}",
				map[string]any{"externalID": data.SteamID},
			)
			if err == nil && playerRecord != nil {
				// Get existing known IPs list or create new one
				existingMetadataStr := playerRecord.GetString("metadata")
				var knownIPs []string

				if existingMetadataStr != "" {
					var metadata map[string]interface{}
					if err := json.Unmarshal([]byte(existingMetadataStr), &metadata); err == nil {
						if ipsInterface, exists := metadata["knownIPs"]; exists {
							if ips, ok := ipsInterface.([]interface{}); ok {
								for _, ip := range ips {
									if ipStr, ok := ip.(string); ok {
										knownIPs = append(knownIPs, ipStr)
									}
								}
							}
						}
					}
				}

				// Add IP if not already in list
				if !slices.Contains(knownIPs, ipStr) {
					knownIPs = append(knownIPs, ipStr)

					// Build new metadata with updated knownIPs
					newMetadata := map[string]interface{}{
						"knownIPs": knownIPs,
					}
					metadataJSON, err := json.Marshal(newMetadata)
					if err != nil {
						log.Debug("Failed to marshal metadata", "error", err)
					} else {
						playerRecord.Set("metadata", string(metadataJSON))
						if err := e.App.Save(playerRecord); err != nil {
							log.Debug("Failed to update player metadata", "error", err)
						} else {
							log.Debug("Added IP to player", "ip", ipStr, "player", data.PlayerName, "knownIPs", len(knownIPs))
						}
					}
				}
			}

			// Remove the stored IP key so it's not used again
			e.App.Store().Set(storeKey, nil)
		}
	}

	log.Debug("Player record created/updated", "player", data.PlayerName, "steamID", data.SteamID)
	return e.Next()
}

// handlePlayerKill processes player kill events
// Handles regular kills, assists, friendly fire, and suicides
func (h *GameEventHandlers) handlePlayerKill(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()

	// Use Killevent proxy for all data access
	killevent := &Killevent{}
	killevent.SetProxyRecord(e.Record)

	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	killers := killevent.Killers()
	victimSteamID := killevent.VictimSteamID()
	victimName := killevent.VictimName()
	victimTeam := killevent.VictimTeam()
	weapon := killevent.Weapon()

	log.Debug("Processing kill event", "killerCount", len(killers), "victim", victimName, "weapon", weapon, "serverID", serverID)

	// Get active match for this server
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found", "serverID", serverID)
		return e.Next()
	}

	// Detect suicide (killer == victim)
	isSuicide := false
	if len(killers) > 0 && killers[0].SteamID == victimSteamID && victimSteamID != "" && victimSteamID != "INVALID" {
		isSuicide = true
	}

	// PocketBase hooks run within transactions automatically, so we use e.App directly
	// For suicides: only increment victim deaths
	if isSuicide && killevent.VictimIsPlayer() {
		victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, victimSteamID, victimName)
		if err != nil {
			log.Debug("Failed to get/create suicide victim player", "victim", victimName, "error", err)
			return e.Next()
		}

		// Upsert player into match
		if err := database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, victimPlayer.ID, nil, nil); err != nil {
			log.Debug("Failed to upsert suicide victim into match", "error", err)
			return e.Next()
		}

		// Increment deaths (only stat for suicide)
		if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, victimPlayer.ID, "deaths"); err != nil {
			log.Debug("Failed to increment deaths for suicide", "error", err)
			return e.Next()
		}

		return e.Next()
	}

	// For non-suicides: process killer(s) and victim
	for i, killer := range killers {
		if killer.SteamID == "" || killer.SteamID == "INVALID" {
			continue
		}

		killerPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, killer.SteamID, killer.Name)
		if err != nil {
			log.Debug("Failed to get/create killer player", "player", killer.Name, "error", err)
			return e.Next()
		}

		// Upsert player into match
		if err := database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, killerPlayer.ID, nil, nil); err != nil {
			log.Debug("Failed to upsert killer into match", "error", err)
			return e.Next()
		}

		// Check if this is a friendly fire kill
		isTeamKill := killer.Team == victimTeam && victimTeam >= 0 && killer.Team >= 0

		if isTeamKill {
			// Friendly fire: record incident and increment friendly_fire_kills
			if killevent.VictimIsPlayer() {
				if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, killerPlayer.ID, "friendly_fire_kills"); err != nil {
					log.Debug("Failed to increment friendly_fire_kills", "error", err)
					return e.Next()
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
				if victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, victimSteamID, victimName); err == nil {
					ff.VictimID = victimPlayer.ID
					if err := database.RecordFriendlyFireIncident(ctx, e.App, ff); err != nil {
						log.Debug("Failed to record friendly fire incident", "error", err)
					}
				}
			}
		} else if i == 0 {
			// Regular kill: first killer gets the kill credit
			if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, killerPlayer.ID, "kills"); err != nil {
				log.Debug("Failed to increment kills for killer", "error", err)
				return e.Next()
			}

			// Update weapon stats (only for primary killer)
			killCount := int64(1)
			assistCount := int64(0)
			if err := database.UpsertMatchWeaponStats(ctx, e.App, activeMatch.ID, killerPlayer.ID, weapon, &killCount, &assistCount); err != nil {
				log.Debug("Failed to update weapon stats", "error", err)
				return e.Next()
			}
		} else {
			// Regular assist: non-first killers get assist credit
			if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, killerPlayer.ID, "assists"); err != nil {
				log.Debug("Failed to increment assists", "error", err)
				return e.Next()
			}

			// Update weapon stats with assist
			killCount := int64(0)
			assistCount := int64(1)
			if err := database.UpsertMatchWeaponStats(ctx, e.App, activeMatch.ID, killerPlayer.ID, weapon, &killCount, &assistCount); err != nil {
				log.Debug("Failed to update weapon stats for assist", "error", err)
				return e.Next()
			}
		}
	}

	// Update victim stats (if not empty - could be bot death)
	if killevent.VictimIsPlayer() && !isSuicide {
		victimPlayer, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, victimSteamID, victimName)
		if err != nil {
			log.Debug("Failed to get/create victim player", "player", victimName, "error", err)
			return e.Next()
		}

		// Upsert player into match
		if err := database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, victimPlayer.ID, nil, nil); err != nil {
			log.Debug("Failed to upsert victim into match", "error", err)
			return e.Next()
		}

		// Increment deaths (always incremented except for suicides)
		if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, victimPlayer.ID, "deaths"); err != nil {
			log.Debug("Failed to increment deaths for victim", "error", err)
			return e.Next()
		}
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
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.PlayerJoinData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse player join event data", "error", err)
		return e.Next()
	}

	// Get or create player by name (may not have Steam ID yet)
	player, err := database.GetPlayerByName(ctx, e.App, data.PlayerName)
	if err != nil {
		// Create player with name only if doesn't exist
		player, err = database.CreatePlayer(ctx, e.App, "", data.PlayerName)
		if err != nil {
			log.Debug("Failed to create player", "error", err)
			return e.Next()
		}
	}

	playerID := player.ID

	log.Debug("Processing player join", "player", playerID, "server", serverID) // Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found for player join", "serverID", serverID)
		return e.Next()
	}

	// Add player to match (upsert creates row if needed)
	timestamp := e.Record.GetDateTime("created").Time()
	err = database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, playerID, nil, &timestamp)
	if err != nil {
		log.Debug("Failed to add player to match", "error", err)
		return e.Next()
	}

	log.Debug("Player added to match", "player", playerID, "match", activeMatch.ID)
	return e.Next()
}

// handlePlayerLeave processes player leave events
func (h *GameEventHandlers) handlePlayerLeave(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.PlayerLeaveData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse player leave event data", "error", err)
		return e.Next()
	}

	if data.SteamID == "" || data.SteamID == "INVALID" {
		log.Debug("Player leave event has invalid Steam ID")
		return e.Next()
	}

	// Get player by Steam ID
	player, err := database.GetPlayerByExternalID(ctx, e.App, data.SteamID)
	if err != nil || player == nil {
		log.Debug("Failed to find player with Steam ID", "steamID", data.SteamID, "error", err)
		return e.Next()
	}

	playerID := player.ID
	log.Debug("Processing player leave", "player", playerID, "server", serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found for player leave", "serverID", serverID)
		return e.Next()
	}

	// Get timestamp from event
	timestamp := e.Record.GetDateTime("created").Time()

	// Mark player as disconnected from the match
	err = database.DisconnectPlayerFromMatch(ctx, e.App, activeMatch.ID, playerID, &timestamp)
	if err != nil {
		log.Debug("Failed to disconnect player from match", "player", playerID, "match", activeMatch.ID, "error", err)
	} else {
		log.Debug("Player disconnected from match", "player", playerID, "match", activeMatch.ID)
	}

	return e.Next()
}

// handleRoundEnd processes round end events
func (h *GameEventHandlers) handleRoundEnd(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.RoundEndData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse round end event data", "error", err)
		return e.Next()
	}

	log.Debug("Processing round end", "winningTeam", data.WinningTeam, "server", serverID)

	// Get active match to increment round counter
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found for round end event", "serverID", serverID)
		return e.Next()
	}

	// Increment the round counter
	if err := database.IncrementMatchRound(ctx, e.App, activeMatch.ID); err != nil {
		log.Debug("Failed to increment round for match", "match", activeMatch.ID, "error", err)
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
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.MatchStartData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse match start event data", "error", err)
		return e.Next()
	}

	log.Debug("Match started on server", "serverID", serverID, "map", data.Map, "scenario", data.Scenario)
	return e.Next()
}

// handleMatchEnd processes match end events
func (h *GameEventHandlers) handleMatchEnd(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.MatchEndData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse match end event data", "error", err)
		return e.Next()
	}

	log.Debug("Match end event processed", "serverID", serverID)

	// Trigger immediate score update when match ends
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.ExecuteImmediately(serverID)
	}

	return e.Next()
}

// handleObjectiveCaptured processes objective captured events
func (h *GameEventHandlers) handleObjectiveCaptured(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.ObjectiveCapturedData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse objective captured event data", "error", err)
		return e.Next()
	}

	if len(data.Players) == 0 {
		log.Debug("No players in objective captured event")
		return e.Next()
	}

	log.Debug("Processing objective captured", "players", len(data.Players), "objective", data.Objective, "team", data.CapturingTeam, "server", serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found for objective captured", "serverID", serverID)
		return e.Next()
	}

	timestamp := e.Record.GetDateTime("created").Time()
	team := int64(data.CapturingTeam)

	// PocketBase hooks run within transactions automatically
	// Process all players involved in objective
	for _, p := range data.Players {
		if p.SteamID == "" || p.SteamID == "INVALID" {
			continue
		}

		player, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, p.SteamID, p.PlayerName)
		if err != nil {
			log.Debug("Failed to get/create player", "player", p.PlayerName, "error", err)
			return e.Next()
		}

		// Ensure player is in match and increment objectives_captured
		if err := database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, player.ID, &team, &timestamp); err != nil {
			log.Debug("Failed to upsert player into match", "error", err)
			return e.Next()
		}

		if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, player.ID, "objectives_captured"); err != nil {
			log.Debug("Failed to increment objectives_captured for player", "player", p.PlayerName, "error", err)
			return e.Next()
		}
	}

	// Increment round_objective counter (once per objective event, not per player)
	if err := database.IncrementMatchRoundObjective(ctx, e.App, activeMatch.ID); err != nil {
		log.Debug("Failed to increment round_objective", "error", err)
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
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.ObjectiveDestroyedData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse objective destroyed event data", "error", err)
		return e.Next()
	}

	if len(data.Players) == 0 {
		log.Debug("No players in objective destroyed event")
		return e.Next()
	}

	log.Debug("Processing objective destroyed", "players", len(data.Players), "objective", data.Objective, "team", data.DestroyingTeam, "server", serverID)

	// Get active match
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found for objective destroyed", "serverID", serverID)
		return e.Next()
	}

	timestamp := e.Record.GetDateTime("created").Time()
	team := int64(data.DestroyingTeam)

	// PocketBase hooks run within transactions automatically
	// Process all players involved in objective
	for _, p := range data.Players {
		if p.SteamID == "" || p.SteamID == "INVALID" {
			continue
		}

		player, err := database.GetOrCreatePlayerBySteamID(ctx, e.App, p.SteamID, p.PlayerName)
		if err != nil {
			log.Debug("Failed to get/create player", "player", p.PlayerName, "error", err)
			return e.Next()
		}

		// Ensure player is in match and increment objectives_destroyed
		if err := database.UpsertMatchPlayerStats(ctx, e.App, activeMatch.ID, player.ID, &team, &timestamp); err != nil {
			log.Debug("Failed to upsert player into match", "error", err)
			return e.Next()
		}

		if err := database.IncrementMatchPlayerStat(ctx, e.App, activeMatch.ID, player.ID, "objectives_destroyed"); err != nil {
			log.Debug("Failed to increment objectives_destroyed for player", "player", p.PlayerName, "error", err)
			return e.Next()
		}
	}

	// Trigger fixed 10s delay score update for objectives (outside transaction)
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.TriggerScoreUpdateFixed(serverID, 10*time.Second)
	}

	return e.Next()
}

// handleMapLoad processes map load events and creates a new match
func (h *GameEventHandlers) handleMapLoad(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.MapLoadData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse map load event data", "error", err)
		return e.Next()
	}

	// End any active match and create a new one
	// EndActiveMatchAndCreateNew expects serverID (external_id), not the record ID
	if err := database.EndActiveMatchAndCreateNew(ctx, e.App, serverID, data.Map, data.Scenario, data.Timestamp, data.PlayerTeam); err != nil {
		log.Debug("Failed to end/create match for map load", "error", err)
		return e.Next()
	}

	// Get the newly created match to emit a match_start event
	// GetActiveMatch expects the serverID (external_id), not the record ID
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("Failed to get active match after creation", "error", err)
		return e.Next()
	}

	// Emit match_start event with the new match ID
	eventsCollection, err := e.App.FindCollectionByNameOrId("events")
	if err != nil {
		log.Debug("Failed to find events collection", "error", err)
		return e.Next()
	}

	matchStartRecord := core.NewRecord(eventsCollection)
	matchStartRecord.Set("type", events.TypeMatchStart)
	matchStartRecord.Set("server", serverRecordID)
	matchStartRecord.Set("timestamp", time.Now())

	startData := events.MatchStartData{
		MatchID:   activeMatch.ID,
		Map:       data.Map,
		Scenario:  data.Scenario,
		Timestamp: data.Timestamp,
		IsCatchup: data.IsCatchup,
	}
	dataJSON, _ := json.Marshal(startData)
	matchStartRecord.Set("data", string(dataJSON))

	if err := e.App.Save(matchStartRecord); err != nil {
		log.Debug("Failed to create match_start event", "error", err)
		return e.Next()
	}

	log.Debug("Map load processed", "map", data.Map, "scenario", data.Scenario, "server", serverID, "matchID", activeMatch.ID)
	return e.Next()
}

// handleMapTravel processes map travel events and creates a new match
func (h *GameEventHandlers) handleMapTravel(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract typed data from event
	var data events.MapTravelData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse map travel event data", "error", err)
		return e.Next()
	}

	// End any active match and create a new one
	// EndActiveMatchAndCreateNew expects serverID (external_id), not the record ID
	if err := database.EndActiveMatchAndCreateNew(ctx, e.App, serverID, data.Map, data.Scenario, data.Timestamp, data.PlayerTeam); err != nil {
		log.Debug("Failed to end/create match for map travel", "error", err)
		return e.Next()
	}

	// Get the newly created match
	// GetActiveMatch expects the serverID (external_id), not the record ID
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("Failed to get active match after creation", "error", err)
		return e.Next()
	}

	// Emit match_start event with the new match ID
	eventsCollection, err := e.App.FindCollectionByNameOrId("events")
	if err != nil {
		log.Debug("Failed to find events collection", "error", err)
		return e.Next()
	}

	matchStartEvent := core.NewRecord(eventsCollection)
	matchStartEvent.Set("type", events.TypeMatchStart)
	matchStartEvent.Set("server", serverRecordID)
	matchStartEvent.Set("timestamp", time.Now())

	startData := events.MatchStartData{
		MatchID:   activeMatch.ID,
		Map:       data.Map,
		Scenario:  data.Scenario,
		Timestamp: data.Timestamp,
		IsCatchup: false,
	}
	dataJSON, _ := json.Marshal(startData)
	matchStartEvent.Set("data", string(dataJSON))

	if err := e.App.Save(matchStartEvent); err != nil {
		log.Debug("Failed to create match_start event", "error", err)
		return e.Next()
	}

	log.Debug("Map travel processed", "map", data.Map, "scenario", data.Scenario, "server", serverID)
	return e.Next()
}

// handleGameOver processes game over events and finishes the current match
// - Ends the current match gracefully
// - Sets all player match_player_stats to not connected
// - Triggers score update
func (h *GameEventHandlers) handleGameOver(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Get the active match for this server
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err != nil || activeMatch == nil {
		log.Debug("No active match found for server", "serverID", serverID)
		return e.Next()
	}

	// End the match using database helper
	endTime := time.Now()
	if err := database.EndMatch(ctx, e.App, activeMatch.ID, &endTime, nil, nil); err != nil {
		log.Debug("Failed to end match", "error", err)
		return e.Next()
	}

	// Disconnect all players from the match
	if err := database.DisconnectAllPlayersInMatch(ctx, e.App, activeMatch.ID, &endTime); err != nil {
		log.Debug("Failed to disconnect players from match", "error", err)
		return e.Next()
	}

	// Emit match_end event
	eventsCollection, err := e.App.FindCollectionByNameOrId("events")
	if err != nil {
		log.Debug("Failed to find events collection", "error", err)
		return e.Next()
	}

	matchEndEvent := core.NewRecord(eventsCollection)
	matchEndEvent.Set("type", events.TypeMatchEnd)
	matchEndEvent.Set("server", serverRecordID)
	matchEndEvent.Set("timestamp", time.Now())

	endData := events.MatchEndData{
		MatchID: activeMatch.ID,
		EndTime: endTime,
	}
	dataJSON, _ := json.Marshal(endData)
	matchEndEvent.Set("data", string(dataJSON))

	if err := e.App.Save(matchEndEvent); err != nil {
		log.Debug("Failed to create match_end event", "error", err)
		return e.Next()
	}

	log.Debug("Game over processed for server, match ended gracefully", "serverID", serverID)

	// Trigger immediate score update when match ends
	if h.scoreDebouncer != nil {
		h.scoreDebouncer.ExecuteImmediately(serverID)
	}

	return e.Next()
}

// handleLogFileCreated processes log file created events
// - Updates the server's file_creation_time field
// - Ensures no active match exists (cleans up stale matches from server crash)
func (h *GameEventHandlers) handleLogFileCreated(e *core.RecordEvent) error {
	log := getLogger(e)
	ctx := context.Background()
	serverRecordID := e.Record.GetString("server")
	serverID, err := h.getServerExternalID(ctx, serverRecordID)
	if err != nil {
		log.Debug("Failed to get server external_id", "error", err)
		return e.Next()
	}

	// Extract timestamp from event data
	var data events.LogFileCreatedData
	if err := json.Unmarshal([]byte(e.Record.GetString("data")), &data); err != nil {
		log.Debug("Failed to parse log file created event data", "error", err)
		return e.Next()
	}

	// Update the server's file_creation_time
	serverRecord, err := e.App.FindRecordById("servers", serverRecordID)
	if err != nil {
		log.Debug("Failed to find server record", "error", err)
		return e.Next()
	}

	serverRecord.Set("log_file_creation_time", data.Timestamp.Format("2006-01-02 15:04:05.000Z"))
	if err := e.App.Save(serverRecord); err != nil {
		log.Debug("Failed to update server file_creation_time", "error", err)
		return e.Next()
	}

	log.Debug("Updated log file creation time for server", "serverID", serverID, "timestamp", data.Timestamp)

	// Check if there's an active match and end it gracefully
	// (This handles the case where the server crashed mid-match)
	activeMatch, err := database.GetActiveMatch(ctx, e.App, serverID)
	if err == nil && activeMatch != nil {
		log.Debug("Found stale active match after log file creation, marking as crashed", "matchID", activeMatch.ID, "serverID", serverID)

		endTime := data.Timestamp
		crashed := "crashed"

		if err := database.EndMatch(ctx, e.App, activeMatch.ID, &endTime, nil, &crashed); err != nil {
			log.Debug("Failed to end stale match", "error", err)
		}

		if err := database.DisconnectAllPlayersInMatch(ctx, e.App, activeMatch.ID, &endTime); err != nil {
			log.Debug("Failed to disconnect players from stale match", "error", err)
		}
	}

	return e.Next()
}

// handleChatCommand processes chat command events
// Uses the functional HandleChatCommand to process the event
func (h *GameEventHandlers) handleChatCommand(e *core.RecordEvent) error {
	return HandleChatCommand(h.app.SendRconCommand)(e)
}
