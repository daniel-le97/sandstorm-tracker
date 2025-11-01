-- name: UpsertMatchPlayerStats :exec
INSERT INTO match_player_stats (
    match_id, player_id, team, first_joined_at, session_count, is_currently_connected
)
VALUES (?, ?, ?, ?, 1, 1)
ON CONFLICT(match_id, player_id) DO UPDATE SET
    session_count = session_count + 1,
    is_currently_connected = 1,
    updated_at = CURRENT_TIMESTAMP;

-- name: UpdateMatchPlayerDisconnect :exec
UPDATE match_player_stats
SET 
    is_currently_connected = 0,
    last_left_at = ?,
    total_play_time = total_play_time + ?,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: IncrementMatchPlayerKills :exec
UPDATE match_player_stats
SET 
    kills = kills + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: IncrementMatchPlayerAssists :exec
UPDATE match_player_stats
SET 
    assists = assists + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: IncrementMatchPlayerDeaths :exec
UPDATE match_player_stats
SET 
    deaths = deaths + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: IncrementMatchPlayerFFKills :exec
UPDATE match_player_stats
SET 
    friendly_fire_kills = friendly_fire_kills + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: UpdateMatchPlayerScore :exec
UPDATE match_player_stats
SET 
    score = score + ?,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: IncrementMatchPlayerObjectiveCaptured :exec
UPDATE match_player_stats
SET 
    objectives_captured = objectives_captured + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: IncrementMatchPlayerObjectiveDestroyed :exec
UPDATE match_player_stats
SET 
    objectives_destroyed = objectives_destroyed + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND player_id = ?;

-- name: GetMatchPlayerStats :one
SELECT * FROM match_player_stats
WHERE match_id = ? AND player_id = ?;

-- name: GetAllPlayersInMatch :many
SELECT 
    mps.*,
    p.name as player_name,
    p.external_id as player_external_id
FROM match_player_stats mps
JOIN players p ON p.id = mps.player_id
WHERE mps.match_id = ?
ORDER BY mps.score DESC;

-- name: GetPlayerMatchHistory :many
SELECT 
    mps.*,
    m.map,
    m.mode,
    m.start_time,
    m.end_time,
    m.winner_team
FROM match_player_stats mps
JOIN matches m ON m.id = mps.match_id
WHERE mps.player_id = ?
ORDER BY m.start_time DESC
LIMIT ?;

-- name: GetPlayerStatsForPeriod :one
SELECT 
    COUNT(DISTINCT mps.match_id) as total_matches,
    SUM(mps.kills) as total_kills,
    SUM(mps.assists) as total_assists,
    SUM(mps.deaths) as total_deaths,
    SUM(mps.score) as total_score,
    SUM(mps.total_play_time) as total_play_time,
    SUM(mps.session_count) as total_sessions
FROM match_player_stats mps
JOIN matches m ON m.id = mps.match_id
WHERE mps.player_id = ? 
  AND m.server_id = ?
  AND m.start_time >= ?;

-- name: GetTopPlayersByKillsForPeriod :many
SELECT 
    p.id,
    p.name,
    p.external_id,
    SUM(mps.kills) as total_kills,
    SUM(mps.deaths) as total_deaths,
    SUM(mps.assists) as total_assists,
    COUNT(DISTINCT mps.match_id) as matches_played
FROM match_player_stats mps
JOIN players p ON p.id = mps.player_id
JOIN matches m ON m.id = mps.match_id
WHERE m.server_id = ?
  AND m.start_time >= ?
GROUP BY p.id, p.name, p.external_id
ORDER BY total_kills DESC
LIMIT ?;

-- name: GetPlayerBestMatch :one
SELECT 
    mps.*,
    m.map,
    m.mode,
    m.start_time
FROM match_player_stats mps
JOIN matches m ON m.id = mps.match_id
WHERE mps.player_id = ?
ORDER BY mps.kills DESC
LIMIT 1;

-- name: GetCurrentlyConnectedPlayers :many
SELECT 
    mps.*,
    p.name,
    p.external_id
FROM match_player_stats mps
JOIN players p ON p.id = mps.player_id
WHERE mps.match_id = ? 
  AND mps.is_currently_connected = 1;

-- name: DisconnectAllPlayersInMatch :exec
UPDATE match_player_stats
SET 
    is_currently_connected = 0,
    last_left_at = COALESCE(last_left_at, ?),
    updated_at = CURRENT_TIMESTAMP
WHERE match_id = ? AND is_currently_connected = 1;

-- name: GetStaleConnections :many
-- Find players still marked as connected in matches that ended
SELECT mps.*, m.end_time
FROM match_player_stats mps
JOIN matches m ON m.id = mps.match_id
WHERE mps.is_currently_connected = 1
  AND m.end_time IS NOT NULL;

