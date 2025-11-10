package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// Match represents a match record from PocketBase
type Match struct {
	ID         string
	ServerID   string
	Map        *string
	WinnerTeam *int64
	StartTime  *time.Time
	EndTime    *time.Time
	Mode       string
	PlayerTeam *string
	UpdatedAt  *time.Time
}

// Player represents a player record from PocketBase
type Player struct {
	ID         string
	ExternalID string
	Name       string
}

// MatchPlayerStat represents a player's stats in a match
type MatchPlayerStat struct {
	ID            string
	MatchID       string
	PlayerID      string
	Team          *int64
	FirstJoinedAt *time.Time
	UpdatedAt     *time.Time
}

// GetActiveMatch returns the current active match for a server (end_time IS NULL)
func GetActiveMatch(ctx context.Context, pbApp core.App, serverID string) (*Match, error) {
	// Find server record first
	serverRecord, err := pbApp.FindFirstRecordByFilter(
		"servers",
		"external_id = {:serverID}",
		map[string]any{"serverID": serverID},
	)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Find active match (no end_time)
	// In PocketBase, empty date fields are stored as empty strings, not NULL
	log.Printf("[DB] Looking for active match: server ID = %s, filter = 'server = %s && end_time = \"\"'", serverID, serverRecord.Id)
	matchRecord, err := pbApp.FindFirstRecordByFilter(
		"matches",
		"server = {:server} && end_time = \"\"",
		map[string]any{"server": serverRecord.Id},
	)
	if err != nil {
		// No active match found
		log.Printf("[DB] No active match found for server %s: %v", serverID, err)
		return nil, fmt.Errorf("no active match found for server %s", serverID)
	}

	log.Printf("[DB] Found active match %s for server %s", matchRecord.Id, serverID)

	match := &Match{
		ID:       matchRecord.Id,
		ServerID: matchRecord.GetString("server"),
		Mode:     matchRecord.GetString("mode"),
	}

	if mapName := matchRecord.GetString("map"); mapName != "" {
		match.Map = &mapName
	}

	if startTime := matchRecord.GetDateTime("start_time"); !startTime.IsZero() {
		t := startTime.Time()
		match.StartTime = &t
	}

	return match, nil
}

// CreateMatch creates a new match record
func CreateMatch(ctx context.Context, pbApp core.App, serverID string, mapName, mode *string, startTime *time.Time, playerTeam ...*string) (*Match, error) {
	// Find server record
	serverRecord, err := pbApp.FindFirstRecordByFilter(
		"servers",
		"external_id = {:serverID}",
		map[string]any{"serverID": serverID},
	)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	collection, err := pbApp.FindCollectionByNameOrId("matches")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("server", serverRecord.Id)
	log.Printf("[DB] Creating match for server %s (record ID: %s)", serverID, serverRecord.Id)
	if mapName != nil {
		record.Set("map", *mapName)
	}
	if mode != nil {
		record.Set("mode", *mode)
	} else {
		record.Set("mode", "Unknown")
	}
	if startTime != nil {
		record.Set("start_time", startTime.Format(time.RFC3339))
	}
	if len(playerTeam) > 0 && playerTeam[0] != nil {
		record.Set("player_team", *playerTeam[0])
	}

	if err := pbApp.Save(record); err != nil {
		return nil, err
	}

	log.Printf("[DB] Successfully created match %s for server %s", record.Id, serverID)

	match := &Match{
		ID:       record.Id,
		ServerID: serverRecord.Id,
		Mode:     record.GetString("mode"),
	}

	if mapName != nil {
		match.Map = mapName
	}
	if startTime != nil {
		match.StartTime = startTime
	}
	if len(playerTeam) > 0 && playerTeam[0] != nil {
		match.PlayerTeam = playerTeam[0]
	}

	return match, nil
}

// GetOrCreatePlayerBySteamID finds or creates a player by Steam ID
func GetOrCreatePlayerBySteamID(ctx context.Context, pbApp core.App, steamID, name string) (*Player, error) {
	// Try to find existing player
	player, err := GetPlayerByExternalID(ctx, pbApp, steamID)
	if err == nil {
		return player, nil
	}

	// Create new player
	return CreatePlayer(ctx, pbApp, steamID, name)
}

