-- +goose Up
CREATE TABLE admin_permissions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id    uuid NOT NULL REFERENCES admin_roles (id) ON DELETE CASCADE,
    action     text NOT NULL,
    subject    text,
    -- null = todos os campos
    fields     jsonb,
    -- lista de identificadores de condição; combinadas com OR
    conditions jsonb NOT NULL DEFAULT '[]'::jsonb,
    properties jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT admin_permissions_unique UNIQUE NULLS NOT DISTINCT (role_id, action, subject),
    CONSTRAINT admin_permissions_fields_array CHECK (fields IS NULL OR jsonb_typeof(fields) = 'array'),
    CONSTRAINT admin_permissions_conditions_array CHECK (jsonb_typeof(conditions) = 'array')
);
CREATE INDEX admin_permissions_action_idx ON admin_permissions (action);

-- +goose Down
DROP TABLE admin_permissions;
