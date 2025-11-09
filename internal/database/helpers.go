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
	} else {
		// Record already exists - player is already in the match
		// Don't increment session_count - it should only count when player first joins
		// Just ensure the record exists (no updates needed here)
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

// IncrementMatchPlayerObjectivesDestroyed increments objectives destroyed count for a player in a match
func IncrementMatchPlayerObjectivesDestroyed(ctx context.Context, pbApp core.App, matchID, playerID string) error {
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

	objectivesDestroyed := record.GetInt("objectives_destroyed")
	record.Set("objectives_destroyed", objectivesDestroyed+1)

	return pbApp.Save(record)
}

// IncrementMatchPlayerObjectivesCaptured increments objectives captured count for a player in a match
func IncrementMatchPlayerObjectivesCaptured(ctx context.Context, pbApp core.App, matchID, playerID string) error {
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

// PlayerStats represents aggregated player statistics
type PlayerStats struct {
	TotalScore           int
	TotalDurationSeconds int
}

// TopPlayer represents a player with their score/min
type TopPlayer struct {
	Name        string
	ScorePerMin float64
}

// TopWeapon represents a weapon with kill count
type TopWeapon struct {
	Name  string
	Kills int
}

// GetPlayerTotalKD returns total kills and deaths for a player across all matches
func GetPlayerTotalKD(ctx context.Context, pbApp core.App, playerID string) (kills int, deaths int, err error) {
	// Sum up kills from match_player_stats
	records, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"player = {:playerID}",
		"",
		-1,
		0,
		map[string]any{"playerID": playerID},
	)
	if err != nil {
		return 0, 0, err
	}

	totalKills := 0
	totalDeaths := 0
	for _, record := range records {
		totalKills += record.GetInt("kills")
		totalDeaths += record.GetInt("deaths")
	}

	return totalKills, totalDeaths, nil
}

// GetPlayerStats returns aggregated stats for a player
func GetPlayerStats(ctx context.Context, pbApp core.App, playerID string) (*PlayerStats, error) {
	records, err := pbApp.FindRecordsByFilter(
		"match_player_stats",
		"player = {:playerID}",
		"",
		-1,
		0,
		map[string]any{"playerID": playerID},
	)
	if err != nil {
		return nil, err
	}

	stats := &PlayerStats{}
	for _, record := range records {
		stats.TotalScore += record.GetInt("score")

		// Calculate duration from first_joined_at to last_left_at (or now if still playing)
		joinedStr := record.GetString("first_joined_at")
		leftStr := record.GetString("last_left_at")

		if joinedStr != "" {
			joined, err := time.Parse("2006-01-02 15:04:05.000Z", joinedStr)
			if err == nil {
				var left time.Time
				if leftStr != "" {
					left, err = time.Parse("2006-01-02 15:04:05.000Z", leftStr)
					if err != nil {
						left = time.Now()
					}
				} else {
					left = time.Now()
				}

				duration := left.Sub(joined)
				stats.TotalDurationSeconds += int(duration.Seconds())
			}
		}
	}

	return stats, nil
}

// GetPlayerRank returns the player's rank based on score/min and total number of players
func GetPlayerRank(ctx context.Context, pbApp core.App, playerID string) (rank int, totalPlayers int, err error) {
	// Get this player's stats
	playerStats, err := GetPlayerStats(ctx, pbApp, playerID)
	if err != nil {
		return 0, 0, err
	}

	playerScorePerMin := 0.0
	if playerStats.TotalDurationSeconds > 0 {
		playerScorePerMin = float64(playerStats.TotalScore) / (float64(playerStats.TotalDurationSeconds) / 60.0)
	}

	// Get all players who have stats
	allPlayers, err := pbApp.FindRecordsByFilter("players", "", "", -1, 0)
	if err != nil {
		return 0, 0, err
	}

	totalPlayers = 0
	rank = 1

	for _, player := range allPlayers {
		stats, err := GetPlayerStats(ctx, pbApp, player.Id)
		if err != nil || stats.TotalDurationSeconds == 0 {
			continue
		}

		totalPlayers++

		scorePerMin := float64(stats.TotalScore) / (float64(stats.TotalDurationSeconds) / 60.0)
		if scorePerMin > playerScorePerMin {
			rank++
		}
	}

	return rank, totalPlayers, nil
}

// GetTopPlayersByScorePerMin returns top N players by score per minute
func GetTopPlayersByScorePerMin(ctx context.Context, pbApp core.App, limit int) ([]TopPlayer, error) {
	allPlayers, err := pbApp.FindRecordsByFilter("players", "", "", -1, 0)
	if err != nil {
		return nil, err
	}

	type playerWithScore struct {
		name        string
		scorePerMin float64
	}

	playersWithScores := []playerWithScore{}

	for _, player := range allPlayers {
		stats, err := GetPlayerStats(ctx, pbApp, player.Id)
		if err != nil || stats.TotalDurationSeconds < 60 { // At least 1 minute of play time
			continue
		}

		scorePerMin := float64(stats.TotalScore) / (float64(stats.TotalDurationSeconds) / 60.0)
		playersWithScores = append(playersWithScores, playerWithScore{
			name:        player.GetString("name"),
			scorePerMin: scorePerMin,
		})
	}

	// Sort by score per minute (descending)
	for i := 0; i < len(playersWithScores); i++ {
		for j := i + 1; j < len(playersWithScores); j++ {
			if playersWithScores[j].scorePerMin > playersWithScores[i].scorePerMin {
				playersWithScores[i], playersWithScores[j] = playersWithScores[j], playersWithScores[i]
			}
		}
	}

	// Take top N
	result := []TopPlayer{}
	for i := 0; i < limit && i < len(playersWithScores); i++ {
		result = append(result, TopPlayer{
			Name:        playersWithScores[i].name,
			ScorePerMin: playersWithScores[i].scorePerMin,
		})
	}

	return result, nil
}

// GetTopWeapons returns top N weapons for a specific player by total kills across all matches
func GetTopWeapons(ctx context.Context, pbApp core.App, playerID string, limit int) ([]TopWeapon, error) {
	// Find all weapon stats for this player
	filter := fmt.Sprintf("player = '%s'", playerID)
	records, err := pbApp.FindRecordsByFilter("match_weapon_stats", filter, "", -1, 0)
	if err != nil {
		return nil, err
	}

	// Aggregate kills by weapon
	weaponKills := make(map[string]int)
	for _, record := range records {
		weapon := record.GetString("weapon")
		kills := record.GetInt("kills")
		weaponKills[weapon] += kills
	}

	// Convert to slice
	type weaponWithKills struct {
		name  string
		kills int
	}

	weapons := []weaponWithKills{}
	for weapon, kills := range weaponKills {
		weapons = append(weapons, weaponWithKills{name: weapon, kills: kills})
	}

	// Sort by kills (descending)
	for i := 0; i < len(weapons); i++ {
		for j := i + 1; j < len(weapons); j++ {
			if weapons[j].kills > weapons[i].kills {
				weapons[i], weapons[j] = weapons[j], weapons[i]
			}
		}
	}

	// Take top N
	result := []TopWeapon{}
	for i := 0; i < limit && i < len(weapons); i++ {
		result = append(result, TopWeapon{
			Name:  weapons[i].name,
			Kills: weapons[i].kills,
		})
	}

	return result, nil
}
