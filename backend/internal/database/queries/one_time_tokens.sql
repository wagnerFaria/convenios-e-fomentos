-- name: CreateOneTimeToken :exec
INSERT INTO one_time_tokens (subject_type, subject_id, purpose, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- Consome o token de forma atômica (uso único).
-- name: ConsumeOneTimeToken :one
UPDATE one_time_tokens SET used_at = now()
WHERE token_hash = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > now()
RETURNING *;

-- name: InvalidateOneTimeTokens :exec
UPDATE one_time_tokens SET used_at = now()
WHERE subject_type = $1 AND subject_id = $2 AND purpose = $3 AND used_at IS NULL;
