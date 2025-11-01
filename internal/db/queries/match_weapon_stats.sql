-- name: UpsertMatchWeaponStats :exec
INSERT INTO match_weapon_stats (match_id, player_id, weapon_name, kills, assists)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(match_id, player_id, weapon_name) DO UPDATE SET
    kills = match_weapon_stats.kills + excluded.kills,
    assists = match_weapon_stats.assists + excluded.assists;

-- name: GetPlayerWeaponsForMatch :many
SELECT * FROM match_weapon_stats
WHERE match_id = ? AND player_id = ?
ORDER BY kills DESC;

-- name: GetTopWeaponsForMatch :many
SELECT 
    weapon_name,
    SUM(kills) as total_kills,
    SUM(assists) as total_assists,
    COUNT(DISTINCT player_id) as players_used
FROM match_weapon_stats
WHERE match_id = ?
GROUP BY weapon_name
ORDER BY total_kills DESC
LIMIT ?;

-- name: GetPlayerTopWeaponsForPeriod :many
SELECT 
    mws.weapon_name,
    SUM(mws.kills) as total_kills,
    SUM(mws.assists) as total_assists,
    COUNT(DISTINCT mws.match_id) as matches_used
FROM match_weapon_stats mws
JOIN matches m ON m.id = mws.match_id
WHERE mws.player_id = ?
  AND m.start_time >= ?
GROUP BY mws.weapon_name
ORDER BY total_kills DESC
LIMIT ?;
