-- name: GetEndUserRole :one
SELECT * FROM end_user_roles WHERE id = $1;

-- name: GetEndUserRoleForUpdate :one
SELECT * FROM end_user_roles WHERE id = $1 FOR UPDATE;

-- name: GetEndUserRoleByType :one
SELECT * FROM end_user_roles WHERE type = $1;

-- name: UpsertEndUserRole :one
INSERT INTO end_user_roles (type, name, description, is_system) VALUES ($1, $2, $3, $4)
ON CONFLICT (type) DO UPDATE SET type = EXCLUDED.type
RETURNING *;

-- name: ListEndUserRolesWithCount :many
SELECT r.*, (SELECT count(*) FROM end_users u WHERE u.role_id = r.id) AS user_count
FROM end_user_roles r
ORDER BY r.is_system DESC, r.name;

-- name: GetEndUserRoleWithCount :one
SELECT r.*, (SELECT count(*) FROM end_users u WHERE u.role_id = r.id) AS user_count
FROM end_user_roles r WHERE r.id = $1;

-- name: EndUserRoleTypeExists :one
SELECT EXISTS (SELECT 1 FROM end_user_roles WHERE type = $1);

-- name: CreateEndUserRole :one
INSERT INTO end_user_roles (type, name, description, is_system) VALUES ($1, $2, $3, false)
RETURNING *;

-- name: UpdateEndUserRole :one
UPDATE end_user_roles SET name = $2, description = $3, updated_at = now() WHERE id = $1
RETURNING *;

-- name: DeleteEndUserRole :execrows
DELETE FROM end_user_roles WHERE id = $1;

-- name: ListEndUserPermissionActions :many
SELECT action FROM end_user_permissions WHERE role_id = $1 ORDER BY action;

-- name: DeleteEndUserPermissionsByRole :exec
DELETE FROM end_user_permissions WHERE role_id = $1;

-- name: InsertEndUserPermission :exec
INSERT INTO end_user_permissions (role_id, action) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: MoveEndUsersToRole :execrows
UPDATE end_users SET role_id = $2, updated_at = now() WHERE role_id = $1;

-- name: DeleteEndUserPermissionsNotInActions :many
DELETE FROM end_user_permissions WHERE NOT (action = ANY($1::text[]))
RETURNING role_id, action;
