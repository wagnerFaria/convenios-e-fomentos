-- +goose Up

-- Ciclo de vida depois do envio: etapas da complementação, arquivamento (definitivo) e a origem de um projeto duplicado.
ALTER TABLE projects DROP CONSTRAINT projects_status_check;
ALTER TABLE projects DROP CONSTRAINT projects_stage_check;
ALTER TABLE projects
    ADD CONSTRAINT projects_status_check CHECK (status IN ('draft', 'submitted', 'archived')),
    ADD CONSTRAINT projects_stage_check CHECK (stage IN ('awaiting_checklist', 'in_analysis', 'awaiting_proponent')),
    ADD COLUMN archived_at       timestamptz,
    -- Nulo quando quem arquiva é o sistema (prazo de complementação vencido).
    ADD COLUMN archived_by_id    uuid REFERENCES admins (id) ON DELETE SET NULL,
    ADD COLUMN archive_reason    text CHECK (archive_reason IN ('complement_deadline', 'withdrawn', 'denied', 'missing_documents', 'other')),
    ADD COLUMN archive_note      text NOT NULL DEFAULT '',
    ADD COLUMN origin_project_id uuid REFERENCES projects (id) ON DELETE SET NULL,
    ADD CONSTRAINT projects_archived_consistency CHECK ((status = 'archived') = (archived_at IS NOT NULL AND archive_reason IS NOT NULL));

-- O número do SIGCon fica livre quando o projeto é arquivado: só os projetos ativos (enviados) o seguram.
DROP INDEX projects_sigcon_submitted_key;
CREATE UNIQUE INDEX projects_sigcon_submitted_key ON projects (sigcon_number) WHERE status = 'submitted' AND sigcon_number IS NOT NULL;

-- Pedido de complementação da equipe: o que o proponente precisa corrigir e até quando.
CREATE TABLE project_complement_requests (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    number           integer NOT NULL,
    requested_by_id  uuid REFERENCES admins (id) ON DELETE SET NULL,
    requested_at     timestamptz NOT NULL DEFAULT now(),
    due_at           timestamptz NOT NULL,
    message          text NOT NULL DEFAULT '',
    status           text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'answered', 'expired', 'cancelled')),
    answered_at      timestamptz,
    -- Número da versão do projeto criada com o reenvio.
    answered_version integer,
    UNIQUE (project_id, number)
);
CREATE UNIQUE INDEX project_complement_open_key ON project_complement_requests (project_id) WHERE status = 'open';
CREATE INDEX project_complement_due_idx ON project_complement_requests (due_at) WHERE status = 'open';

-- Cada item devolvido: uma seção (proponent, institution, proposal, funding, document-options) ou um documento do
-- checklist ("codigo" ou "codigo@id-da-emenda"). answered_at marca que o proponente mexeu nele depois do pedido.
CREATE TABLE project_complement_items (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id  uuid NOT NULL REFERENCES project_complement_requests (id) ON DELETE CASCADE,
    kind        text NOT NULL CHECK (kind IN ('section', 'document')),
    target      text NOT NULL,
    note        text NOT NULL DEFAULT '',
    answered_at timestamptz,
    UNIQUE (request_id, kind, target)
);

-- Arquivo "travado": já fez parte de uma versão enviada. Excluí-lo depois só o tira do checklist (superseded_at);
-- o objeto e o registro ficam, para a versão enviada continuar íntegra.
ALTER TABLE project_files
    ADD COLUMN locked        boolean NOT NULL DEFAULT false,
    ADD COLUMN superseded_at timestamptz;

-- +goose Down
ALTER TABLE project_files DROP COLUMN superseded_at, DROP COLUMN locked;
DROP TABLE project_complement_items;
DROP TABLE project_complement_requests;
DROP INDEX projects_sigcon_submitted_key;
CREATE UNIQUE INDEX projects_sigcon_submitted_key ON projects (sigcon_number) WHERE status <> 'draft' AND sigcon_number IS NOT NULL;
ALTER TABLE projects
    DROP CONSTRAINT projects_archived_consistency, DROP COLUMN origin_project_id, DROP COLUMN archive_note, DROP COLUMN archive_reason,
    DROP COLUMN archived_by_id, DROP COLUMN archived_at, DROP CONSTRAINT projects_stage_check, DROP CONSTRAINT projects_status_check;
ALTER TABLE projects
    ADD CONSTRAINT projects_status_check CHECK (status IN ('draft', 'submitted')),
    ADD CONSTRAINT projects_stage_check CHECK (stage IN ('awaiting_checklist'));
