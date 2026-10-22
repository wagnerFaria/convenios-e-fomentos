-- +goose Up
-- Arquivos anexados aos documentos. O conteúdo fica no armazenamento de objetos (MinIO); aqui ficam
-- só os metadados. A linha nasce "pendente" (envio autorizado) e vira "confirmado" quando a API
-- confere o objeto no armazenamento.
CREATE TABLE document_files (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id   uuid NOT NULL REFERENCES documents (id) ON DELETE CASCADE,
    -- gerada pelo servidor (nunca derivada do nome enviado pelo cliente)
    object_key    text NOT NULL UNIQUE,
    filename      text NOT NULL,
    content_type  text NOT NULL,
    size_bytes    bigint NOT NULL CHECK (size_bytes > 0),
    status        text NOT NULL DEFAULT 'pendente' CHECK (status IN ('pendente', 'confirmado')),
    created_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    confirmed_at  timestamptz,
    CONSTRAINT document_files_confirmed_at_matches_status CHECK ((status = 'confirmado') = (confirmed_at IS NOT NULL))
);
CREATE INDEX document_files_document_idx ON document_files (document_id, created_at);
CREATE INDEX document_files_pending_idx ON document_files (created_at) WHERE status = 'pendente';

-- +goose Down
DROP TABLE document_files;
