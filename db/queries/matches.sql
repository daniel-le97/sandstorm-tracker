-- Match management queries

-- name: StartMatch :one
INSERT INTO matches (server_id, map_id, start_time)
VALUES (?, ?, ?)
RETURNING id;

-- name: EndMatch :exec
UPDATE matches
SET end_time = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND server_id = ?;

-- name: AbortMatch :exec
UPDATE matches
SET end_time = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND server_id = ?;

-- name: UpdateMatchPlayerCount :exec
-- simplified schema does not track counts here; placeholder
UPDATE matches
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND server_id = ?;

-- name: GetActiveMatches :many
SELECT * FROM matches
WHERE server_id = ? AND end_time IS NULL
ORDER BY start_time DESC;

-- name: GetMatchHistory :many
SELECT
    m.*,
    COUNT(DISTINCT mp.player_id) as participant_count
FROM matches m
LEFT JOIN match_participant mp ON m.id = mp.match_id
WHERE m.server_id = ?
GROUP BY m.id
ORDER BY m.start_time DESC
LIMIT ?;

-- name: GetMatchDetails :one
SELECT
    m.*,
    (SELECT COUNT(*) FROM match_participant WHERE match_id = m.id) as participant_count
FROM matches m
WHERE m.id = ? AND m.server_id = ?;