-- name: GetAdminRole :one
SELECT * FROM admin_roles WHERE id = $1;

-- name: GetAdminRoleForUpdate :one
SELECT * FROM admin_roles WHERE id = $1 FOR UPDATE;

-- name: ListAdminRolesWithCount :many
SELECT r.*, (SELECT count(*) FROM admin_users_roles ur WHERE ur.role_id = r.id) AS user_count
FROM admin_roles r
ORDER BY r.name;

-- name: GetAdminRoleWithCount :one
SELECT r.*, (SELECT count(*) FROM admin_users_roles ur WHERE ur.role_id = r.id) AS user_count
FROM admin_roles r
WHERE r.id = $1;

-- name: AdminRoleCodeExists :one
SELECT EXISTS (SELECT 1 FROM admin_roles WHERE code = $1);

-- name: CreateAdminRole :one
INSERT INTO admin_roles (code, name, description, is_system) VALUES ($1, $2, $3, false)
RETURNING *;

-- name: UpdateAdminRole :one
UPDATE admin_roles SET name = $2, description = $3, updated_at = now() WHERE id = $1
RETURNING *;

-- name: DeleteAdminRole :execrows
DELETE FROM admin_roles WHERE id = $1;

-- name: CountAdminsWithRole :one
SELECT count(*) FROM admin_users_roles WHERE role_id = $1;

-- name: DeleteAdminPermissionsByRole :exec
DELETE FROM admin_permissions WHERE role_id = $1;
