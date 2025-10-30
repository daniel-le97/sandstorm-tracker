-- name: UpsertDailyWeaponStats :one
INSERT INTO daily_weapon_stats (player_id, server_id, date, weapon_name, kills, assists)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(player_id, server_id, date, weapon_name) DO UPDATE SET
    kills = daily_weapon_stats.kills + excluded.kills,
    assists = daily_weapon_stats.assists + excluded.assists
RETURNING *;

-- name: GetTopWeaponsForPlayerForPeriod :many
SELECT 
    weapon_name,
    CAST(COALESCE(SUM(kills), 0) AS INTEGER) as total_kills,
    CAST(COALESCE(SUM(assists), 0) AS INTEGER) as total_assists
FROM daily_weapon_stats
WHERE player_id = ?
  AND server_id = ?
  AND date >= ?
GROUP BY weapon_name
HAVING total_kills > 0
ORDER BY total_kills DESC
LIMIT ?;

-- name: GetWeaponStatsForPlayerForPeriod :many
SELECT 
    weapon_name,
    CAST(COALESCE(SUM(kills), 0) AS INTEGER) as total_kills,
    CAST(COALESCE(SUM(assists), 0) AS INTEGER) as total_assists
FROM daily_weapon_stats
WHERE player_id = ?
  AND server_id = ?
  AND date >= ?
GROUP BY weapon_name;

-- name: DeleteOldDailyWeaponStats :exec
DELETE FROM daily_weapon_stats WHERE date < ?;
