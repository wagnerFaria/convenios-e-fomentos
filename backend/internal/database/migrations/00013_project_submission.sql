-- +goose Up

-- Envio do projeto: o rascunho vira "submitted", ganha o código (AAAA-NNNNNN) e a etapa inicial do fluxo da equipe.
ALTER TABLE projects DROP CONSTRAINT projects_status_check;
ALTER TABLE projects
    ADD CONSTRAINT projects_status_check CHECK (status IN ('draft', 'submitted')),
    ADD COLUMN code            text UNIQUE,
    ADD COLUMN stage           text CHECK (stage IN ('awaiting_checklist')),
    ADD COLUMN submitted_at    timestamptz,
    ADD COLUMN submitted_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    ADD COLUMN submission_ip   text NOT NULL DEFAULT '',
    -- Enviado depois da data-limite de protocolo do ano eleitoral (Portaria Conjunta 71/2026): segue só com atos preparatórios.
    ADD COLUMN late_submission boolean NOT NULL DEFAULT false,
    ADD CONSTRAINT projects_submitted_consistency CHECK ((status = 'draft') = (code IS NULL AND submitted_at IS NULL AND stage IS NULL));

-- O número da proposta no SIGCon é único entre os projetos enviados (rascunhos repetidos só geram aviso).
CREATE UNIQUE INDEX projects_sigcon_submitted_key ON projects (sigcon_number) WHERE status <> 'draft' AND sigcon_number IS NOT NULL;

-- Contador do código por ano. A linha é travada pelo UPSERT, então envios simultâneos nunca repetem o número.
CREATE TABLE project_code_counters (
    year        integer PRIMARY KEY,
    last_number integer NOT NULL
);

-- Cada envio (e, nas fases seguintes, cada reenvio) guarda um retrato imutável do que foi enviado.
CREATE TABLE project_versions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    number          integer NOT NULL,
    reason          text NOT NULL CHECK (reason IN ('initial', 'complement')),
    submitted_at    timestamptz NOT NULL,
    submitted_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    submission_ip   text NOT NULL DEFAULT '',
    snapshot        jsonb NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (project_id, number)
);

-- +goose StatementBegin
CREATE FUNCTION project_versions_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'project_versions é imutável' USING ERRCODE = 'restrict_violation';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER project_versions_no_update BEFORE UPDATE ON project_versions
    FOR EACH ROW EXECUTE FUNCTION project_versions_immutable();

-- +goose Down
DROP TRIGGER project_versions_no_update ON project_versions;
DROP FUNCTION project_versions_immutable();
DROP TABLE project_versions;
DROP TABLE project_code_counters;
DROP INDEX projects_sigcon_submitted_key;
ALTER TABLE projects
    DROP CONSTRAINT projects_submitted_consistency, DROP COLUMN late_submission, DROP COLUMN submission_ip,
    DROP COLUMN submitted_by_id, DROP COLUMN submitted_at, DROP COLUMN stage, DROP COLUMN code,
    DROP CONSTRAINT projects_status_check;
ALTER TABLE projects ADD CONSTRAINT projects_status_check CHECK (status IN ('draft'));
