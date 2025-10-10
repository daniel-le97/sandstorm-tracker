-- Player session management queries

-- name: StartSession :one
-- simplified: track active players via match_participant entries with null leave_time
INSERT INTO match_participant (player_id, match_id, join_time, team)
VALUES (?, ?, ?, ?)
RETURNING id;

-- name: EndSession :exec
UPDATE match_participant
SET leave_time = ?
WHERE id = ?;

-- name: GetActiveSessions :many
SELECT mp.*, p.name as player_name, p.external_id as external_id
FROM match_participant mp
JOIN matches m ON mp.match_id = m.id
JOIN players p ON mp.player_id = p.id
WHERE m.server_id = ? AND mp.leave_time IS NULL;

-- name: GetPlayerSessions :many
SELECT mp.*, maps.map_name, m.mode
FROM match_participant mp
JOIN matches m ON mp.match_id = m.id
LEFT JOIN maps ON m.map_id = maps.id
WHERE mp.player_id = ? AND m.server_id = ?
ORDER BY mp.join_time DESC
LIMIT ?;