-- name: CountAdmins :one
SELECT count(*) FROM admins;

-- name: LockAdminSetup :exec
SELECT pg_advisory_xact_lock(7001);

-- name: GetAdminByID :one
SELECT * FROM admins WHERE id = $1;

-- name: GetAdminByEmail :one
SELECT * FROM admins WHERE email = $1;

-- name: CreateAdmin :one
INSERT INTO admins (first_name, last_name, email, password_hash, is_active, department_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: SetAdminPassword :exec
UPDATE admins SET password_hash = $2, updated_at = now() WHERE id = $1;

-- name: ActivateAdmin :exec
UPDATE admins SET is_active = true, updated_at = now() WHERE id = $1;

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
UPDATE admins SET first_name = $2, last_name = $3, email = $4, department_id = $5, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetAdminActive :exec
UPDATE admins SET is_active = $2, updated_at = now() WHERE id = $1;

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
