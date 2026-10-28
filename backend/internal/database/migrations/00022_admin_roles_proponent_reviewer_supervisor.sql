-- +goose Up
-- Perfis de sistema: Proponente, Revisor do Departamento e Supervisor Administrador. Idempotente: se o seed já criou a role,
-- nada muda; as mesmas permissões padrão do seed são aplicadas só onde ainda faltam.
INSERT INTO admin_roles (code, name, description, is_system) VALUES
    ('proponent', 'Proponente', 'Quem entra pelo MT Login e cadastra projetos de convênio ou fomento. Recebe esta role automaticamente no primeiro acesso.', true),
    ('department-reviewer', 'Revisor do Departamento', 'Lê documentos; só altera status e observação interna dos documentos do próprio departamento.', false),
    ('administrative-supervisor', 'Supervisor Administrador', 'Acompanha todos os projetos enviados e conduz a análise: pede complementação, arquiva ou nega.', true)
ON CONFLICT (code) DO NOTHING;

WITH defaults (role_code, action, subject, fields, conditions) AS (VALUES
    ('proponent', 'admin::catalog.entity-types.read', 'catalog.entity-types', NULL, '[]'),
    ('proponent', 'admin::catalog.municipalities.read', 'catalog.municipalities', NULL, '[]'),
    ('proponent', 'admin::catalog.parliamentarians.read', 'catalog.parliamentarians', NULL, '[]'),
    ('proponent', 'admin::settings.rule-parameters.read', 'settings.rule-parameters', NULL, '[]'),
    ('proponent', 'admin::proponent-profile.read', 'proponent-profile', NULL, '[]'),
    ('proponent', 'admin::proponent-profile.update', 'proponent-profile', NULL, '[]'),
    ('proponent', 'admin::projects.create', 'project', NULL, '[]'),
    ('proponent', 'admin::projects.read', 'project', NULL, '["admin::is-creator","admin::project-member"]'),
    ('proponent', 'admin::projects.update', 'project', NULL, '["admin::is-creator","admin::project-editor"]'),
    ('proponent', 'admin::projects.delete', 'project', NULL, '["admin::is-creator"]'),
    ('proponent', 'admin::projects.submit', 'project', NULL, '["admin::is-creator","admin::project-editor"]'),
    ('proponent', 'admin::projects.withdraw', 'project', NULL, '["admin::is-creator"]'),
    ('proponent', 'admin::projects.share', 'project', NULL, '["admin::is-creator"]'),
    ('department-reviewer', 'admin::content.document.read', 'document', NULL, '[]'),
    ('department-reviewer', 'admin::content.document.file.read', 'document', NULL, '[]'),
    ('department-reviewer', 'admin::content.document.update', 'document', '["observacao_interna","status"]', '["admin::same-department"]'),
    ('department-reviewer', 'admin::content.document.publish', 'document', NULL, '["admin::same-department"]'),
    ('administrative-supervisor', 'admin::projects.read', 'project', NULL, '[]'),
    ('administrative-supervisor', 'admin::projects.request-complement', 'project', NULL, '[]'),
    ('administrative-supervisor', 'admin::projects.archive', 'project', NULL, '[]'),
    ('administrative-supervisor', 'admin::projects.deny', 'project', NULL, '[]'),
    ('administrative-supervisor', 'admin::catalog.entity-types.read', 'catalog.entity-types', NULL, '[]'),
    ('administrative-supervisor', 'admin::catalog.municipalities.read', 'catalog.municipalities', NULL, '[]'),
    ('administrative-supervisor', 'admin::catalog.parliamentarians.read', 'catalog.parliamentarians', NULL, '[]'),
    ('administrative-supervisor', 'admin::settings.rule-parameters.read', 'settings.rule-parameters', NULL, '[]')
)
INSERT INTO admin_permissions (role_id, action, subject, fields, conditions)
SELECT r.id, d.action, d.subject, d.fields::jsonb, d.conditions::jsonb
FROM defaults d
JOIN admin_roles r ON r.code = d.role_code
-- Role que já existia mantém o que tem: só entra a ação que ainda não está lá.
WHERE NOT EXISTS (SELECT 1 FROM admin_permissions p WHERE p.role_id = r.id AND p.action = d.action)
ON CONFLICT DO NOTHING;

-- +goose Down
-- Só remove perfil que ninguém usa (a atribuição é ON DELETE RESTRICT).
DELETE FROM admin_roles r
WHERE r.code IN ('proponent', 'department-reviewer', 'administrative-supervisor')
  AND NOT EXISTS (SELECT 1 FROM admin_users_roles ur WHERE ur.role_id = r.id);
