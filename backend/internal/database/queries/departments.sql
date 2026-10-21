-- name: ListDepartments :many
SELECT id, name FROM departments ORDER BY name;

-- name: UpsertDepartment :one
INSERT INTO departments (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;
