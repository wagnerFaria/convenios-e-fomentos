-- name: CreateAdminSession :one
INSERT INTO admin_sessions (admin_id, family_id, family_started_at, refresh_token_hash, user_agent, ip, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAdminSessionByTokenHash :one
SELECT * FROM admin_sessions WHERE refresh_token_hash = $1;

-- Consome o refresh token de forma atômica: só uma requisição concorrente vence.
-- name: ConsumeAdminSession :one
UPDATE admin_sessions SET revoked_at = now(), last_used_at = now()
WHERE refresh_token_hash = $1 AND revoked_at IS NULL
RETURNING *;

-- name: RevokeAdminSessionFamily :exec
UPDATE admin_sessions SET revoked_at = now()
WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeAdminSessionFamilyOfAdmin :execrows
UPDATE admin_sessions SET revoked_at = now()
WHERE family_id = $1 AND admin_id = $2 AND revoked_at IS NULL;

-- name: RevokeAllAdminSessions :exec
UPDATE admin_sessions SET revoked_at = now()
WHERE admin_id = $1 AND revoked_at IS NULL;

-- name: RevokeOtherAdminSessions :exec
UPDATE admin_sessions SET revoked_at = now()
WHERE admin_id = $1 AND family_id <> $2 AND revoked_at IS NULL;

-- name: ListActiveAdminSessions :many
SELECT * FROM admin_sessions
WHERE admin_id = $1 AND revoked_at IS NULL AND expires_at > now()
ORDER BY family_started_at DESC;

-- Valida a sessão referenciada por um access token e devolve o administrador.
-- name: AuthenticateAdminSession :one
SELECT a.id, a.email, a.first_name, a.last_name, a.department_id
FROM admin_sessions s
JOIN admins a ON a.id = s.admin_id
WHERE s.family_id = $1 AND s.admin_id = $2
  AND s.revoked_at IS NULL AND s.expires_at > now()
  AND a.is_active AND NOT a.blocked
LIMIT 1;
