-- +goose Up
CREATE TABLE api_tokens (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text NOT NULL,
    description   text NOT NULL DEFAULT '',
    type          text NOT NULL CHECK (type IN ('read-only', 'full-access', 'custom')),
    -- HMAC-SHA256(API_TOKEN_SALT, token) em hexadecimal; o token em claro nunca é guardado
    token_hash    text NOT NULL UNIQUE,
    -- últimos caracteres do token, só para o administrador reconhecê-lo na lista
    hint          text NOT NULL,
    -- ações api:: liberadas (somente para type = custom)
    permissions   jsonb NOT NULL DEFAULT '[]'::jsonb,
    -- 7, 30 ou 90; nulo = ilimitado. Usado também ao regenerar.
    lifespan_days integer CHECK (lifespan_days IN (7, 30, 90)),
    expires_at    timestamptz,
    last_used_at  timestamptz,
    created_by    uuid REFERENCES admins (id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    revoked_at    timestamptz,
    CONSTRAINT api_tokens_permissions_array CHECK (jsonb_typeof(permissions) = 'array')
);
-- O nome é único entre os tokens não revogados.
CREATE UNIQUE INDEX api_tokens_name_active_idx ON api_tokens (lower(name)) WHERE revoked_at IS NULL;
CREATE INDEX api_tokens_created_by_idx ON api_tokens (created_by);

-- +goose Down
DROP TABLE api_tokens;
