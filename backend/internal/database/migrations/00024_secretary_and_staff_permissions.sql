-- +goose Up

-- Só departamento marcado como área finalística pode ser indicado para enquadrar um projeto. Os que já existem ficam sem a
-- marca até alguém marcá-los na tela de departamentos (passo de implantação).
ALTER TABLE departments ADD COLUMN is_finalistic boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE departments DROP COLUMN is_finalistic;