// GetPlayerByExternalID fetches a player by external_id (e.g., Steam ID)
func GetPlayerByExternalID(ctx context.Context, pbApp core.App, externalID string) (*Player, error) {
	record, err := pbApp.FindFirstRecordByFilter(
		"players",
		"external_id = {:externalID}",
		map[string]any{"externalID": externalID},
	)
	if err != nil {
		return nil, err
	}

	return &Player{
		ID:         record.Id,
		ExternalID: record.GetString("external_id"),
		Name:       record.GetString("name"),
	}, nil
}

// GetPlayerByName finds a player by their display name
func GetPlayerByName(ctx context.Context, pbApp core.App, name string) (*Player, error) {
	record, err := pbApp.FindFirstRecordByFilter(
		"players",
		"name = {:name}",
		map[string]any{"name": name},
	)
	if err != nil {
		return nil, err
	}

	return &Player{
		ID:         record.Id,
		ExternalID: record.GetString("external_id"),
		Name:       record.GetString("name"),
	}, nil
}

// CreatePlayer creates a new player record
func CreatePlayer(ctx context.Context, pbApp core.App, externalID, name string) (*Player, error) {
	collection, err := pbApp.FindCollectionByNameOrId("players")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("external_id", externalID)
	record.Set("name", name)

	if err := pbApp.Save(record); err != nil {
		return nil, err
	}

	return &Player{
		ID:         record.Id,
		ExternalID: externalID,
		Name:       name,
	}, nil
}

// UpdatePlayerName updates a player's name if it has changed
func UpdatePlayerName(ctx context.Context, pbApp core.App, player *Player, newName string) error {
	if player.Name == newName {
		return nil // No change needed
	}

	record, err := pbApp.FindRecordById("players", player.ID)
	if err != nil {
		return err
	}

	record.Set("name", newName)
	if err := pbApp.Save(record); err != nil {
		return err
	}

	player.Name = newName // Update the in-memory struct too
	return nil
}

// UpdatePlayerExternalID updates a player's external_id (Steam ID) if it's currently empty
func UpdatePlayerExternalID(ctx context.Context, pbApp core.App, player *Player, externalID string) error {
	if player.ExternalID != "" {
		return nil // Already has an external_id
	}

	record, err := pbApp.FindRecordById("players", player.ID)
	if err != nil {
		return err
	}

	record.Set("external_id", externalID)
	if err := pbApp.Save(record); err != nil {
		return err
	}

	player.ExternalID = externalID // Update the in-memory struct too
	return nil
}

// UpsertMatchPlayerStats creates or updates match player stats
func UpsertMatchPlayerStats(ctx context.Context, pbApp core.App, matchID, playerID string, team *int64, firstJoinedAt *time.Time) error {
	// Try to find existing record (always get the most recent one if duplicates exist)
	records, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		"-created", // Order by created DESC to get the latest record first
		1,          // Limit to 1 record
		0,          // No offset
		map[string]any{
			"match":  matchID,
			"player": playerID,
		},
	)

	var record *core.Record
	if err == nil && len(records) > 0 {
		record = records[0]
	}

	if err != nil {
		// Create new record - player is joining the match for the first time
		collection, err := pbApp.FindCollectionByNameOrId("match_player_stats")
		if err != nil {
			return err
		}

		record = core.NewRecord(collection)
		record.Set("match", matchID)
		record.Set("player", playerID)
		if team != nil {
			record.Set("team", *team)
		}
		if firstJoinedAt != nil {
			record.Set("first_joined_at", firstJoinedAt.Format(time.RFC3339))
		}
		record.Set("kills", 0)
		record.Set("deaths", 0)
		record.Set("assists", 0)
		record.Set("session_count", 1)
		record.Set("is_currently_connected", true)
		record.Set("objectives_destroyed", 0)
		record.Set("objectives_captured", 0)
		record.Set("status", "ongoing")
	} else {
		// Record already exists - player is already in the match
		// Don't increment session_count - it should only count when player first joins
		// Just ensure the record exists (no updates needed here)
	}

	return pbApp.Save(record)
}

