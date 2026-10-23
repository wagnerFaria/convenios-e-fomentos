-- +goose Up

-- Catálogo das exigências documentais. Cada exigência vale para um ou dois instrumentos e só se aplica quando as condições
-- (todas, E lógico) casam com as respostas do projeto. Fica no banco para que as normas possam mudar sem alterar código; a
-- carga inicial (internal/project/requirements.go) só acrescenta o que ainda não existe.
-- conditions: [{"field":"objectNature","in":["works"]}]; sources: {"convenio":"C2","fomento":"F2"} (item do checklist da SEDEC);
-- templates: {"convenio":{"name":"Anexo II","url":"..."}}.
CREATE TABLE document_requirements (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                  text NOT NULL UNIQUE,
    name                  text NOT NULL,
    description           text NOT NULL DEFAULT '',
    grp                   text NOT NULL,
    sort_order            integer NOT NULL,
    instruments           text[] NOT NULL CHECK (cardinality(instruments) > 0 AND instruments <@ ARRAY['convenio', 'fomento']),
    conditions            jsonb NOT NULL DEFAULT '[]',
    -- project: um por projeto; amendment: um por emenda parlamentar do projeto.
    scope                 text NOT NULL DEFAULT 'project' CHECK (scope IN ('project', 'amendment')),
    accepts_justification boolean NOT NULL DEFAULT true,
    sources               jsonb NOT NULL DEFAULT '{}',
    templates             jsonb NOT NULL DEFAULT '{}',
    active                boolean NOT NULL DEFAULT true,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- Situações informadas pelo proponente que acrescentam documentos ao checklist.
ALTER TABLE projects
    ADD COLUMN exclusive_supplier boolean NOT NULL DEFAULT false,
    ADD COLUMN public_space_use   boolean NOT NULL DEFAULT false,
    ADD COLUMN reduced_deadline   boolean NOT NULL DEFAULT false;

-- Um item é o espaço de uma exigência no projeto (ou, nas exigências por emenda, naquela emenda). Só existe depois que
-- o proponente anexa algo ou escreve a justificativa de não envio; o checklist é calculado pelas respostas do projeto.
CREATE TABLE project_document_items (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    requirement_code text NOT NULL REFERENCES document_requirements (code) ON UPDATE CASCADE ON DELETE RESTRICT,
    amendment_id     uuid REFERENCES project_amendments (id) ON DELETE CASCADE,
    justification    text NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX project_document_items_project_key ON project_document_items (project_id, requirement_code) WHERE amendment_id IS NULL;
CREATE UNIQUE INDEX project_document_items_amendment_key ON project_document_items (project_id, requirement_code, amendment_id) WHERE amendment_id IS NOT NULL;

CREATE TABLE project_files (
    id            uuid PRIMARY KEY,
    item_id       uuid NOT NULL REFERENCES project_document_items (id) ON DELETE CASCADE,
    object_key    text NOT NULL UNIQUE,
    filename      text NOT NULL,
    content_type  text NOT NULL,
    size_bytes    bigint NOT NULL CHECK (size_bytes > 0),
    status        text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed')),
    created_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    confirmed_at  timestamptz
);
CREATE INDEX project_files_item_idx ON project_files (item_id, created_at);
CREATE INDEX project_files_pending_idx ON project_files (created_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE project_files;
DROP TABLE project_document_items;
ALTER TABLE projects DROP COLUMN reduced_deadline, DROP COLUMN public_space_use, DROP COLUMN exclusive_supplier;
DROP TABLE document_requirements;
