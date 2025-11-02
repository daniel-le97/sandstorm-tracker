package app

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
	matchRecord, err := pbApp.FindFirstRecordByFilter(
		"matches",
		"server = {:server} && end_time = ''",
		map[string]any{"server": serverRecord.Id},
	)
	if err != nil {
		// No active match found
		return nil, fmt.Errorf("no active match found for server %s", serverID)
	}

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
func CreateMatch(ctx context.Context, pbApp core.App, serverID string, mapName, mode *string, startTime *time.Time) (*Match, error) {
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

	if err := pbApp.Save(record); err != nil {
		return nil, err
	}

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

	return match, nil
}

// GetPlayerByExternalID finds a player by their Steam ID
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

// UpsertMatchPlayerStats creates or updates match player stats
func UpsertMatchPlayerStats(ctx context.Context, pbApp core.App, matchID, playerID string, team *int64, firstJoinedAt *time.Time) error {
	// Try to find existing record
	record, err := pbApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{
			"match":  matchID,
			"player": playerID,
		},
	)

	if err != nil {
		// Create new record
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
	} else {
		// Update existing - increment session_count
		sessionCount := record.GetInt("session_count")
		record.Set("session_count", sessionCount+1)
	}

	return pbApp.Save(record)
}

// IncrementMatchPlayerKills increments kill count for a player in a match
func IncrementMatchPlayerKills(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := pbApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{
			"match":  matchID,
			"player": playerID,
		},
	)
	if err != nil {
		return err
	}

	kills := record.GetInt("kills")
	record.Set("kills", kills+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerAssists increments assist count for a player in a match
func IncrementMatchPlayerAssists(ctx context.Context, pbApp core.App, matchID, playerID string) error {
	record, err := pbApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{
			"match":  matchID,
			"player": playerID,
		},
	)
	if err != nil {
		return err
	}

	assists := record.GetInt("assists")
	record.Set("assists", assists+1)

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

// GetOrCreateServer gets or creates a server record by external ID and path
func GetOrCreateServer(ctx context.Context, pbApp core.App, externalID, path string) (string, error) {
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
	record.Set("log_path", path)
	record.Set("enabled", true)

	if err := pbApp.Save(record); err != nil {
		return "", err
	}

	log.Printf("Created new server with external_id: %s, path: %s", externalID, path)
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

	return pbApp.Save(record)
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
		if err := pbApp.Save(record); err != nil {
			log.Printf("Failed to disconnect player %s: %v", record.Id, err)
		}
	}

	return nil
}
