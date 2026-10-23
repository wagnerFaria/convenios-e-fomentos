-- +goose Up

-- Dados de contato do proponente, preenchidos uma vez e usados como ponto de partida de cada projeto.
-- Nome e CPF vêm do MT Login (tabela admins). phones: [{"number":"65999998888","type":"mobile|landline"}];
-- emails: [{"address":"a@b.c","primary":true}].
CREATE TABLE proponent_profiles (
    admin_id          uuid PRIMARY KEY REFERENCES admins (id) ON DELETE CASCADE,
    phones            jsonb NOT NULL DEFAULT '[]',
    emails            jsonb NOT NULL DEFAULT '[]',
    uf                char(2),
    municipality_ibge text REFERENCES municipalities (ibge_code),
    job_title         text NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

-- O projeto guarda uma cópia dos dados do proponente e da instituição, para que a edição de um perfil ou de outro
-- projeto nunca mude o que foi informado aqui. Por enquanto só existe o rascunho; as colunas de recursos,
-- documentos e do ciclo de vida entram nas fases seguintes.
CREATE TABLE projects (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by_id uuid NOT NULL REFERENCES admins (id) ON DELETE RESTRICT,
    updated_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    -- Definida pelo analista na conferência; nula enquanto o projeto é do proponente.
    department_id uuid REFERENCES departments (id) ON DELETE SET NULL,
    status        text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft')),
    -- Controle de edição simultânea: cada gravação exige a versão lida e a incrementa.
    version       integer NOT NULL DEFAULT 1,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    -- Proponente.
    proponent_name              text NOT NULL,
    proponent_cpf               text NOT NULL CHECK (proponent_cpf ~ '^[0-9]{11}$'),
    proponent_phones            jsonb NOT NULL DEFAULT '[]',
    proponent_emails            jsonb NOT NULL DEFAULT '[]',
    proponent_uf                char(2),
    proponent_municipality_ibge text REFERENCES municipalities (ibge_code),
    proponent_job_title         text NOT NULL DEFAULT '',
    is_legal_representative     boolean,
    legal_rep_name              text NOT NULL DEFAULT '',
    legal_rep_cpf               text CHECK (legal_rep_cpf IS NULL OR legal_rep_cpf ~ '^[0-9]{11}$'),
    legal_rep_job_title         text NOT NULL DEFAULT '',

    -- Instituição.
    instrument                  text CHECK (instrument IN ('convenio', 'fomento')),
    institution_cnpj            text CHECK (institution_cnpj IS NULL OR institution_cnpj ~ '^[0-9]{14}$'),
    institution_name            text NOT NULL DEFAULT '',
    entity_type_id              uuid REFERENCES entity_types (id) ON DELETE RESTRICT,
    entity_type_other           text NOT NULL DEFAULT '',
    entity_nature               text CHECK (entity_nature IN ('public', 'private')),
    entity_non_profit           boolean NOT NULL DEFAULT false,
    institution_uf              char(2),
    institution_municipality_ibge text REFERENCES municipalities (ibge_code),
    institution_zip             text NOT NULL DEFAULT '',
    institution_street          text NOT NULL DEFAULT '',
    institution_number          text NOT NULL DEFAULT '',
    institution_complement      text NOT NULL DEFAULT '',
    institution_district        text NOT NULL DEFAULT '',
    institution_email           text NOT NULL DEFAULT '',
    institution_phone           text NOT NULL DEFAULT '',
    habilitation_status         text CHECK (habilitation_status IN ('full', 'partial')),
    habilitation_number         text NOT NULL DEFAULT '',
    habilitation_valid_until    date,

    -- Proposta.
    sigcon_number               text CHECK (sigcon_number IS NULL OR sigcon_number ~ '^[0-9]{1,6}/[0-9]{4}$'),
    sigcon_registered_on        date,
    title                       text NOT NULL DEFAULT '',
    object_description          text NOT NULL DEFAULT '',
    justification               text NOT NULL DEFAULT '',
    start_date                  date,
    end_date                    date,
    object_nature               text CHECK (object_nature IN ('works', 'goods', 'services', 'event', 'other')),
    has_artistic_hiring         boolean NOT NULL DEFAULT false,
    calamity                    boolean NOT NULL DEFAULT false,
    execution_place_detail      text NOT NULL DEFAULT '',
    CONSTRAINT projects_period CHECK (start_date IS NULL OR end_date IS NULL OR end_date >= start_date)
);
CREATE INDEX projects_created_by_idx ON projects (created_by_id, updated_at DESC);
CREATE INDEX projects_sigcon_idx ON projects (sigcon_number) WHERE sigcon_number IS NOT NULL;
CREATE INDEX projects_cnpj_idx ON projects (institution_cnpj) WHERE institution_cnpj IS NOT NULL;

-- Municípios onde o projeto será executado (todos de Mato Grosso; a regra fica na aplicação).
CREATE TABLE project_municipalities (
    project_id uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    ibge_code  text NOT NULL REFERENCES municipalities (ibge_code),
    PRIMARY KEY (project_id, ibge_code)
);

-- +goose Down
DROP TABLE project_municipalities;
DROP TABLE projects;
DROP TABLE proponent_profiles;
