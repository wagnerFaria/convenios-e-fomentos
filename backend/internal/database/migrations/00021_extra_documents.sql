-- +goose Up

-- Documentos adicionais: o proponente anexa documentos que nenhuma exigência do catálogo cobre. São itens sem exigência
-- (requirement_code nulo), com título e descrição informados por ele. Reaproveitam project_files, então o envio, o
-- travamento, a duplicação e a limpeza de arquivos funcionam como nos demais documentos.
ALTER TABLE project_document_items
    ALTER COLUMN requirement_code DROP NOT NULL,
    ADD COLUMN title       text NOT NULL DEFAULT '',
    ADD COLUMN description text NOT NULL DEFAULT '',
    ADD CONSTRAINT project_document_items_kind_check CHECK (
        (requirement_code IS NOT NULL AND title = '' AND description = '')
        OR (requirement_code IS NULL AND title <> '' AND amendment_id IS NULL)
    );

-- +goose Down
DELETE FROM project_document_items WHERE requirement_code IS NULL;
ALTER TABLE project_document_items
    DROP CONSTRAINT project_document_items_kind_check,
    DROP COLUMN description,
    DROP COLUMN title,
    ALTER COLUMN requirement_code SET NOT NULL;
