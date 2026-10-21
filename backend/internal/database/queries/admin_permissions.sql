-- name: ListAdminPermissionsByRole :many
SELECT * FROM admin_permissions WHERE role_id = $1 ORDER BY action, subject;

-- name: CreateAdminPermission :one
INSERT INTO admin_permissions (role_id, action, subject, fields, conditions, properties)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- Remove permissões de ações que não existem mais no código.
-- name: DeleteAdminPermissionsNotInActions :many
DELETE FROM admin_permissions WHERE NOT (action = ANY($1::text[]))
RETURNING role_id, action;

-- name: ListAdminIDsSharingRoles :many
SELECT DISTINCT admin_id FROM admin_users_roles WHERE role_id = ANY($1::uuid[]);
