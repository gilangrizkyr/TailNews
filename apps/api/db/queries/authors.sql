-- name: GetAuthorByID :one
SELECT id, user_id, name, slug, bio, avatar_media_id, social_links, created_at, updated_at
FROM authors
WHERE id = ? LIMIT 1;

-- name: GetAuthorBySlug :one
SELECT id, user_id, name, slug, bio, avatar_media_id, social_links, created_at, updated_at
FROM authors
WHERE slug = ? LIMIT 1;

-- name: ListAuthors :many
SELECT id, user_id, name, slug, bio, avatar_media_id, social_links, created_at, updated_at
FROM authors
ORDER BY name ASC;

-- name: CreateAuthor :execresult
INSERT INTO authors (user_id, name, slug, bio, avatar_media_id, social_links)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateAuthor :exec
UPDATE authors
SET name = ?,
    slug = ?,
    bio = ?,
    avatar_media_id = ?,
    social_links = ?,
    updated_at = NOW()
WHERE id = ?;
