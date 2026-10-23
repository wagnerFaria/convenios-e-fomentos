-- +goose Up

-- Recursos do projeto. Valores em centavos. total = recurso próprio da SEDEC + soma das emendas (calculado pela
-- aplicação a cada gravação). Os anexos exigidos (ofícios das emendas, declaração de contrapartida, edital do chamamento)
-- entram com os documentos, na fase seguinte.
ALTER TABLE projects
    ADD COLUMN funding_type text CHECK (funding_type IN ('amendment', 'own', 'mixed')),
    ADD COLUMN own_amount_cents bigint NOT NULL DEFAULT 0 CHECK (own_amount_cents >= 0),
    ADD COLUMN amendment_amount_cents bigint NOT NULL DEFAULT 0 CHECK (amendment_amount_cents >= 0),
    ADD COLUMN total_amount_cents bigint NOT NULL DEFAULT 0 CHECK (total_amount_cents >= 0),
    -- Contrapartida do proponente. Convênio: financeira ou em bens e serviços; Fomento: só em bens e serviços.
    ADD COLUMN counterpart_kind text CHECK (counterpart_kind IN ('financial', 'in_kind')),
    ADD COLUMN counterpart_cents bigint NOT NULL DEFAULT 0 CHECK (counterpart_cents >= 0),
    ADD COLUMN counterpart_description text NOT NULL DEFAULT '',
    -- Custos indiretos (Termo de Fomento).
    ADD COLUMN indirect_costs_cents bigint NOT NULL DEFAULT 0 CHECK (indirect_costs_cents >= 0),
    -- Chamamento público: realizado, dispensado ou inexigível (not_applicable = não se exige).
    ADD COLUMN tender_status text CHECK (tender_status IN ('held', 'waived', 'unfeasible', 'not_applicable')),
    ADD COLUMN tender_notice_number text NOT NULL DEFAULT '',
    ADD COLUMN tender_approved_on date,
    ADD COLUMN tender_legal_basis text NOT NULL DEFAULT '',
    ADD COLUMN tender_justification text NOT NULL DEFAULT '',
    -- Shows e eventos musicais (Lei 12.082/2023).
    ADD COLUMN artist_fees_cents bigint NOT NULL DEFAULT 0 CHECK (artist_fees_cents >= 0),
    ADD COLUMN regional_artist_fees_cents bigint NOT NULL DEFAULT 0 CHECK (regional_artist_fees_cents >= 0),
    ADD COLUMN governor_waiver boolean NOT NULL DEFAULT false,
    -- Data de coleta da referência de preço mais antiga (IN 002/2026: no máximo 60 dias).
    ADD COLUMN price_oldest_collected_on date;

-- Emendas parlamentares que compõem o projeto. O parlamentar vem do cadastro; se não estiver nele, o proponente
-- informa o nome por extenso (parliamentarian_free_name) e a administração organiza depois.
CREATE TABLE project_amendments (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id              uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    position                integer NOT NULL,
    parliamentarian_id      uuid REFERENCES parliamentarians (id) ON DELETE RESTRICT,
    parliamentarian_free_name text NOT NULL DEFAULT '',
    amendment_number        text NOT NULL DEFAULT '',
    year                    integer,
    amount_cents            bigint NOT NULL DEFAULT 0 CHECK (amount_cents >= 0),
    CONSTRAINT project_amendments_who CHECK (parliamentarian_id IS NULL OR parliamentarian_free_name = '')
);
CREATE INDEX project_amendments_project_idx ON project_amendments (project_id, position);
CREATE INDEX project_amendments_parliamentarian_idx ON project_amendments (parliamentarian_id);

-- +goose Down
DROP TABLE project_amendments;
ALTER TABLE projects
    DROP COLUMN price_oldest_collected_on, DROP COLUMN governor_waiver, DROP COLUMN regional_artist_fees_cents,
    DROP COLUMN artist_fees_cents, DROP COLUMN tender_justification, DROP COLUMN tender_legal_basis,
    DROP COLUMN tender_approved_on, DROP COLUMN tender_notice_number, DROP COLUMN tender_status,
    DROP COLUMN indirect_costs_cents, DROP COLUMN counterpart_description, DROP COLUMN counterpart_cents,
    DROP COLUMN counterpart_kind, DROP COLUMN total_amount_cents, DROP COLUMN amendment_amount_cents,
    DROP COLUMN own_amount_cents, DROP COLUMN funding_type;
