-- name: GetRedirectByFromPath :one
SELECT id, from_path, to_path, http_status, hit_count, created_at, updated_at
FROM redirects
WHERE from_path = ? LIMIT 1;

-- name: CreateRedirect :execresult
INSERT INTO redirects (from_path, to_path, http_status)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE to_path = VALUES(to_path), http_status = VALUES(http_status), updated_at = NOW();

-- name: IncrementRedirectHit :exec
UPDATE redirects
SET hit_count = hit_count + 1
WHERE id = ?;

-- name: ListRedirects :many
SELECT id, from_path, to_path, http_status, hit_count, created_at, updated_at
FROM redirects
ORDER BY id DESC
LIMIT ? OFFSET ?;
