-- +goose Up
-- Perfil de sistema Analista: confere os projetos enviados e devolve para complementação ou arquiva (ver docs/planejamento-cadastro-projeto.md, 3.2).
-- Idempotente: se o seed já criou a role, só entram as ações que ainda faltam.
INSERT INTO admin_roles (code, name, description, is_system) VALUES
    ('analyst', 'Analista', 'Vê todos os projetos enviados, confere a documentação, pede complementação ao proponente ou arquiva o projeto.', true)
ON CONFLICT (code) DO NOTHING;

WITH defaults (action, subject) AS (VALUES
    ('admin::projects.read', 'project'),
    ('admin::projects.request-complement', 'project'),
    ('admin::projects.archive', 'project'),
    ('admin::catalog.entity-types.read', 'catalog.entity-types'),
    ('admin::catalog.municipalities.read', 'catalog.municipalities'),
    ('admin::catalog.parliamentarians.read', 'catalog.parliamentarians'),
    ('admin::settings.rule-parameters.read', 'settings.rule-parameters')
)
INSERT INTO admin_permissions (role_id, action, subject, conditions)
SELECT r.id, d.action, d.subject, '[]'::jsonb
FROM defaults d
CROSS JOIN admin_roles r
WHERE r.code = 'analyst'
  AND NOT EXISTS (SELECT 1 FROM admin_permissions p WHERE p.role_id = r.id AND p.action = d.action)
ON CONFLICT DO NOTHING;

-- +goose Down
-- Só remove o perfil se ninguém o usa (a atribuição é ON DELETE RESTRICT).
DELETE FROM admin_roles r
WHERE r.code = 'analyst'
  AND NOT EXISTS (SELECT 1 FROM admin_users_roles ur WHERE ur.role_id = r.id);
