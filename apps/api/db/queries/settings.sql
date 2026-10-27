-- name: GetSetting :one
SELECT key_name, value, updated_at
FROM settings
WHERE key_name = ? LIMIT 1;

-- name: UpsertSetting :exec
INSERT INTO settings (key_name, value)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = NOW();

-- name: ListSettings :many
SELECT key_name, value, updated_at
FROM settings
ORDER BY key_name ASC;
