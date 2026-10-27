-- name: GetPageBySlug :one
SELECT id, title, slug, content, is_published, created_at, updated_at
FROM pages
WHERE slug = ? LIMIT 1;

-- name: ListPages :many
SELECT id, title, slug, is_published, created_at, updated_at
FROM pages
ORDER BY title ASC;

-- name: CreatePage :execresult
INSERT INTO pages (title, slug, content, is_published)
VALUES (?, ?, ?, ?);

-- name: UpdatePage :exec
UPDATE pages
SET title = ?,
    slug = ?,
    content = ?,
    is_published = ?,
    updated_at = NOW()
WHERE id = ?;

-- name: DeletePage :exec
DELETE FROM pages WHERE id = ?;
