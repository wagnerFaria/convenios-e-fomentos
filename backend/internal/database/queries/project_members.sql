-- name: InsertProjectMember :one
INSERT INTO project_members (project_id, cpf, label, level, invited_by_id, email)
VALUES ($1, $2, $3, $4, $5, sqlc.narg('email'))
RETURNING *;

-- name: GetProjectMemberByCPF :one
SELECT * FROM project_members WHERE project_id = $1 AND cpf = $2;

-- name: GetProjectMember :one
SELECT * FROM project_members WHERE id = $1 AND project_id = $2;

-- name: ListProjectMembers :many
SELECT * FROM project_members WHERE project_id = $1 AND status = 'active' ORDER BY created_at, id;

-- name: CountActiveProjectMembers :one
SELECT count(*) FROM project_members WHERE project_id = $1 AND status = 'active';

-- Convite de quem já teve acesso revogado: reativa a mesma linha.
-- name: ReactivateProjectMember :one
UPDATE project_members SET status = 'active', level = $2, label = $3, invited_by_id = $4, email = sqlc.narg('email'), created_at = now(), updated_at = now(),
    accepted_at = NULL, revoked_at = NULL, revoked_by_id = NULL
WHERE id = $1 AND status = 'revoked'
RETURNING *;

-- name: UpdateProjectMember :one
UPDATE project_members SET level = $2, label = $3, email = COALESCE(sqlc.narg('email'), email), updated_at = now() WHERE id = $1 AND status = 'active' RETURNING *;

-- name: RevokeProjectMember :one
UPDATE project_members SET status = 'revoked', revoked_at = now(), revoked_by_id = $2, updated_at = now()
WHERE id = $1 AND status = 'active'
RETURNING *;

-- Projetos em que o administrador é membro (o acesso vem do CPF da conta).
-- name: ListMembershipsOfAdmin :many
SELECT m.project_id, m.level FROM project_members m JOIN admins a ON a.cpf = m.cpf
WHERE a.id = $1 AND m.status = 'active';

-- name: GetMemberLevel :one
SELECT m.level FROM project_members m JOIN admins a ON a.cpf = m.cpf
WHERE a.id = $1 AND m.project_id = $2 AND m.status = 'active';

-- name: MarkMemberAccepted :exec
UPDATE project_members m SET accepted_at = now() FROM admins a
WHERE a.cpf = m.cpf AND a.id = $1 AND m.project_id = $2 AND m.status = 'active' AND m.accepted_at IS NULL;

-- E-mail da conta que tem o CPF, para avisar o convidado (conta bloqueada ou inativa não recebe).
-- name: GetNotifiableAdminEmail :one
SELECT email::text FROM admins WHERE cpf = $1 AND is_active AND NOT blocked;
