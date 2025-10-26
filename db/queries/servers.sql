-- name: CreateServer :one
INSERT INTO servers (external_id, name, path)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetServerByID :one
SELECT * FROM servers WHERE id = ?;

-- name: GetServerByExternalID :one
SELECT * FROM servers WHERE external_id = ?;

-- name: ListServers :many
SELECT * FROM servers ORDER BY id;

-- name: UpdateServer :one
UPDATE servers
SET name = ?, path = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteServer :exec
DELETE FROM servers WHERE id = ?;
