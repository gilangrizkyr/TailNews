-- name: GetUserByEmail :one
SELECT id, role_id, name, email, password_hash, is_active, failed_login_attempts, locked_until, created_at, updated_at
FROM users
WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT id, role_id, name, email, password_hash, is_active, failed_login_attempts, locked_until, created_at, updated_at
FROM users
WHERE id = ? LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users (role_id, name, email, password_hash, is_active)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateUserFailedLogin :exec
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE WHEN failed_login_attempts + 1 >= 5 THEN DATE_ADD(NOW(), INTERVAL 15 MINUTE) ELSE locked_until END,
    updated_at = NOW()
WHERE id = ?;

-- name: ResetUserFailedLogin :exec
UPDATE users
SET failed_login_attempts = 0,
    locked_until = NULL,
    updated_at = NOW()
WHERE id = ?;

-- name: ListUsers :many
SELECT id, role_id, name, email, is_active, created_at, updated_at
FROM users
ORDER BY id DESC;
