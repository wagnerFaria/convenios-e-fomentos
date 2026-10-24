-- +goose Up

-- Obrigatoriedade da exigência. A aplicabilidade (instrumentos + condições) decide se o documento entra na lista do projeto; este
-- campo decide se, entrando, ele bloqueia a continuidade e o envio. Falso = opcional: aparece marcado e nunca bloqueia.
-- Todas as exigências existentes continuam obrigatórias.
ALTER TABLE document_requirements ADD COLUMN mandatory boolean NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE document_requirements DROP COLUMN mandatory;
