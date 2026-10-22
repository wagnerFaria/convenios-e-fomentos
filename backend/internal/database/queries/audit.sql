-- name: InsertAuditLog :exec
INSERT INTO audit_logs (actor_type, actor_id, action, entity, entity_id, payload, ip)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE (sqlc.narg('actor_type')::text IS NULL OR actor_type = sqlc.narg('actor_type'))
  AND (sqlc.narg('actor_id')::text IS NULL OR actor_id = sqlc.narg('actor_id'))
  AND (sqlc.narg('action')::text IS NULL OR action = sqlc.narg('action'))
  AND (sqlc.narg('entity')::text IS NULL OR entity = sqlc.narg('entity'))
  AND (sqlc.narg('entity_id')::text IS NULL OR entity_id = sqlc.narg('entity_id'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR created_at >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR created_at <= sqlc.narg('to_time'))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('page_limit') OFFSET sqlc.arg('page_offset');

-- name: CountAuditLogs :one
SELECT count(*) FROM audit_logs
WHERE (sqlc.narg('actor_type')::text IS NULL OR actor_type = sqlc.narg('actor_type'))
  AND (sqlc.narg('actor_id')::text IS NULL OR actor_id = sqlc.narg('actor_id'))
  AND (sqlc.narg('action')::text IS NULL OR action = sqlc.narg('action'))
  AND (sqlc.narg('entity')::text IS NULL OR entity = sqlc.narg('entity'))
  AND (sqlc.narg('entity_id')::text IS NULL OR entity_id = sqlc.narg('entity_id'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR created_at >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR created_at <= sqlc.narg('to_time'));
