-- name: ListDepartments :many
SELECT id, name FROM departments ORDER BY name;

-- name: UpsertDepartment :one
INSERT INTO departments (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;

-- name: ListDepartmentsWithCounts :many
SELECT d.id, d.name,
       (SELECT count(*) FROM admins a WHERE a.department_id = d.id) AS user_count,
       (SELECT count(*) FROM documents x WHERE x.department_id = d.id) AS document_count
FROM departments d
ORDER BY d.name;

-- name: GetDepartmentWithCounts :one
SELECT d.id, d.name,
       (SELECT count(*) FROM admins a WHERE a.department_id = d.id) AS user_count,
       (SELECT count(*) FROM documents x WHERE x.department_id = d.id) AS document_count
FROM departments d
WHERE d.id = $1;

-- Trava a linha: inserções/atualizações de administradores e documentos que referenciam o
-- departamento esperam a transação terminar.
-- name: GetDepartmentForUpdate :one
SELECT id, name FROM departments WHERE id = $1 FOR UPDATE;

-- name: CreateDepartment :one
INSERT INTO departments (name) VALUES ($1) RETURNING id, name;

-- name: UpdateDepartment :one
UPDATE departments SET name = $2 WHERE id = $1 RETURNING id, name;

-- name: DeleteDepartment :execrows
DELETE FROM departments WHERE id = $1;
