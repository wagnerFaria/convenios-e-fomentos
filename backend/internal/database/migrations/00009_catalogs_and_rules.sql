-- +goose Up

-- Tipos de entidade que podem ser proponentes. A natureza (pública ou privada) decide o checklist e as
-- regras de instrumento; "Outro" (is_system) é sempre mantida e pede o texto livre no cadastro do projeto.
CREATE TABLE entity_types (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code        text NOT NULL UNIQUE,
    name        text NOT NULL,
    nature      text NOT NULL CHECK (nature IN ('public', 'private')),
    non_profit  boolean NOT NULL DEFAULT false,
    active      boolean NOT NULL DEFAULT true,
    is_system   boolean NOT NULL DEFAULT false,
    sort_order  integer NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT entity_types_name_not_blank CHECK (btrim(name) <> '')
);
CREATE UNIQUE INDEX entity_types_name_key ON entity_types (lower(name));

-- Municípios do IBGE (referência somente leitura, carregada pela aplicação). search_name é o nome sem
-- acentos e em minúsculas, usado na busca.
CREATE TABLE municipalities (
    ibge_code   text PRIMARY KEY CHECK (ibge_code ~ '^[0-9]{7}$'),
    name        text NOT NULL,
    uf          char(2) NOT NULL,
    search_name text NOT NULL
);
CREATE INDEX municipalities_uf_search_idx ON municipalities (uf, search_name text_pattern_ops);
CREATE INDEX municipalities_search_idx ON municipalities (search_name text_pattern_ops);

-- Parlamentares que destinam emendas. Mantidos pelos administradores.
CREATE TABLE parliamentarians (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    chamber    text NOT NULL CHECK (chamber IN ('state', 'federal')),
    active     boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT parliamentarians_name_not_blank CHECK (btrim(name) <> '')
);
CREATE UNIQUE INDEX parliamentarians_name_chamber_key ON parliamentarians (lower(name), chamber);

-- Parâmetros das regras de negócio (limites, prazos, datas), versionados por vigência. A chave é validada
-- pela aplicação (internal/rules). Duas versões da mesma chave não podem se sobrepor no tempo.
CREATE TABLE rule_parameters (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key         text NOT NULL,
    value       jsonb NOT NULL,
    valid_from  date NOT NULL,
    valid_to    date,
    note        text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT rule_parameters_period CHECK (valid_to IS NULL OR valid_to >= valid_from),
    CONSTRAINT rule_parameters_key_from_key UNIQUE (key, valid_from)
);
CREATE INDEX rule_parameters_key_idx ON rule_parameters (key, valid_from DESC);

-- +goose Down
DROP TABLE rule_parameters;
DROP TABLE parliamentarians;
DROP TABLE municipalities;
DROP TABLE entity_types;
