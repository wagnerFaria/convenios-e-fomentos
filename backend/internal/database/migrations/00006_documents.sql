-- +goose Up
CREATE TABLE documents (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo             text NOT NULL,
    conteudo           text NOT NULL DEFAULT '',
    status             text NOT NULL DEFAULT 'rascunho' CHECK (status IN ('rascunho', 'publicado')),
    -- nunca exposta na API pública
    observacao_interna text NOT NULL DEFAULT '',
    department_id      uuid REFERENCES departments (id) ON DELETE SET NULL,
    published_at       timestamptz,
    -- criador administrador (usado pelas condições is-creator e has-same-role-as-creator)
    created_by_id      uuid REFERENCES admins (id) ON DELETE SET NULL,
    updated_by_id      uuid REFERENCES admins (id) ON DELETE SET NULL,
    -- dono usuário final (usado pela policy is-owner na API pública)
    owner_user_id      uuid REFERENCES end_users (id) ON DELETE SET NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT documents_published_at_matches_status CHECK ((status = 'publicado') = (published_at IS NOT NULL))
);
CREATE INDEX documents_status_published_idx ON documents (status, published_at DESC);
CREATE INDEX documents_department_idx ON documents (department_id);
CREATE INDEX documents_created_by_idx ON documents (created_by_id);
CREATE INDEX documents_owner_idx ON documents (owner_user_id);
CREATE INDEX documents_created_at_idx ON documents (created_at DESC);

-- +goose Down
DROP TABLE documents;
