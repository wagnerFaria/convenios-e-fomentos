-- +goose Up
CREATE TABLE end_user_roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    -- public, authenticated ou custom-<slug>
    type        text NOT NULL UNIQUE,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    is_system   boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE end_user_permissions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id    uuid NOT NULL REFERENCES end_user_roles (id) ON DELETE CASCADE,
    action     text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (role_id, action)
);
CREATE INDEX end_user_permissions_action_idx ON end_user_permissions (action);

-- O e-mail não é único no banco: "uma conta por e-mail" é uma configuração,
-- aplicada pela aplicação. password_hash é nulo para provedores externos (futuro).
CREATE TABLE end_users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username      citext NOT NULL UNIQUE,
    email         citext NOT NULL,
    password_hash text,
    confirmed     boolean NOT NULL DEFAULT false,
    blocked       boolean NOT NULL DEFAULT false,
    role_id       uuid NOT NULL REFERENCES end_user_roles (id) ON DELETE RESTRICT,
    provider      text NOT NULL DEFAULT 'local',
    -- campos extras aceitos no cadastro (register.allowedFields)
    profile       jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX end_users_email_idx ON end_users (email);
CREATE INDEX end_users_role_idx ON end_users (role_id);

CREATE TABLE end_user_sessions (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    end_user_id        uuid NOT NULL REFERENCES end_users (id) ON DELETE CASCADE,
    family_id          uuid NOT NULL,
    family_started_at  timestamptz NOT NULL DEFAULT now(),
    refresh_token_hash text NOT NULL UNIQUE,
    user_agent         text NOT NULL DEFAULT '',
    ip                 text NOT NULL DEFAULT '',
    created_at         timestamptz NOT NULL DEFAULT now(),
    last_used_at       timestamptz NOT NULL DEFAULT now(),
    expires_at         timestamptz NOT NULL,
    revoked_at         timestamptz
);
CREATE INDEX end_user_sessions_user_idx ON end_user_sessions (end_user_id);
CREATE INDEX end_user_sessions_family_idx ON end_user_sessions (family_id);

CREATE TABLE email_templates (
    key          text PRIMARY KEY,
    sender_name  text NOT NULL DEFAULT '',
    sender_email text NOT NULL,
    reply_to     text NOT NULL DEFAULT '',
    subject      text NOT NULL,
    body_html    text NOT NULL,
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE email_templates;
DROP TABLE end_user_sessions;
DROP TABLE end_users;
DROP TABLE end_user_permissions;
DROP TABLE end_user_roles;
