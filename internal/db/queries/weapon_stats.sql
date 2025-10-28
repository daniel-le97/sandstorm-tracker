-- name: UpsertWeaponStats :one
INSERT INTO weapon_stats (player_stats_id, weapon_name, kills, assists)
VALUES (?, ?, ?, ?)
ON CONFLICT(player_stats_id, weapon_name) DO UPDATE SET
    kills = weapon_stats.kills + excluded.kills,
    assists = weapon_stats.assists + excluded.assists
RETURNING *;

-- name: GetWeaponStatsForPlayerStats :many
SELECT * FROM weapon_stats WHERE player_stats_id = ?;

-- name: DeleteWeaponStats :exec
DELETE FROM weapon_stats WHERE player_stats_id = ? AND weapon_name = ?;

-- name: GetTotalKillsForPlayerStats :one
SELECT CAST(COALESCE(SUM(kills), 0) AS INTEGER) as total_kills FROM weapon_stats WHERE player_stats_id = ?;

-- name: GetPlayerStatsWithKills :many
SELECT ps.*, p.name as player_name, CAST(COALESCE(SUM(ws.kills), 0) AS INTEGER) as total_kills
FROM player_stats ps
JOIN players p ON ps.player_id = p.id
LEFT JOIN weapon_stats ws ON ps.id = ws.player_stats_id
WHERE ps.server_id = ?
GROUP BY ps.id, p.name
ORDER BY total_kills DESC;

-- name: GetTopWeaponsForPlayer :many
SELECT weapon_name, kills, assists 
FROM weapon_stats 
WHERE player_stats_id = ? AND kills > 0
ORDER BY kills DESC
LIMIT ?;
