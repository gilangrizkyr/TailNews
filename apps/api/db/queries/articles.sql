-- name: GetArticleBySlug :one
SELECT id, title, slug, excerpt, body, cover_media_id, category_id, author_id, created_by,
       status, is_headline, is_breaking, view_count, published_at, seo_title, seo_description,
       created_at, updated_at, deleted_at
FROM articles
WHERE slug = ? AND deleted_at IS NULL LIMIT 1;

-- name: GetArticleByID :one
SELECT id, title, slug, excerpt, body, cover_media_id, category_id, author_id, created_by,
       status, is_headline, is_breaking, view_count, published_at, seo_title, seo_description,
       created_at, updated_at, deleted_at
FROM articles
WHERE id = ? LIMIT 1;

-- name: CreateArticle :execresult
INSERT INTO articles (
    title, slug, excerpt, body, cover_media_id, category_id, author_id, created_by,
    status, is_headline, is_breaking, view_count, published_at, seo_title, seo_description
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateArticle :exec
UPDATE articles
SET title = ?,
    slug = ?,
    excerpt = ?,
    body = ?,
    cover_media_id = ?,
    category_id = ?,
    author_id = ?,
    status = ?,
    is_headline = ?,
    is_breaking = ?,
    published_at = ?,
    seo_title = ?,
    seo_description = ?,
    updated_at = NOW()
WHERE id = ?;

-- name: SoftDeleteArticle :exec
UPDATE articles
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = ?;

-- name: PermanentDeleteArticle :exec
DELETE FROM articles WHERE id = ?;

-- name: ListArticlesFirstPage :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListArticlesCursor :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE status = 'published' AND deleted_at IS NULL
  AND (published_at < ? OR (published_at = ? AND id < ?))
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListArticlesByCategoryFirstPage :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE category_id = ? AND status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListArticlesByCategoryCursor :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE category_id = ? AND status = 'published' AND deleted_at IS NULL
  AND (published_at < ? OR (published_at = ? AND id < ?))
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListArticlesByAuthorFirstPage :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE author_id = ? AND status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListArticlesByAuthorCursor :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE author_id = ? AND status = 'published' AND deleted_at IS NULL
  AND (published_at < ? OR (published_at = ? AND id < ?))
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListHeadlines :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE is_headline = TRUE AND status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: ListBreakingNews :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE is_breaking = TRUE AND status = 'published' AND deleted_at IS NULL
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: IncrementViewCountBatch :exec
UPDATE articles
SET view_count = view_count + ?
WHERE id = ?;

-- name: SearchArticlesFulltext :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id,
       is_headline, is_breaking, view_count, published_at
FROM articles
WHERE status = 'published' AND deleted_at IS NULL
  AND MATCH(title, excerpt) AGAINST(sqlc.arg('query') IN BOOLEAN MODE)
ORDER BY published_at DESC, id DESC
LIMIT ?;

-- name: GetScheduledArticlesToPublish :many
SELECT id, title, slug, category_id, author_id, published_at
FROM articles
WHERE status = 'scheduled' AND published_at <= ? AND deleted_at IS NULL
ORDER BY published_at ASC
LIMIT 100;

-- name: PublishScheduledArticle :exec
UPDATE articles
SET status = 'published',
    updated_at = NOW()
WHERE id = ?;

-- name: ListArticlesAdmin :many
SELECT id, title, slug, excerpt, cover_media_id, category_id, author_id, created_by,
       status, is_headline, is_breaking, view_count, published_at, created_at, updated_at
FROM articles
WHERE deleted_at IS NULL
ORDER BY id DESC
LIMIT ? OFFSET ?;
