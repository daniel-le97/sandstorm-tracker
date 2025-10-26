-- name: UpsertWeaponStats :one
INSERT INTO weapon_stats (player_stats_id, weapon_name, kills, assists)
VALUES (?, ?, ?, ?)
ON CONFLICT(player_stats_id, weapon_name) DO UPDATE SET
    kills = excluded.kills,
    assists = excluded.assists
RETURNING *;

-- name: GetWeaponStatsForPlayerStats :many
SELECT * FROM weapon_stats WHERE player_stats_id = ?;

-- name: DeleteWeaponStats :exec
DELETE FROM weapon_stats WHERE player_stats_id = ? AND weapon_name = ?;
