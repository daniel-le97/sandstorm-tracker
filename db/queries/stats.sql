-- Player statistics queries

-- name: GetPlayerStats :one
SELECT
    p.name as player_name,
    p.external_id as external_id,
    COALESCE(k.total_kills, 0) as total_kills,
    COALESCE(d.total_deaths, 0) as total_deaths,
    COALESCE(k.team_kills, 0) as team_kills,
    COALESCE(k.suicides, 0) as suicides,
    CASE WHEN COALESCE(d.total_deaths,0)=0 THEN COALESCE(k.total_kills,0) ELSE ROUND(CAST(COALESCE(k.total_kills,0) AS FLOAT)/CAST(COALESCE(d.total_deaths,0) AS FLOAT),2) END as kdr
FROM players p
LEFT JOIN (
  SELECT killer_id, SUM(CASE WHEN is_team_kill=0 AND is_suicide=0 THEN 1 ELSE 0 END) as total_kills,
         SUM(CASE WHEN is_team_kill=1 THEN 1 ELSE 0 END) as team_kills,
         SUM(CASE WHEN is_suicide=1 THEN 1 ELSE 0 END) as suicides
  FROM kills WHERE kills.server_id = ? GROUP BY killer_id
) k ON p.id = k.killer_id
LEFT JOIN (
  SELECT killer_id as victim_id, COUNT(*) as total_deaths FROM kills WHERE kills.server_id = ? GROUP BY killer_id
) d ON p.id = d.victim_id
WHERE p.external_id = ?;

-- name: GetTopPlayers :many
SELECT p.name as player_name, COALESCE(k.total_kills,0) as total_kills, COALESCE(d.total_deaths,0) as total_deaths,
  CASE WHEN COALESCE(d.total_deaths,0)=0 THEN COALESCE(k.total_kills,0) ELSE ROUND(CAST(COALESCE(k.total_kills,0) AS FLOAT)/CAST(COALESCE(d.total_deaths,0) AS FLOAT),2) END as kdr
FROM players p
LEFT JOIN (SELECT killer_id, SUM(CASE WHEN is_team_kill=0 AND is_suicide=0 THEN 1 ELSE 0 END) as total_kills FROM kills WHERE kills.server_id = ? GROUP BY killer_id) k ON p.id = k.killer_id
LEFT JOIN (SELECT killer_id as victim_id, COUNT(*) as total_deaths FROM kills WHERE kills.server_id = ? GROUP BY killer_id) d ON p.id = d.victim_id
WHERE COALESCE(k.total_kills,0) > 0
ORDER BY total_kills DESC, kdr DESC
LIMIT ?;



-- name: GetPlayerStatsGlobal :one
SELECT p.external_id as external_id, MIN(p.name) as player_name, COALESCE(SUM(k.total_kills),0) as total_kills, COALESCE(SUM(d.total_deaths),0) as total_deaths,
  CASE WHEN COALESCE(SUM(d.total_deaths),0)=0 THEN COALESCE(SUM(k.total_kills),0) ELSE ROUND(CAST(COALESCE(SUM(k.total_kills),0) AS FLOAT)/CAST(COALESCE(SUM(d.total_deaths),0) AS FLOAT),2) END as kdr
FROM players p
LEFT JOIN (SELECT killer_id, SUM(CASE WHEN is_team_kill=0 AND is_suicide=0 THEN 1 ELSE 0 END) as total_kills FROM kills GROUP BY killer_id) k ON p.id = k.killer_id
LEFT JOIN (SELECT killer_id as victim_id, COUNT(*) as total_deaths FROM kills GROUP BY killer_id) d ON p.id = d.victim_id
WHERE p.external_id = ?
GROUP BY p.external_id;