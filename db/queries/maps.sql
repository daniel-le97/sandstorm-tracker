-- Map and scenario management queries

-- name: UpsertMap :one
INSERT INTO maps (map_name, scenario)
VALUES (?, ?)
ON CONFLICT(map_name) DO UPDATE SET
    scenario = excluded.scenario,
    updated_at = CURRENT_TIMESTAMP
RETURNING id;

-- name: GetMapByName :one
SELECT * FROM maps 
WHERE map_name = ?;

-- name: GetAllMaps :many
SELECT * FROM maps 
ORDER BY map_name;



-- name: GetMapStats :many
SELECT map_name, scenario FROM maps ORDER BY map_name;