// getLatestMatchPlayerStats retrieves the most recent match_player_stats record for a player in a match
// This handles cases where duplicate records may exist
func getLatestMatchPlayerStats(pbApp core.App, matchID, playerID string) (*core.Record, error) {
	records, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		"-created", // Order by created DESC to get the latest record first
		1,          // Limit to 1 record
		0,          // No offset
		map[string]any{
			"match":  matchID,
			"player": playerID,
		},
	)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	return records[0], nil
}

// IncrementMatchPlayerKills increments kill count for a player in a match
func IncrementMatchPlayerKills(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	kills := record.GetInt("kills")
	record.Set("kills", kills+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerAssists increments assist count for a player in a match
func IncrementMatchPlayerAssists(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	assists := record.GetInt("assists")
	record.Set("assists", assists+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerDeaths increments death count for a player in a match
func IncrementMatchPlayerDeaths(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	deaths := record.GetInt("deaths")
	record.Set("deaths", deaths+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerFriendlyFire increments friendly fire kills count for a player in a match
func IncrementMatchPlayerFriendlyFire(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	friendlyFireKills := record.GetInt("friendly_fire_kills")
	record.Set("friendly_fire_kills", friendlyFireKills+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerObjectivesDestroyed increments objectives destroyed count for a player in a match
func IncrementMatchPlayerObjectivesDestroyed(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	objectivesDestroyed := record.GetInt("objectives_destroyed")
	record.Set("objectives_destroyed", objectivesDestroyed+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerObjectivesCaptured increments objectives captured count for a player in a match
func IncrementMatchPlayerObjectivesCaptured(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	objectivesCaptured := record.GetInt("objectives_captured")
	record.Set("objectives_captured", objectivesCaptured+1)

	return pbApp.Save(record)
}

// UpsertMatchWeaponStats creates or updates weapon stats for a player in a match
func UpsertMatchWeaponStats(ctx context.Context, pbApp core.App, matchID, playerID, weaponName string, kills, assists *int64) error {
	// Try to find existing record
	record, err := pbApp.FindFirstRecordByFilter(
		"match_weapon_stats",
		"match = {:match} && player = {:player} && weapon_name = {:weapon}",
		map[string]any{
			"match":  matchID,
			"player": playerID,
			"weapon": weaponName,
		},
	)

	if err != nil {
		// Create new record
		collection, err := pbApp.FindCollectionByNameOrId("match_weapon_stats")
		if err != nil {
			return err
		}

		record = core.NewRecord(collection)
		record.Set("match", matchID)
		record.Set("player", playerID)
		record.Set("weapon_name", weaponName)
		record.Set("kills", 0)

		if kills != nil {
			record.Set("kills", *kills)
		}
	} else {
		// Update existing - increment kills
		currentKills := record.GetInt("kills")
		if kills != nil {
			record.Set("kills", currentKills+int(*kills))
		}
	}

	return pbApp.Save(record)
}

// GetOrCreateServer gets or creates a server record by external ID, name, and path
func GetOrCreateServer(ctx context.Context, pbApp core.App, externalID, name, path string) (string, error) {
	// Try to find existing server
	record, err := pbApp.FindFirstRecordByFilter(
		"servers",
		"external_id = {:externalID}",
		map[string]any{"externalID": externalID},
	)

	if err == nil {
		// Server exists, return its ID
		return record.Id, nil
	}

	// Server doesn't exist, create it
	collection, err := pbApp.FindCollectionByNameOrId("servers")
	if err != nil {
		return "", err
	}

	record = core.NewRecord(collection)
	record.Set("external_id", externalID)
	record.Set("name", name)
	record.Set("path", path)
	record.Set("enabled", true)

	if err := pbApp.Save(record); err != nil {
		return "", err
	}

	log.Printf("Created new server: name='%s', external_id='%s', path='%s'", name, externalID, path)
	return record.Id, nil
}

// GetAllPlayersInMatch retrieves all player stats in a match
func GetAllPlayersInMatch(ctx context.Context, pbApp core.App, matchID string) ([]MatchPlayerStat, error) {
	records, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"match = {:match}",
		"-updated",
		-1,
		0,
		map[string]any{"match": matchID},
	)

	if err != nil {
		return nil, err
	}

	var stats []MatchPlayerStat
	for _, record := range records {
		stat := MatchPlayerStat{
			ID:       record.Id,
			MatchID:  record.GetString("match"),
			PlayerID: record.GetString("player"),
		}

		if team := record.GetInt("team"); team != 0 {
			teamInt := int64(team)
			stat.Team = &teamInt
		}

		if joined := record.GetDateTime("first_joined_at"); !joined.IsZero() {
			joinedTime := joined.Time()
			stat.FirstJoinedAt = &joinedTime
		}

		if updated := record.GetDateTime("updated"); !updated.IsZero() {
			updatedTime := updated.Time()
			stat.UpdatedAt = &updatedTime
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// EndMatch updates a match with end time and winner team
func EndMatch(ctx context.Context, pbApp core.App, matchID string, endTime *time.Time, winnerTeam *int64) error {
	record, err := pbApp.FindRecordById("matches", matchID)
	if err != nil {
		return err
	}

	if endTime != nil {
		record.Set("end_time", endTime.Format("2006-01-02 15:04:05.000Z"))
	}

	if winnerTeam != nil {
		record.Set("winner_team", *winnerTeam)
	}

	if err := pbApp.Save(record); err != nil {
		return err
	}

	// Update all player stats for this match to "finished" status
	playerRecords, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"match = {:match}",
		"",
		-1,
		0,
		map[string]any{"match": matchID},
	)

	if err == nil {
		for _, playerRecord := range playerRecords {
			playerRecord.Set("status", "finished")
			if err := pbApp.Save(playerRecord); err != nil {
				log.Printf("Failed to update player status to finished: %v", err)
			}
		}
	}

	return nil
}

// DisconnectAllPlayersInMatch marks all players in a match as disconnected
func DisconnectAllPlayersInMatch(ctx context.Context, pbApp core.App, matchID string, lastLeftAt *time.Time) error {
	records, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"match = {:match}",
		"",
		-1,
		0,
		map[string]any{"match": matchID},
	)

	if err != nil {
		return err
	}

	for _, record := range records {
		if lastLeftAt != nil {
			record.Set("last_left_at", lastLeftAt.Format("2006-01-02 15:04:05.000Z"))
		}
		record.Set("status", "disconnected")
		if err := pbApp.Save(record); err != nil {
			log.Printf("Failed to disconnect player %s: %v", record.Id, err)
		}
	}

	return nil
}

// DisconnectPlayerFromMatch marks a specific player as disconnected from a match
func DisconnectPlayerFromMatch(ctx context.Context, pbApp core.App, matchID, playerID string, lastLeftAt *time.Time) error {
	// Get the latest match_player_stats record for this player
	record, err := getLatestMatchPlayerStats(pbApp, matchID, playerID)
	if err != nil {
		return err
	}

	if lastLeftAt != nil {
		record.Set("last_left_at", lastLeftAt.Format("2006-01-02 15:04:05.000Z"))
	}
	record.Set("is_currently_connected", false)
	record.Set("status", "disconnected")

	return pbApp.Save(record)
}

// DeleteMatchIfEmpty deletes a match only if it has no player stats or weapon stats
func DeleteMatchIfEmpty(ctx context.Context, pbApp core.App, matchID string) error {
	// Check for match_player_stats
	playerStats, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"match = {:match}",
		"",
		1,
		0,
		map[string]any{"match": matchID},
	)
	if err != nil {
		return fmt.Errorf("failed to check player stats: %w", err)
	}

	if len(playerStats) > 0 {
		log.Printf("Match %s has player stats, keeping it", matchID)
		return nil
	}

	// Check for match_weapon_stats
	weaponStats, err := pbApp.FindRecordsByFilter(
		"match_weapon_stats",
		"match = {:match}",
		"",
		1,
		0,
		map[string]any{"match": matchID},
	)
	if err != nil {
		return fmt.Errorf("failed to check weapon stats: %w", err)
	}

	if len(weaponStats) > 0 {
		log.Printf("Match %s has weapon stats, keeping it", matchID)
		return nil
	}

	// No stats found, safe to delete the match
	matchRecord, err := pbApp.FindRecordById("matches", matchID)
	if err != nil {
		return fmt.Errorf("failed to find match: %w", err)
	}

	if err := pbApp.Delete(matchRecord); err != nil {
		return fmt.Errorf("failed to delete match: %w", err)
	}

	log.Printf("Deleted empty match %s (no player or weapon stats)", matchID)
	return nil
}
