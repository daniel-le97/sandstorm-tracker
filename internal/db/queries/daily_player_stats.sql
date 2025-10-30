-- name: UpsertDailyPlayerStats :one
INSERT INTO daily_player_stats (player_id, server_id, date, kills, assists, deaths, games_played, total_score)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(player_id, server_id, date) DO UPDATE SET
    kills = daily_player_stats.kills + excluded.kills,
    assists = daily_player_stats.assists + excluded.assists,
    deaths = daily_player_stats.deaths + excluded.deaths,
    games_played = daily_player_stats.games_played + excluded.games_played,
    total_score = daily_player_stats.total_score + excluded.total_score
RETURNING *;

-- name: GetPlayerStatsForPeriod :one
SELECT 
    CAST(COALESCE(SUM(kills), 0) AS INTEGER) as total_kills,
    CAST(COALESCE(SUM(assists), 0) AS INTEGER) as total_assists,
    CAST(COALESCE(SUM(deaths), 0) AS INTEGER) as total_deaths,
    CAST(COALESCE(SUM(games_played), 0) AS INTEGER) as total_games,
    CAST(COALESCE(SUM(total_score), 0) AS INTEGER) as total_score
FROM daily_player_stats
WHERE player_id = ?
  AND server_id = ?
  AND date >= ?;

-- name: GetTopPlayersByKillsForPeriod :many
SELECT 
    p.id,
    p.name,
    CAST(COALESCE(SUM(d.kills), 0) AS INTEGER) as total_kills,
    CAST(COALESCE(SUM(d.assists), 0) AS INTEGER) as total_assists,
    CAST(COALESCE(SUM(d.deaths), 0) AS INTEGER) as total_deaths
FROM daily_player_stats d
JOIN players p ON d.player_id = p.id
WHERE d.server_id = ?
  AND d.date >= ?
GROUP BY d.player_id, p.id, p.name
ORDER BY total_kills DESC
LIMIT ?;

-- name: DeleteOldDailyPlayerStats :exec
DELETE FROM daily_player_stats WHERE date < ?;

-- name: GetPlayerDailyTrend :many
SELECT date, kills, assists, deaths, games_played
FROM daily_player_stats
WHERE player_id = ?
  AND server_id = ?
  AND date >= ?
ORDER BY date DESC;
