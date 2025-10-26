-- name: AddPlayerToMatch :exec
INSERT INTO match_players (match_id, player_id)
VALUES (?, ?);

-- name: RemovePlayerFromMatch :exec
DELETE FROM match_players WHERE match_id = ? AND player_id = ?;

-- name: GetPlayersForMatch :many
SELECT p.* FROM players p
JOIN match_players mp ON p.id = mp.player_id
WHERE mp.match_id = ?;

-- name: GetMatchesForPlayer :many
SELECT m.* FROM matches m
JOIN match_players mp ON m.id = mp.match_id
WHERE mp.player_id = ?;
