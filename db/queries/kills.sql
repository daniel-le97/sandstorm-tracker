-- Kill tracking queries

-- name: InsertKill :exec
INSERT INTO kills (killer_id, victim_name, server_id, weapon_name, kill_type, match_id, created_at, multiplier)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetKillsByPlayer :many
SELECT k.id, k.killer_id, k.victim_name, k.server_id, k.weapon_name, k.kill_type, k.match_id, k.created_at,
       p.name as killer_name, p.external_id as killer_external_id
FROM kills k
LEFT JOIN players p ON k.killer_id = p.id
WHERE (k.killer_id = ? ) AND k.server_id = ?
ORDER BY k.created_at DESC
LIMIT ?;

-- name: GetKillsInTimeRange :many
SELECT k.id, k.killer_id, k.victim_name, k.server_id, k.weapon_name, k.kill_type, k.match_id, k.created_at,
       p.name as killer_name, p.external_id as killer_external_id
FROM kills k
LEFT JOIN players p ON k.killer_id = p.id
WHERE k.server_id = ? AND k.created_at BETWEEN ? AND ?
ORDER BY k.created_at DESC;

-- name: GetKillStatsByPlayer :one
SELECT 
    SUM(CASE WHEN kill_type = 0 THEN 1 ELSE 0 END) as total_kills,
    SUM(CASE WHEN kill_type = 2 THEN 1 ELSE 0 END) as team_kills,
    SUM(CASE WHEN kill_type = 1 THEN 1 ELSE 0 END) as suicides
FROM kills
WHERE killer_id = ? AND server_id = ?;

-- name: ListAllKills :many
SELECT * FROM kills ORDER BY created_at DESC;