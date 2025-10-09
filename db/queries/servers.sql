-- Server management queries

-- name: UpsertServer :one
INSERT INTO servers (external_id, name, path)
VALUES (?, ?, ?)
ON CONFLICT(external_id) DO UPDATE SET
    name = excluded.name,
    path = excluded.path,
    updated_at = CURRENT_TIMESTAMP
RETURNING id;

-- name: GetServerByUuid :one
SELECT id, external_id, name, path, created_at, updated_at
FROM servers 
WHERE external_id = ?;

-- name: GetServerByConfigId :one
SELECT id, external_id, name, path
FROM servers 
WHERE external_id = ?;

-- name: GetAllServers :many
SELECT id, external_id, name
FROM servers 
ORDER BY name;

-- name: UpdateServerEnabled :exec
UPDATE servers 
SET updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;