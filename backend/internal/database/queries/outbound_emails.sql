-- Devolve 0 linhas quando a chave de deduplicação já existe (o mesmo e-mail não entra duas vezes).
-- name: EnqueueEmail :execrows
INSERT INTO outbound_emails (kind, project_id, to_email, subject, body, dedup_key)
VALUES (sqlc.arg(kind), sqlc.arg(project_id), sqlc.arg(to_email), sqlc.arg(subject), sqlc.arg(body), sqlc.arg(dedup_key))
ON CONFLICT (dedup_key) DO NOTHING;

-- Reserva um lote de mensagens vencidas e sem concessão, contando a tentativa. SKIP LOCKED deixa várias instâncias da API
-- trabalharem ao mesmo tempo sem pegar a mesma mensagem.
-- name: ClaimDueEmails :many
UPDATE outbound_emails
SET attempts = attempts + 1, locked_until = now() + make_interval(secs => sqlc.arg(lease_seconds)::int)
WHERE id IN (
    SELECT o.id FROM outbound_emails o
    WHERE o.status = 'pending' AND o.next_attempt_at <= now() AND (o.locked_until IS NULL OR o.locked_until < now())
    ORDER BY o.next_attempt_at, o.created_at
    LIMIT sqlc.arg(batch)::int
    FOR UPDATE SKIP LOCKED
)
RETURNING id, kind, to_email, subject, body, attempts;

-- name: MarkEmailSent :exec
UPDATE outbound_emails SET status = 'sent', sent_at = now(), locked_until = NULL, last_error = '' WHERE id = $1;

-- name: RescheduleEmail :exec
UPDATE outbound_emails SET locked_until = NULL, next_attempt_at = sqlc.arg(next_attempt_at), last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: FailEmail :exec
UPDATE outbound_emails SET status = 'failed', locked_until = NULL, last_error = $2 WHERE id = $1;

-- Apaga as mensagens já resolvidas (enviadas ou desistidas) criadas antes do limite: o corpo tem dados pessoais.
-- name: PurgeResolvedEmails :execrows
DELETE FROM outbound_emails WHERE status <> 'pending' AND created_at < $1;

-- E-mails de administradores ativos de um perfil (quem está inativo ou bloqueado não recebe).
-- name: ListStaffEmailsByRole :many
SELECT a.email::text AS email
FROM admins a
JOIN admin_users_roles ur ON ur.admin_id = a.id
JOIN admin_roles r ON r.id = ur.role_id
WHERE r.code = $1 AND a.is_active AND NOT a.blocked
ORDER BY a.email;

-- Idem, só de um departamento.
-- name: ListStaffEmailsByRoleAndDepartment :many
SELECT a.email::text AS email
FROM admins a
JOIN admin_users_roles ur ON ur.admin_id = a.id
JOIN admin_roles r ON r.id = ur.role_id
WHERE r.code = $1 AND a.department_id = $2 AND a.is_active AND NOT a.blocked
ORDER BY a.email;
