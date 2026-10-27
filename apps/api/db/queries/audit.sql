-- name: CreateAuditLog :execresult
INSERT INTO audit_logs (user_id, action, entity, entity_id, diff_summary, ip_address, user_agent)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListAuditLogs :many
SELECT id, user_id, action, entity, entity_id, diff_summary, ip_address, user_agent, created_at
FROM audit_logs
ORDER BY id DESC
LIMIT ? OFFSET ?;
