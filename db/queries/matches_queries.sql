-- Match participant and map queries

-- name: AddMatchParticipant :one
INSERT INTO match_participant (match_id, player_id, join_time, team)
VALUES (?, ?, ?, ?)
ON CONFLICT(match_id, player_id) DO UPDATE SET
    join_time = excluded.join_time,
    team = excluded.team
RETURNING id;

-- name: EndMatchParticipant :exec
UPDATE match_participant
SET leave_time = ?, team = ?
WHERE match_id = ? AND player_id = ?;

-- name: GetMatchParticipants :many
SELECT
    mp.*,
    p.name as player_name,
    p.external_id as external_id
FROM match_participant mp
JOIN players p ON mp.player_id = p.id
WHERE mp.match_id = ?
ORDER BY mp.join_time ASC;

-- name: AddMatchMap :one
-- match maps not used in simplified schema; keep placeholder
INSERT INTO maps (map_name, scenario)
VALUES (?, ?)
ON CONFLICT(map_name) DO UPDATE SET
    scenario = excluded.scenario
RETURNING id;

-- name: EndMatchMap :exec
UPDATE maps SET updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: GetMatchMaps :many
-- match_maps no longer present in simplified schema; return empty set (placeholder)
SELECT id, map_name, scenario FROM maps WHERE 0=1;

-- name: GetPlayerMatchHistory :many
SELECT
    m.id as match_id,
    m.map,
    m.mode,
    m.start_time,
    m.end_time,
    mp.join_time,
    mp.leave_time,
    mp.team
FROM matches m
JOIN match_participant mp ON m.id = mp.match_id
WHERE mp.player_id = (SELECT id FROM players WHERE external_id = ?)
AND m.server_id = ?
ORDER BY m.start_time DESC
LIMIT ?;