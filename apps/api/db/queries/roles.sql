-- name: GetRoleByID :one
SELECT id, name, permissions, created_at, updated_at
FROM roles
WHERE id = ? LIMIT 1;

-- name: ListRoles :many
SELECT id, name, permissions, created_at, updated_at
FROM roles
ORDER BY id ASC;
