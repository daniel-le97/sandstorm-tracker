-- Player management queries

-- name: UpsertPlayer :one
INSERT INTO players (external_id, name, created_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(external_id) DO UPDATE SET
    name = excluded.name,
    updated_at = CURRENT_TIMESTAMP
RETURNING id;

-- name: GetPlayer :one
SELECT * FROM players 
WHERE external_id = ?;

-- name: GetPlayerByName :one
SELECT * FROM players 
WHERE name = ?;

-- name: GetPlayerGlobal :many
SELECT * FROM players 
WHERE external_id = ?;

-- name: UpdatePlayerSteamId :exec
UPDATE players 
SET external_id = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: UpdatePlayerPlaytime :exec
-- note: simplified schema does not track total_playtime_minutes by default
-- keep a no-op placeholder to preserve call sites
UPDATE players
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: ListAllPlayers :many
SELECT * FROM players ORDER BY name;