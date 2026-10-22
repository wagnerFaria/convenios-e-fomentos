-- name: InsertDocumentFile :one
INSERT INTO document_files (id, document_id, object_key, filename, content_type, size_bytes, created_by_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetDocumentFile :one
SELECT * FROM document_files WHERE id = $1 AND document_id = $2;

-- name: ListDocumentFiles :many
SELECT * FROM document_files
WHERE document_id = $1 AND (NOT @only_confirmed::boolean OR status = 'confirmado')
ORDER BY created_at, id;

-- name: ConfirmDocumentFile :one
UPDATE document_files SET status = 'confirmado', size_bytes = $2, confirmed_at = now()
WHERE id = $1 AND status = 'pendente'
RETURNING *;

-- name: DeleteDocumentFile :execrows
DELETE FROM document_files WHERE id = $1 AND document_id = $2;

-- name: ListDocumentFileKeys :many
SELECT object_key FROM document_files WHERE document_id = $1;

-- name: ListStalePendingDocumentFiles :many
SELECT id, object_key FROM document_files
WHERE status = 'pendente' AND created_at < now() - make_interval(hours => @older_than_hours::int)
LIMIT 500;

-- name: DeleteDocumentFileByID :exec
DELETE FROM document_files WHERE id = $1;
