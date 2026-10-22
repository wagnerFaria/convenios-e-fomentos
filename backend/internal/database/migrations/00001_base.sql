-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE departments (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE
);

CREATE TABLE settings (
    key        text PRIMARY KEY,
    value      jsonb NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type text NOT NULL,
    actor_id   text,
    action     text NOT NULL,
    entity     text NOT NULL,
    entity_id  text,
    payload    jsonb NOT NULL DEFAULT '{}'::jsonb,
    ip         text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_actor_idx ON audit_logs (actor_type, actor_id);
CREATE INDEX audit_logs_entity_idx ON audit_logs (entity, entity_id);
CREATE INDEX audit_logs_action_idx ON audit_logs (action);
CREATE INDEX audit_logs_created_at_idx ON audit_logs (created_at DESC);

-- +goose Down
DROP TABLE audit_logs;
DROP TABLE settings;
DROP TABLE departments;
