-- name: CreateMatch :one
INSERT INTO matches (
    server_id, map, winner_team, start_time, end_time, mode
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMatchByID :one
SELECT * FROM matches WHERE id = ?;

-- name: ListMatches :many
SELECT * FROM matches ORDER BY id DESC;

-- name: DeleteMatch :exec
DELETE FROM matches WHERE id = ?;
