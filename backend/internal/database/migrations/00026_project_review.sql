-- +goose Up

-- Analista responsável: quem assumiu a análise (nulo até alguém assumir). Informativo: qualquer Analista pode agir; serve às telas e
-- ao e-mail de arquivamento.
ALTER TABLE projects ADD COLUMN analyst_id uuid REFERENCES admins (id) ON DELETE SET NULL;

-- Conferência: situação de cada item analisado pela equipe. Ausência de linha = pendente (nada é criado no envio). O alvo usa as
-- mesmas chaves dos pedidos de complementação: nome da seção, "codigo" ou "codigo@id-da-emenda" do checklist, e o id de um
-- documento adicional.
CREATE TABLE project_item_reviews (
    project_id     uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    kind           text NOT NULL CHECK (kind IN ('section', 'document', 'extra')),
    target         text NOT NULL,
    status         text NOT NULL CHECK (status IN ('pending', 'approved')),
    reviewed_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    reviewed_at    timestamptz,
    PRIMARY KEY (project_id, kind, target)
);

-- Quem abriu qual arquivo. Aprovar um documento exige que o próprio Analista o tenha aberto. Arquivo novo tem outro id, então exige
-- nova abertura.
CREATE TABLE project_file_views (
    file_id   uuid NOT NULL REFERENCES project_files (id) ON DELETE CASCADE,
    admin_id  uuid NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
    opened_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (file_id, admin_id)
);
CREATE INDEX project_file_views_admin_idx ON project_file_views (admin_id);

-- Só o Analista assume a análise e confere os itens (como as demais ações dele, só sobre projeto enviado). Idempotente.
INSERT INTO admin_permissions (role_id, action, subject, conditions)
SELECT r.id, d.action, 'project', '["admin::project-submitted"]'::jsonb
FROM (VALUES ('admin::projects.start-analysis'), ('admin::projects.review-item')) AS d (action)
CROSS JOIN admin_roles r
WHERE r.code = 'analyst'
  AND NOT EXISTS (SELECT 1 FROM admin_permissions p WHERE p.role_id = r.id AND p.action = d.action)
ON CONFLICT DO NOTHING;

-- O cache de permissões não tem prazo de validade (ver a 00024): avisa as instâncias em execução.
SELECT pg_notify('authz_invalidate', '[]');

-- +goose Down
DELETE FROM admin_permissions p USING admin_roles r
WHERE p.role_id = r.id AND r.code = 'analyst' AND p.action IN ('admin::projects.start-analysis', 'admin::projects.review-item');
SELECT pg_notify('authz_invalidate', '[]');
DROP TABLE project_file_views;
DROP TABLE project_item_reviews;
ALTER TABLE projects DROP COLUMN analyst_id;
