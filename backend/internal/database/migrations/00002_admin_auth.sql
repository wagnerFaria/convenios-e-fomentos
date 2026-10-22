-- +goose Up
CREATE TABLE admin_roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code        text NOT NULL UNIQUE,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    is_system   boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- password_hash é nulo enquanto o convite não for aceito.
CREATE TABLE admins (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name    text NOT NULL,
    last_name     text NOT NULL,
    email         citext NOT NULL UNIQUE,
    username      text,
    password_hash text,
    is_active     boolean NOT NULL DEFAULT false,
    blocked       boolean NOT NULL DEFAULT false,
    department_id uuid REFERENCES departments (id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX admins_department_idx ON admins (department_id);

CREATE TABLE admin_users_roles (
    admin_id uuid NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
    role_id  uuid NOT NULL REFERENCES admin_roles (id) ON DELETE RESTRICT,
    PRIMARY KEY (admin_id, role_id)
);
CREATE INDEX admin_users_roles_role_idx ON admin_users_roles (role_id);

-- Cada refresh token é uma linha; a rotação revoga a linha usada e cria outra
-- na mesma família (family_id = ID da sessão exposto à API e ao JWT).
CREATE TABLE admin_sessions (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id           uuid NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
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
CREATE INDEX admin_sessions_admin_idx ON admin_sessions (admin_id);
CREATE INDEX admin_sessions_family_idx ON admin_sessions (family_id);

CREATE TABLE one_time_tokens (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type text NOT NULL,
    subject_id   uuid NOT NULL,
    purpose      text NOT NULL CHECK (purpose IN ('confirmation', 'reset', 'invite')),
    token_hash   text NOT NULL UNIQUE,
    expires_at   timestamptz NOT NULL,
    used_at      timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX one_time_tokens_subject_idx ON one_time_tokens (subject_type, subject_id, purpose);

-- +goose Down
DROP TABLE one_time_tokens;
DROP TABLE admin_sessions;
DROP TABLE admin_users_roles;
DROP TABLE admins;
DROP TABLE admin_roles;
