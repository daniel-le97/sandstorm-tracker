-- name: CreatePlayer :one
INSERT INTO players (external_id, name)
VALUES (?, ?)
RETURNING *;

-- name: GetPlayerByID :one
SELECT * FROM players WHERE id = ?;

-- name: GetPlayerByExternalID :one
SELECT * FROM players WHERE external_id = ?;

-- name: ListPlayers :many
SELECT * FROM players ORDER BY id;

-- name: UpdatePlayer :one
UPDATE players
SET name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeletePlayer :exec
DELETE FROM players WHERE id = ?;
