-- +goose Up

-- Dados pessoais de quem assina as declarações geradas pelo sistema (nacionalidade, estado civil e endereço residencial). O CPF
-- identifica a pessoa: não há RG. Formato: {"nationality","maritalStatus","zip","street","number","complement","district","uf",
-- "municipalityIbge"}. O perfil guarda os dados do próprio proponente para reaproveitar em cada projeto; o projeto guarda uma cópia
-- do proponente e os do representante legal, quando é outra pessoa.
ALTER TABLE proponent_profiles ADD COLUMN personal jsonb NOT NULL DEFAULT '{}';
ALTER TABLE projects
    ADD COLUMN proponent_personal  jsonb NOT NULL DEFAULT '{}',
    ADD COLUMN legal_rep_personal  jsonb NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE projects DROP COLUMN legal_rep_personal, DROP COLUMN proponent_personal;
ALTER TABLE proponent_profiles DROP COLUMN personal;
