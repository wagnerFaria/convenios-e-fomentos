-- +goose Up

-- Ver o arquivo no visualizador (a primeira página apareceu) e lê-lo até o fim são provas diferentes: reprovar um documento só exige o
-- primeiro; aprovar exige o segundo. read_at fica nulo até o Analista chegar ao fim do arquivo.
ALTER TABLE project_file_views ADD COLUMN read_at timestamptz;

-- +goose Down
ALTER TABLE project_file_views DROP COLUMN read_at;
