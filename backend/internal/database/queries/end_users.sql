-- name: LockEmail :exec
SELECT pg_advisory_xact_lock(hashtext($1::text));

-- name: CountEndUsersByEmail :one
SELECT count(*) FROM end_users WHERE email = $1;

-- name: CreateEndUser :one
INSERT INTO end_users (username, email, password_hash, confirmed, blocked, role_id, profile)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEndUserByID :one
SELECT * FROM end_users WHERE id = $1;

-- Aceita e-mail ou nome de usuário; sem "uma conta por e-mail" pode haver mais de uma conta.
-- name: GetEndUsersByIdentifier :many
SELECT * FROM end_users WHERE email = $1 OR username = $1 ORDER BY created_at LIMIT 5;

-- name: GetEndUsersByEmail :many
SELECT * FROM end_users WHERE email = $1 ORDER BY created_at LIMIT 5;

-- name: UpdateEndUser :one
UPDATE end_users SET username = $2, email = $3, updated_at = now() WHERE id = $1
RETURNING *;

-- name: SetEndUserPassword :exec
UPDATE end_users SET password_hash = $2, updated_at = now() WHERE id = $1;

-- name: SetEndUserConfirmed :exec
UPDATE end_users SET confirmed = $2, updated_at = now() WHERE id = $1;

-- name: SetEndUserBlocked :exec
UPDATE end_users SET blocked = $2, updated_at = now() WHERE id = $1;

-- name: SetEndUserRole :exec
UPDATE end_users SET role_id = $2, updated_at = now() WHERE id = $1;

-- name: DeleteEndUser :execrows
DELETE FROM end_users WHERE id = $1;
