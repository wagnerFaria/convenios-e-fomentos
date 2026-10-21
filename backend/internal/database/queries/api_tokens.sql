-- name: CreateAPIToken :one
INSERT INTO api_tokens (name, description, type, token_hash, hint, permissions, lifespan_days, expires_at, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAPIToken :one
SELECT * FROM api_tokens WHERE id = $1;

-- name: GetAPITokenForUpdate :one
SELECT * FROM api_tokens WHERE id = $1 FOR UPDATE;

-- Só tokens válidos: não revogados e não expirados.
-- name: GetActiveAPITokenByHash :one
SELECT * FROM api_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now());

-- name: ListAPITokens :many
SELECT * FROM api_tokens
WHERE sqlc.arg('include_revoked')::boolean OR revoked_at IS NULL
ORDER BY created_at DESC, id;

-- name: UpdateAPIToken :one
UPDATE api_tokens SET name = $2, description = $3, type = $4, permissions = $5, updated_at = now()
WHERE id = $1 AND revoked_at IS NULL
RETURNING *;

-- name: RegenerateAPIToken :one
UPDATE api_tokens SET token_hash = $2, hint = $3, expires_at = $4, last_used_at = NULL, updated_at = now()
WHERE id = $1 AND revoked_at IS NULL
RETURNING *;

-- name: RevokeAPIToken :execrows
UPDATE api_tokens SET revoked_at = now(), updated_at = now() WHERE id = $1 AND revoked_at IS NULL;

-- Grava no máximo uma vez por minuto, mesmo com várias instâncias.
-- name: TouchAPIToken :exec
UPDATE api_tokens SET last_used_at = now()
WHERE id = $1 AND (last_used_at IS NULL OR last_used_at < now() - interval '1 minute');
