-- +goose Up

-- SHA-256 do conteúdo, calculado pela API na confirmação do envio (lendo o objeto no armazenamento). Prova que o arquivo da versão
-- enviada é o mesmo que foi analisado. Arquivos anteriores a esta migração ficam sem hash.
ALTER TABLE project_files ADD COLUMN sha256 text CHECK (sha256 IS NULL OR sha256 ~ '^[0-9a-f]{64}$');

-- +goose Down
ALTER TABLE project_files DROP COLUMN sha256;
