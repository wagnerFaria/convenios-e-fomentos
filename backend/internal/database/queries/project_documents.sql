-- name: ListDocumentRequirements :many
SELECT * FROM document_requirements WHERE active ORDER BY sort_order, code;

-- Carga inicial: não altera o que já existe (o catálogo pode ter sido ajustado pela administração).
-- name: InsertDefaultDocumentRequirement :execrows
INSERT INTO document_requirements (code, name, description, grp, sort_order, instruments, conditions, scope, accepts_justification, sources, templates, mandatory)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (code) DO NOTHING;

-- name: UpdateProjectDocumentOptions :one
UPDATE projects SET exclusive_supplier = $2, public_space_use = $3, reduced_deadline = $4,
    version = version + 1, updated_at = now(), updated_by_id = $5
WHERE id = $1
RETURNING *;

-- name: ListProjectDocumentItems :many
SELECT * FROM project_document_items WHERE project_id = $1;

-- name: GetProjectDocumentItem :one
SELECT * FROM project_document_items
WHERE project_id = $1 AND requirement_code = $2 AND amendment_id IS NOT DISTINCT FROM sqlc.narg('amendment_id')::uuid;

-- name: CreateProjectDocumentItem :one
INSERT INTO project_document_items (project_id, requirement_code, amendment_id)
VALUES ($1, $2, sqlc.narg('amendment_id')::uuid)
RETURNING *;

-- name: SetProjectDocumentJustification :one
UPDATE project_document_items SET justification = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- Só apaga o item se não restou nenhum arquivo, nem os travados (que guardam a versão enviada).
-- name: DeleteProjectDocumentItemIfEmpty :execrows
DELETE FROM project_document_items i
WHERE i.id = $1 AND i.justification = '' AND NOT EXISTS (SELECT 1 FROM project_files f WHERE f.item_id = i.id);

-- name: ListProjectFiles :many
SELECT f.*, i.requirement_code, i.amendment_id
FROM project_files f JOIN project_document_items i ON i.id = f.item_id
WHERE i.project_id = $1 AND f.superseded_at IS NULL
ORDER BY f.created_at, f.id;

-- name: GetProjectFile :one
SELECT f.*, i.project_id, i.requirement_code, i.amendment_id
FROM project_files f JOIN project_document_items i ON i.id = f.item_id
WHERE f.id = $1 AND i.project_id = $2;

-- name: CountProjectItemFiles :one
SELECT count(*) FROM project_files WHERE item_id = $1 AND superseded_at IS NULL;

-- name: InsertProjectFile :one
INSERT INTO project_files (id, item_id, object_key, filename, content_type, size_bytes, created_by_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ConfirmProjectFile :one
UPDATE project_files SET status = 'confirmed', size_bytes = $2, sha256 = $3, confirmed_at = now()
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- Totais do projeto para os limites de armazenamento: todos os arquivos (inclusive os substituídos, que continuam guardados).
-- name: ProjectFileTotals :one
SELECT count(*) AS total, count(*) FILTER (WHERE f.status = 'pending') AS pending, COALESCE(sum(f.size_bytes), 0)::bigint AS bytes
FROM project_files f JOIN project_document_items i ON i.id = f.item_id
WHERE i.project_id = $1;

-- name: DeleteProjectFile :execrows
DELETE FROM project_files WHERE id = $1;

-- name: ListProjectFileKeys :many
SELECT f.object_key FROM project_files f JOIN project_document_items i ON i.id = f.item_id WHERE i.project_id = $1;

-- Arquivos das exigências por emenda cuja emenda deixou de existir no projeto.
-- name: ListProjectFileKeysOfRemovedAmendments :many
SELECT f.object_key FROM project_files f JOIN project_document_items i ON i.id = f.item_id
WHERE i.project_id = $1 AND i.amendment_id IS NOT NULL AND NOT (i.amendment_id = ANY(sqlc.arg('keep')::uuid[]));

-- Arquivos já enviados (travados) de emendas que sairiam do projeto: a emenda não pode ser removida.
-- name: CountLockedFilesOfRemovedAmendments :one
SELECT count(*) FROM project_files f JOIN project_document_items i ON i.id = f.item_id
WHERE i.project_id = $1 AND i.amendment_id IS NOT NULL AND NOT (i.amendment_id = ANY(sqlc.arg('keep')::uuid[])) AND f.locked;

-- name: ListStalePendingProjectFiles :many
SELECT id, object_key FROM project_files
WHERE status = 'pending' AND created_at < now() - make_interval(hours => @older_than_hours::int)
LIMIT 500;

-- name: UpdateProjectAmendment :exec
UPDATE project_amendments SET position = $3, parliamentarian_id = $4, parliamentarian_free_name = $5,
    amendment_number = $6, year = $7, amount_cents = $8
WHERE id = $1 AND project_id = $2;

-- name: DeleteProjectAmendmentsExcept :exec
DELETE FROM project_amendments WHERE project_id = $1 AND NOT (id = ANY(sqlc.arg('keep')::uuid[]));
