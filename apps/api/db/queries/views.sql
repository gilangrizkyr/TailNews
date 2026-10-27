-- name: UpsertDailyView :exec
INSERT INTO article_views_daily (article_id, view_date, views)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE views = views + VALUES(views);

-- name: GetTopArticlesByViewsPeriod :many
SELECT a.id, a.title, a.slug, a.excerpt, a.cover_media_id, a.category_id, a.author_id,
       a.is_headline, a.is_breaking, a.view_count, a.published_at, SUM(avd.views) as period_views
FROM article_views_daily avd
JOIN articles a ON avd.article_id = a.id
WHERE avd.view_date >= ? AND a.status = 'published' AND a.deleted_at IS NULL
GROUP BY a.id, a.title, a.slug, a.excerpt, a.cover_media_id, a.category_id, a.author_id,
         a.is_headline, a.is_breaking, a.view_count, a.published_at
ORDER BY period_views DESC
LIMIT ?;
