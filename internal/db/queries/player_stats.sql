
-- name: GetTopPlayersByScorePerMin :many
SELECT ps.*, p.name as player_name
FROM player_stats ps
JOIN players p ON ps.player_id = p.id
WHERE ps.server_id = ?
AND ps.total_play_time > 0
ORDER BY (CAST(ps.total_score AS REAL) / (ps.total_play_time / 60.0)) DESC
LIMIT 3;
-- name: CreatePlayerStats :one
INSERT INTO player_stats (
    id, player_id, server_id, games_played, wins, losses, total_score,
    total_play_time, last_login, total_kills, total_deaths, friendly_fire_kills, highest_score
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPlayerStatsByID :one
SELECT * FROM player_stats WHERE id = ?;

-- name: GetPlayerStatsByPlayerID :one
SELECT * FROM player_stats WHERE player_id = ?;

-- name: UpdatePlayerStats :one
UPDATE player_stats
SET games_played = ?, wins = ?, losses = ?, total_score = ?, total_play_time = ?,
    last_login = ?, total_kills = ?, total_deaths = ?, friendly_fire_kills = ?, highest_score = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlayerStats :exec
DELETE FROM player_stats WHERE id = ?;
