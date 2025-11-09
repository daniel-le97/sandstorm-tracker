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
				COALESCE(SUM(kills), 0) as total_kills,
				COALESCE(SUM(deaths), 0) as total_deaths
			FROM match_player_stats
			WHERE player = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
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
				COALESCE(SUM(score), 0) as total_score,
				COALESCE(SUM(
					CASE 
						WHEN first_joined_at IS NOT NULL THEN
							CAST((julianday(COALESCE(last_left_at, datetime('now'))) - julianday(first_joined_at)) * 86400 AS INTEGER)
						ELSE 0
					END
				), 0) as total_duration_seconds
			FROM match_player_stats
			WHERE player = {:player}
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
	type rankRow struct {
		Rank         int `db:"rank"`
		TotalPlayers int `db:"total_players"`
	}

	var row rankRow

	err = pbApp.DB().
		NewQuery(`
			WITH player_scores AS (
				SELECT 
					player,
					CASE 
						WHEN total_duration_seconds > 0 
						THEN CAST(total_score AS REAL) / (CAST(total_duration_seconds AS REAL) / 60.0)
						ELSE 0
					END as score_per_min
				FROM (
					SELECT 
						player,
						SUM(score) as total_score,
						SUM(
							CASE 
								WHEN first_joined_at IS NOT NULL THEN
									CAST((julianday(COALESCE(last_left_at, datetime('now'))) - julianday(first_joined_at)) * 86400 AS INTEGER)
								ELSE 0
							END
						) as total_duration_seconds
					FROM match_player_stats
					GROUP BY player
					HAVING total_duration_seconds > 0
				)
			),
			target_player AS (
				SELECT score_per_min FROM player_scores WHERE player = {:player}
			)
			SELECT 
				(SELECT COUNT(*) + 1 FROM player_scores WHERE score_per_min > (SELECT score_per_min FROM target_player)) as rank,
				(SELECT COUNT(*) FROM player_scores) as total_players
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
		return 0, 0, err
	}

	return row.Rank, row.TotalPlayers, nil
}

// GetTopPlayersByScorePerMin returns top N players by score per minute
func GetTopPlayersByScorePerMin(ctx context.Context, pbApp core.App, limit int) ([]TopPlayer, error) {
	type playerRow struct {
		Name        string  `db:"name"`
		ScorePerMin float64 `db:"score_per_min"`
	}

	var rows []playerRow

	err := pbApp.DB().
		NewQuery(`
			SELECT 
				p.name,
				CASE 
					WHEN total_duration_seconds > 0 
					THEN CAST(total_score AS REAL) / (CAST(total_duration_seconds AS REAL) / 60.0)
					ELSE 0
				END as score_per_min
			FROM players p
			INNER JOIN (
				SELECT 
					player,
					SUM(score) as total_score,
					SUM(
						CASE 
							WHEN first_joined_at IS NOT NULL THEN
								CAST((julianday(COALESCE(last_left_at, datetime('now'))) - julianday(first_joined_at)) * 86400 AS INTEGER)
							ELSE 0
						END
					) as total_duration_seconds
				FROM match_player_stats
				GROUP BY player
				HAVING total_duration_seconds >= 60
			) stats ON p.id = stats.player
			ORDER BY score_per_min DESC
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
			SELECT weapon_name, SUM(kills) as total_kills
			FROM match_weapon_stats
			WHERE player = {:player}
			GROUP BY weapon_name
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
