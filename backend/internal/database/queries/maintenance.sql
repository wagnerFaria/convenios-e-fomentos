-- Sessões (admin e usuário final) expiradas há mais de N dias. Revogadas ou rotacionadas só
-- saem depois de expirar + N dias, preservando a detecção de reutilização de refresh token.
-- name: DeleteExpiredAdminSessions :execrows
DELETE FROM admin_sessions WHERE expires_at < now() - make_interval(days => @grace_days::int);

-- name: DeleteExpiredEndUserSessions :execrows
DELETE FROM end_user_sessions WHERE expires_at < now() - make_interval(days => @grace_days::int);

-- Tokens de uso único já usados ou expirados há mais de N dias.
-- name: DeleteStaleOneTimeTokens :execrows
DELETE FROM one_time_tokens
WHERE coalesce(used_at, expires_at) < now() - make_interval(days => @grace_days::int);
