-- +goose Up
-- A identidade do administrador passa a ser o CPF vindo do MT Login (OIDC); a senha local deixa de ser
-- usada. As colunas password_hash e a tabela admin_sessions seguem no esquema, sem uso, até uma migração
-- posterior removê-las depois de validado o novo login.
ALTER TABLE admins
    ADD COLUMN cpf           text,
    ADD COLUMN last_login_at timestamptz,
    ADD CONSTRAINT admins_cpf_format CHECK (cpf IS NULL OR cpf ~ '^[0-9]{11}$'),
    ADD CONSTRAINT admins_cpf_key UNIQUE (cpf);

-- O e-mail é um atributo sincronizado a cada login, não identidade: pode repetir ou vir vazio.
ALTER TABLE admins DROP CONSTRAINT admins_email_key;
CREATE INDEX admins_email_idx ON admins (email);

-- +goose Down
DROP INDEX admins_email_idx;
ALTER TABLE admins ADD CONSTRAINT admins_email_key UNIQUE (email);
ALTER TABLE admins
    DROP CONSTRAINT admins_cpf_key,
    DROP CONSTRAINT admins_cpf_format,
    DROP COLUMN last_login_at,
    DROP COLUMN cpf;
