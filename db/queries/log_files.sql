-- Log file tracking queries

-- name: UpsertLogFile :one
INSERT INTO server_logs (server_id, log_path, open_time, lines_processed, file_size_bytes)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(server_id, log_path, open_time) DO UPDATE SET
    updated_at = CURRENT_TIMESTAMP
RETURNING id, lines_processed;

-- name: UpdateLogFileLines :exec
UPDATE server_logs
SET lines_processed = ?, file_size_bytes = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetLogFileByKey :one
SELECT * FROM server_logs 
WHERE server_id = ? AND log_path = ? AND open_time = ?;

-- name: GetLogFilesByServer :many
SELECT * FROM server_logs 
WHERE server_id = ?
ORDER BY open_time DESC;