package database

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
)

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
	type kdRow struct {
		TotalKills  int `db:"total_kills"`
		TotalDeaths int `db:"total_deaths"`
	}

	var row kdRow

	err = pbApp.DB().
		NewQuery(`
			SELECT 
				COALESCE(total_kills, 0) as total_kills,
				COALESCE(total_deaths, 0) as total_deaths
			FROM player_total_stats
			WHERE id = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
		// If player not found in view, return 0 kills and 0 deaths
		if err.Error() == "sql: no rows in result set" {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	return row.TotalKills, row.TotalDeaths, nil
}

// GetPlayerStats returns aggregated stats for a player
func GetPlayerStats(ctx context.Context, pbApp core.App, playerID string) (*PlayerStats, error) {
	type statsRow struct {
		TotalScore           int `db:"total_score"`
		TotalDurationSeconds int `db:"total_duration_seconds"`
	}

	var row statsRow

	err := pbApp.DB().
		NewQuery(`
			SELECT 
				COALESCE(total_score, 0) as total_score,
				COALESCE(total_duration_seconds, 0) as total_duration_seconds
			FROM player_total_stats
			WHERE id = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
		return nil, err
	}

	return &PlayerStats{
		TotalScore:           row.TotalScore,
		TotalDurationSeconds: row.TotalDurationSeconds,
	}, nil
}

// GetPlayerRank returns the player's rank based on score/min and total number of players
func GetPlayerRank(ctx context.Context, pbApp core.App, playerID string) (rank int, totalPlayers int, err error) {
	type playerRow struct {
		ScorePerMin float64 `db:"scorePerMin"`
	}

	var rows []playerRow

	err = pbApp.DB().
		NewQuery(`
			SELECT scorePerMin
			FROM top_players_by_score_per_min
		`).
		All(&rows)

	if err != nil {
		return 0, 0, err
	}

	totalPlayers = len(rows)

	// Query player-specific data
	type playerDataRow struct {
		ScorePerMin float64 `db:"scorePerMin"`
	}

	var playerData playerDataRow
	errPlayerData := pbApp.DB().
		NewQuery(`
			SELECT scorePerMin
			FROM top_players_by_score_per_min
			WHERE player = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&playerData)

	if errPlayerData != nil {
		return 0, totalPlayers, nil // Return 0 rank if player not found
	}

	// Count how many players have higher score_per_min
	rankCount := 1
	for _, row := range rows {
		if row.ScorePerMin > playerData.ScorePerMin {
			rankCount++
		}
	}

	return rankCount, totalPlayers, nil
}

// GetTopPlayersByScorePerMin returns top N players by score per minute
func GetTopPlayersByScorePerMin(ctx context.Context, pbApp core.App, limit int) ([]TopPlayer, error) {
	type playerRow struct {
		Name        string  `db:"name"`
		ScorePerMin float64 `db:"scorePerMin"`
	}

	var rows []playerRow

	err := pbApp.DB().
		NewQuery(`
			SELECT name, scorePerMin
			FROM top_players_by_score_per_min
			LIMIT {:limit}
		`).
		Bind(map[string]any{"limit": limit}).
		All(&rows)

	if err != nil {
		return nil, err
	}

	result := make([]TopPlayer, len(rows))
	for i, row := range rows {
		result[i] = TopPlayer{
			Name:        row.Name,
			ScorePerMin: row.ScorePerMin,
		}
	}

	return result, nil
}

// GetTopWeapons returns top N weapons for a specific player by total kills across all matches
func GetTopWeapons(ctx context.Context, pbApp core.App, playerID string, limit int) ([]TopWeapon, error) {
	type weaponRow struct {
		WeaponName string `db:"weapon_name"`
		TotalKills int    `db:"total_kills"`
	}

	var rows []weaponRow

	err := pbApp.DB().
		NewQuery(`
			SELECT weapon_name, COALESCE(total_kills, 0) as total_kills
			FROM player_weapon_stats
			WHERE player = {:player}
			ORDER BY total_kills DESC
			LIMIT {:limit}
		`).
		Bind(map[string]any{
			"player": playerID,
			"limit":  limit,
		}).
		All(&rows)

	if err != nil {
		return nil, err
	}

	result := make([]TopWeapon, len(rows))
	for i, row := range rows {
		result[i] = TopWeapon{
			Name:  row.WeaponName,
			Kills: row.TotalKills,
		}
	}

	return result, nil
}

// GetPlayerStatsAndRank returns aggregated stats for a player and their rank (single query)
func GetPlayerStatsAndRank(ctx context.Context, pbApp core.App, playerID string) (*PlayerStats, int, int, error) {
	type playerData struct {
		TotalScore           int `db:"total_score"`
		TotalDurationSeconds int `db:"total_duration_seconds"`
	}

	// Get this player's stats
	var playerRow playerData
	err := pbApp.DB().
		NewQuery(`
		SELECT total_score, total_duration_seconds
		FROM player_total_stats
		WHERE id = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&playerRow)

	if err != nil {
		// If player not found (no stats yet), return empty stats
		if err.Error() == "sql: no rows in result set" {
			return &PlayerStats{TotalScore: 0, TotalDurationSeconds: 0}, 0, 0, nil
		}
		return nil, 0, 0, err
	}

	// Get player's score_per_min from view (may not exist if < 60s playtime)
	type scoreRow struct {
		ScorePerMin float64 `db:"scorePerMin"`
	}
	var playerScore scoreRow
	err = pbApp.DB().
		NewQuery(`
		SELECT scorePerMin
		FROM top_players_by_score_per_min
		WHERE player = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&playerScore)

	if err != nil {
		// Player not in ranking view (likely < 60s playtime), use 0 for scorePerMin
		if err.Error() != "sql: no rows in result set" {
			return nil, 0, 0, err
		}
		playerScore.ScorePerMin = 0.0
	}

	// Get all players' scores to calculate rank
	var allScores []scoreRow
	err = pbApp.DB().
		NewQuery(`
		SELECT scorePerMin
		FROM top_players_by_score_per_min
		`).
		All(&allScores)

	if err != nil {
		return nil, 0, 0, err
	}

	// Count rank
	rankCount := 1
	for _, s := range allScores {
		if s.ScorePerMin > playerScore.ScorePerMin {
			rankCount++
		}
	}

	return &PlayerStats{
		TotalScore:           playerRow.TotalScore,
		TotalDurationSeconds: playerRow.TotalDurationSeconds,
	}, rankCount, len(allScores), nil
}

// GetOrCreatePlayerWithStatsAndRank finds or creates a player by SteamID, then returns aggregated stats and rank
func GetOrCreatePlayerWithStatsAndRank(ctx context.Context, pbApp core.App, steamID, name string) (*Player, *PlayerStats, int, int, error) {
	player, err := GetOrCreatePlayerBySteamID(ctx, pbApp, steamID, name)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	stats, rank, total, err := GetPlayerStatsAndRank(ctx, pbApp, player.ID)
	if err != nil {
		return player, &PlayerStats{TotalScore: 0, TotalDurationSeconds: 0}, 0, 0, err
	}

	return player, stats, rank, total, nil
}
