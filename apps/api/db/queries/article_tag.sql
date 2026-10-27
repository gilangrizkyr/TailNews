-- name: AddArticleTag :exec
INSERT IGNORE INTO article_tag (article_id, tag_id)
VALUES (?, ?);

-- name: ClearArticleTags :exec
DELETE FROM article_tag WHERE article_id = ?;

-- name: GetTagsByArticleID :many
SELECT t.id, t.name, t.slug, t.created_at, t.updated_at
FROM tags t
JOIN article_tag at ON t.id = at.tag_id
WHERE at.article_id = ?
ORDER BY t.name ASC;

-- name: ListArticlesByTagCursor :many
SELECT a.id, a.title, a.slug, a.excerpt, a.cover_media_id, a.category_id, a.author_id,
       a.is_headline, a.is_breaking, a.view_count, a.published_at
FROM articles a
JOIN article_tag at ON a.id = at.article_id
WHERE at.tag_id = ? AND a.status = 'published' AND a.deleted_at IS NULL
  AND (a.published_at < ? OR (a.published_at = ? AND a.id < ?))
ORDER BY a.published_at DESC, a.id DESC
LIMIT ?;

-- name: ListArticlesByTagFirstPage :many
SELECT a.id, a.title, a.slug, a.excerpt, a.cover_media_id, a.category_id, a.author_id,
       a.is_headline, a.is_breaking, a.view_count, a.published_at
FROM articles a
JOIN article_tag at ON a.id = at.article_id
WHERE at.tag_id = ? AND a.status = 'published' AND a.deleted_at IS NULL
ORDER BY a.published_at DESC, a.id DESC
LIMIT ?;
