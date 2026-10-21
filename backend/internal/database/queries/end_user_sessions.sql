-- name: CreateEndUserSession :one
INSERT INTO end_user_sessions (end_user_id, family_id, family_started_at, refresh_token_hash, user_agent, ip, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEndUserSessionByTokenHash :one
SELECT * FROM end_user_sessions WHERE refresh_token_hash = $1;

-- name: ConsumeEndUserSession :one
UPDATE end_user_sessions SET revoked_at = now(), last_used_at = now()
WHERE refresh_token_hash = $1 AND revoked_at IS NULL
RETURNING *;

-- name: RevokeEndUserSessionFamily :exec
UPDATE end_user_sessions SET revoked_at = now() WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeAllEndUserSessions :exec
UPDATE end_user_sessions SET revoked_at = now() WHERE end_user_id = $1 AND revoked_at IS NULL;

-- name: RevokeOtherEndUserSessions :exec
UPDATE end_user_sessions SET revoked_at = now()
WHERE end_user_id = $1 AND family_id <> $2 AND revoked_at IS NULL;

-- Valida a sessão do access token; bloqueio e exigência de confirmação são aplicados na hora.
-- name: AuthenticateEndUserSession :one
SELECT u.id, u.username, u.email, u.role_id, r.type AS role_type, u.confirmed
FROM end_user_sessions s
JOIN end_users u ON u.id = s.end_user_id
JOIN end_user_roles r ON r.id = u.role_id
WHERE s.family_id = $1 AND s.end_user_id = $2
  AND s.revoked_at IS NULL AND s.expires_at > now()
  AND NOT u.blocked
LIMIT 1;
