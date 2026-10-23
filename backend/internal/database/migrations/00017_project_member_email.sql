-- +goose Up

-- E-mail informado pelo titular para avisar o convidado. Só vale para quem ainda não tem conta: quem já tem recebe o aviso no e-mail da
-- própria conta, lido na hora do envio e nunca exposto ao titular.
ALTER TABLE project_members ADD COLUMN email text CHECK (email IS NULL OR (char_length(email) <= 254 AND email = lower(email)));

-- +goose Down
ALTER TABLE project_members DROP COLUMN email;
