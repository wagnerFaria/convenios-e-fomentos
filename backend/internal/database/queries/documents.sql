-- name: InsertDocument :one
INSERT INTO documents (titulo, conteudo, observacao_interna, department_id, created_by_id, updated_by_id, owner_user_id)
VALUES ($1, $2, $3, $4, $5, $5, $6)
RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents WHERE id = $1;

-- name: GetDocumentForUpdate :one
SELECT * FROM documents WHERE id = $1 FOR UPDATE;

-- name: UpdateDocument :one
UPDATE documents SET titulo = $2, conteudo = $3, status = $4, observacao_interna = $5, department_id = $6,
  published_at = $7, updated_by_id = $8, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDocument :execrows
DELETE FROM documents WHERE id = $1;

-- name: ListDepartmentsByName :many
SELECT id, name FROM departments WHERE name = ANY($1::text[]);
