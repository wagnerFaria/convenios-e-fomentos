-- name: GetAdminByID :one
SELECT * FROM admins WHERE id = $1;

-- name: GetAdminByCPF :one
SELECT * FROM admins WHERE cpf = $1;

-- Pré-cadastro por um administrador: falha com violação de unicidade se o CPF já existe.
-- name: CreateAdmin :one
INSERT INTO admins (cpf, first_name, last_name, email, is_active, department_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- Primeiro acesso: cria o registro; se outra requisição criou antes, não devolve linha.
-- name: ProvisionAdminFromLogin :one
INSERT INTO admins (cpf, first_name, last_name, email, is_active, last_login_at)
VALUES ($1, $2, $3, $4, true, now())
ON CONFLICT (cpf) DO NOTHING
RETURNING *;

-- Atualiza os dados vindos do MT Login e registra o acesso.
-- name: SyncAdminFromLogin :exec
UPDATE admins SET first_name = $2, last_name = $3, email = $4, last_login_at = now(), updated_at = now()
WHERE id = $1;

-- name: UpsertAdminRole :one
INSERT INTO admin_roles (code, name, description, is_system)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code) DO UPDATE SET code = EXCLUDED.code
RETURNING *;

-- name: GetAdminRolesByIDs :many
SELECT * FROM admin_roles WHERE id = ANY($1::uuid[]);

-- name: AssignAdminRole :exec
INSERT INTO admin_users_roles (admin_id, role_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListRolesForAdmin :many
SELECT r.* FROM admin_roles r
JOIN admin_users_roles ur ON ur.role_id = r.id
WHERE ur.admin_id = $1
ORDER BY r.name;

-- name: UpdateAdmin :one
UPDATE admins SET department_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetAdminActive :exec
UPDATE admins SET is_active = $2, updated_at = now() WHERE id = $1;

-- name: SetAdminBlocked :exec
UPDATE admins SET blocked = $2, updated_at = now() WHERE id = $1;

-- name: GetAdminsByIDs :many
SELECT * FROM admins WHERE id = ANY($1::uuid[]);

-- name: DeleteAdminsByIDs :execrows
DELETE FROM admins WHERE id = ANY($1::uuid[]);

-- name: RemoveAdminRole :execrows
DELETE FROM admin_users_roles WHERE admin_id = $1 AND role_id = $2;

-- name: CountAdminRoles :one
SELECT count(*) FROM admin_users_roles WHERE admin_id = $1;

-- Serializa operações que podem remover o último Super Admin ativo.
-- name: LockSuperAdmins :exec
SELECT pg_advisory_xact_lock(7002);

-- name: CountActiveSuperAdmins :one
SELECT count(DISTINCT a.id)
FROM admins a
JOIN admin_users_roles ur ON ur.admin_id = a.id
JOIN admin_roles r ON r.id = ur.role_id
WHERE r.code = $1 AND a.is_active AND NOT a.blocked;

-- Atribui a role pelo código (sem efeito se a role não existir ou já estiver atribuída).
-- name: AssignAdminRoleByCode :execrows
INSERT INTO admin_users_roles (admin_id, role_id)
SELECT $1, r.id FROM admin_roles r WHERE r.code = $2
ON CONFLICT DO NOTHING;

-- Dá a role a todo administrador ativo que não tem nenhuma (usado ao introduzir a role de proponente).
-- name: AssignRoleToAdminsWithoutRoles :execrows
INSERT INTO admin_users_roles (admin_id, role_id)
SELECT a.id, r.id FROM admins a, admin_roles r
WHERE r.code = $1
  AND NOT EXISTS (SELECT 1 FROM admin_users_roles ur WHERE ur.admin_id = a.id)
ON CONFLICT DO NOTHING;
