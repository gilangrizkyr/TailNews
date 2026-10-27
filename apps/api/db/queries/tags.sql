-- name: GetTagByID :one
SELECT id, name, slug, created_at, updated_at
FROM tags
WHERE id = ? LIMIT 1;

-- name: GetTagBySlug :one
SELECT id, name, slug, created_at, updated_at
FROM tags
WHERE slug = ? LIMIT 1;

-- name: ListTags :many
SELECT id, name, slug, created_at, updated_at
FROM tags
ORDER BY name ASC;

-- name: CreateTag :execresult
INSERT INTO tags (name, slug)
VALUES (?, ?);

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;
