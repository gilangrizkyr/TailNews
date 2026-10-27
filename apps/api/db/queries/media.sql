-- name: GetMediaByID :one
SELECT id, filename, file_path, mime_type, size_bytes, width, height, alt_text, caption, photographer, variants, created_at, updated_at
FROM media
WHERE id = ? LIMIT 1;

-- name: CreateMedia :execresult
INSERT INTO media (filename, file_path, mime_type, size_bytes, width, height, alt_text, caption, photographer, variants)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListMedia :many
SELECT id, filename, file_path, mime_type, size_bytes, width, height, alt_text, caption, photographer, variants, created_at, updated_at
FROM media
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: DeleteMedia :exec
DELETE FROM media WHERE id = ?;
