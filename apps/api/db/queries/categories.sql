-- name: GetCategoryByID :one
SELECT id, name, slug, description, parent_id, sort_order, is_nav, created_at, updated_at
FROM categories
WHERE id = ? LIMIT 1;

-- name: GetCategoryBySlug :one
SELECT id, name, slug, description, parent_id, sort_order, is_nav, created_at, updated_at
FROM categories
WHERE slug = ? LIMIT 1;

-- name: ListCategories :many
SELECT id, name, slug, description, parent_id, sort_order, is_nav, created_at, updated_at
FROM categories
ORDER BY sort_order ASC, name ASC;

-- name: ListNavCategories :many
SELECT id, name, slug, description, parent_id, sort_order, is_nav, created_at, updated_at
FROM categories
WHERE is_nav = TRUE
ORDER BY sort_order ASC, name ASC;

-- name: CreateCategory :execresult
INSERT INTO categories (name, slug, description, parent_id, sort_order, is_nav)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateCategory :exec
UPDATE categories
SET name = ?,
    slug = ?,
    description = ?,
    parent_id = ?,
    sort_order = ?,
    is_nav = ?,
    updated_at = NOW()
WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;
