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

-- name: GetActiveMatch :one
SELECT * FROM matches 
WHERE server_id = ? 
  AND end_time IS NULL 
ORDER BY start_time DESC 
LIMIT 1;

-- name: GetActiveMatches :many
SELECT * FROM matches 
WHERE end_time IS NULL 
ORDER BY start_time DESC;

-- name: EndMatch :exec
UPDATE matches 
SET 
    end_time = ?,
    winner_team = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: ForceEndStaleMatches :exec
-- Force-end matches that have been running for more than X hours (likely crashed)
-- Also disconnects all players in those matches
UPDATE matches 
SET 
    end_time = COALESCE(
        (SELECT MAX(updated_at) 
         FROM match_player_stats 
         WHERE match_id = matches.id),
        start_time
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE end_time IS NULL 
  AND start_time < ?;

-- name: DisconnectPlayersInMatch :exec
UPDATE match_player_stats
SET 
    is_currently_connected = 0,
    last_left_at = COALESCE(last_left_at, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND is_currently_connected = 1;
