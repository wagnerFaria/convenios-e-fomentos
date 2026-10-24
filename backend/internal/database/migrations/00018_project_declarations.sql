-- +goose Up

-- Declarações que o sistema monta a partir do cadastro (internal/project/declarations.go) e que o representante legal assina no
-- MT-ID. Cada declaração tem o seu registro, o seu pedido de assinatura e o seu arquivo: não há pacote nem assinatura em lote.
-- Guarda o texto exato que foi aceito e o resumo (params_sha256) dos dados que o definem; se o cadastro mudar depois, o resumo
-- atual deixa de coincidir e a declaração assinada passa a valer como desatualizada.
--   awaiting_signature: pedido enviado, aguardando o signatário no MTCidadão
--   signed:             assinada; file_id é o arquivo assinado no checklist
--   rejected/expired:   o MT-ID não concluiu a assinatura
--   cancelled:          cancelada por quem pediu, ou trocada por um novo pedido com o cadastro alterado
--   superseded:         assinada, mas substituída por outra declaração mais nova
CREATE TABLE project_declarations (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    requirement_code text NOT NULL REFERENCES document_requirements (code) ON UPDATE CASCADE ON DELETE RESTRICT,
    project_version  integer NOT NULL,
    template_version text NOT NULL,
    params_sha256    text NOT NULL CHECK (params_sha256 ~ '^[0-9a-f]{64}$'),
    rendered_text    text NOT NULL,
    signer_name      text NOT NULL,
    signer_role      text NOT NULL,
    signer_cpf       text NOT NULL CHECK (signer_cpf ~ '^[0-9]{11}$'),
    accepted_by_id   uuid REFERENCES admins (id) ON DELETE SET NULL,
    accepted_by_cpf  text NOT NULL,
    accepted_at      timestamptz NOT NULL DEFAULT now(),
    accepted_ip      text NOT NULL DEFAULT '',
    status           text NOT NULL DEFAULT 'awaiting_signature'
        CHECK (status IN ('awaiting_signature', 'signed', 'rejected', 'expired', 'cancelled', 'superseded')),
    -- PDF enviado para assinatura (sem assinatura), no armazenamento de objetos.
    pdf_object_key   text NOT NULL UNIQUE,
    pdf_sha256       text NOT NULL CHECK (pdf_sha256 ~ '^[0-9a-f]{64}$'),
    mtid_file_id     text NOT NULL DEFAULT '',
    mtid_order_id    text NOT NULL DEFAULT '',
    signed_at        timestamptz,
    file_id          uuid REFERENCES project_files (id) ON DELETE SET NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX project_declarations_project_idx ON project_declarations (project_id, created_at);
-- No máximo uma declaração ativa (aguardando ou assinada) por exigência e projeto.
CREATE UNIQUE INDEX project_declarations_active_key ON project_declarations (project_id, requirement_code)
    WHERE status IN ('awaiting_signature', 'signed');

-- +goose Down
DROP TABLE project_declarations;
