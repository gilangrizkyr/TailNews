-- name: CreateRevision :execresult
INSERT INTO article_revisions (article_id, user_id, title, slug, excerpt, body)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListRevisionsByArticleID :many
SELECT id, article_id, user_id, title, slug, excerpt, body, created_at
FROM article_revisions
WHERE article_id = ?
ORDER BY created_at DESC;

-- name: GetRevisionByID :one
SELECT id, article_id, user_id, title, slug, excerpt, body, created_at
FROM article_revisions
WHERE id = ? LIMIT 1;
