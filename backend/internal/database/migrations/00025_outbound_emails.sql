-- +goose Up

-- Caixa de saída de e-mails. Quem muda o estado de um projeto grava a mensagem NA MESMA transação; um processo em segundo plano
-- (internal/outbox) envia, com tentativas e sem duplicar a criação (dedup_key). O corpo guarda dados pessoais, então as linhas
-- já resolvidas (sent/failed) são apagadas pela rotina de limpeza.
CREATE TABLE outbound_emails (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind            text NOT NULL,
    project_id      uuid REFERENCES projects (id) ON DELETE SET NULL,
    to_email        text NOT NULL,
    subject         text NOT NULL,
    body            text NOT NULL,
    -- Única: o mesmo evento para o mesmo destinatário entra uma só vez.
    dedup_key       text NOT NULL UNIQUE,
    status          text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    attempts        integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    -- Concessão: enquanto no futuro, outro processo não pega a mensagem. Se quem a pegou cair, a concessão vence e ela volta à fila.
    locked_until    timestamptz,
    last_error      text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    sent_at         timestamptz
);
CREATE INDEX outbound_emails_due_idx ON outbound_emails (next_attempt_at) WHERE status = 'pending';
CREATE INDEX outbound_emails_resolved_idx ON outbound_emails (created_at) WHERE status <> 'pending';

-- +goose Down
DROP TABLE outbound_emails;
