-- +goose Up

-- 1. Marcador de área finalística. Só departamento marcado pode ser indicado para enquadrar um projeto. Os que já existem ficam
-- sem a marca até alguém marcá-los na tela de departamentos (passo de implantação).
ALTER TABLE departments ADD COLUMN is_finalistic boolean NOT NULL DEFAULT false;

-- 2. Perfil Secretário de Estado.
INSERT INTO admin_roles (code, name, description, is_system) VALUES
    ('state-secretary', 'Secretário de Estado', 'Vê todos os projetos enviados e pode negá-los a qualquer momento, arquivando-os na hora.', true)
ON CONFLICT (code) DO NOTHING;

-- 3. Descrições dos perfis de equipe, só onde ainda são as originais (não apaga edição feita por quem administra).
UPDATE admin_roles SET description = 'Acompanha todos os projetos enviados, somente leitura. Recebe e-mails sobre documentação regular, enquadramento recusado e arquivamento.'
WHERE code = 'administrative-supervisor' AND description = 'Acompanha todos os projetos enviados e conduz a análise: pede complementação, arquiva ou nega.';
UPDATE admin_roles SET description = 'Assume a análise dos projetos enviados, confere documentos e informações, devolve para complementação, arquiva por falta de documentos obrigatórios e administra os departamentos.'
WHERE code = 'analyst' AND description = 'Vê todos os projetos enviados, confere a documentação, pede complementação ao proponente ou arquiva o projeto.';
UPDATE admin_roles SET description = 'Vê os projetos enviados do próprio departamento.'
WHERE code = 'department-reviewer' AND description = 'Lê documentos; só altera status e observação interna dos documentos do próprio departamento.';

-- 4. Revogações. O Supervisor deixa de decidir (passa a só acompanhar e receber e-mails): essas ações são do Analista
-- (complementar e arquivar) e do Secretário (negar). O Revisor deixa de ter as permissões do módulo genérico de documentos.
DELETE FROM admin_permissions p USING admin_roles r
WHERE p.role_id = r.id AND r.code = 'administrative-supervisor'
  AND p.action IN ('admin::projects.request-complement', 'admin::projects.archive', 'admin::projects.deny');
DELETE FROM admin_permissions p USING admin_roles r
WHERE p.role_id = r.id AND r.code = 'department-reviewer'
  AND p.action IN ('admin::content.document.read', 'admin::content.document.file.read', 'admin::content.document.update', 'admin::content.document.publish');

-- 5. A equipe só enxerga projeto enviado ou arquivado, nunca rascunho. A condição entra nas permissões de projeto que já existem,
-- só onde ainda não há condição nenhuma (preserva o que alguém tenha configurado).
UPDATE admin_permissions p SET conditions = '["admin::project-submitted"]'::jsonb, updated_at = now()
FROM admin_roles r
WHERE p.role_id = r.id AND r.code IN ('analyst', 'administrative-supervisor')
  AND p.action IN ('admin::projects.read', 'admin::projects.request-complement', 'admin::projects.archive')
  AND p.conditions = '[]'::jsonb;

-- 6. Permissões novas, idempotentes (só entra a ação que a role ainda não tem).
WITH defaults (role_code, action, subject, conditions) AS (VALUES
    ('state-secretary', 'admin::projects.read', 'project', '["admin::project-submitted"]'),
    ('state-secretary', 'admin::projects.deny', 'project', '["admin::project-submitted"]'),
    ('state-secretary', 'admin::catalog.entity-types.read', 'catalog.entity-types', '[]'),
    ('state-secretary', 'admin::catalog.municipalities.read', 'catalog.municipalities', '[]'),
    ('state-secretary', 'admin::catalog.parliamentarians.read', 'catalog.parliamentarians', '[]'),
    ('state-secretary', 'admin::settings.rule-parameters.read', 'settings.rule-parameters', '[]'),
    ('department-reviewer', 'admin::projects.read', 'project', '["admin::project-submitted-same-department"]'),
    ('department-reviewer', 'admin::catalog.entity-types.read', 'catalog.entity-types', '[]'),
    ('department-reviewer', 'admin::catalog.municipalities.read', 'catalog.municipalities', '[]'),
    ('department-reviewer', 'admin::catalog.parliamentarians.read', 'catalog.parliamentarians', '[]'),
    ('department-reviewer', 'admin::settings.rule-parameters.read', 'settings.rule-parameters', '[]'),
    ('analyst', 'admin::settings.departments.read', 'settings.departments', '[]'),
    ('analyst', 'admin::settings.departments.update', 'settings.departments', '[]'),
    ('analyst', 'admin::settings.departments.delete', 'settings.departments', '[]')
)
INSERT INTO admin_permissions (role_id, action, subject, conditions)
SELECT r.id, d.action, d.subject, d.conditions::jsonb
FROM defaults d
JOIN admin_roles r ON r.code = d.role_code
WHERE NOT EXISTS (SELECT 1 FROM admin_permissions p WHERE p.role_id = r.id AND p.action = d.action)
ON CONFLICT DO NOTHING;

-- 7. O cache de permissões não tem prazo de validade: avisa as instâncias em execução para recarregarem todos os perfis.
-- (A lista vazia significa "todos"; o aviso só é entregue quando esta transação confirma.)
SELECT pg_notify('authz_invalidate', '[]');

-- +goose Down

-- Devolve ao Supervisor as três ações e ao Revisor as quatro de documentos, como estavam na 00022.
INSERT INTO admin_permissions (role_id, action, subject, fields, conditions)
SELECT r.id, d.action, d.subject, d.fields::jsonb, d.conditions::jsonb
FROM (VALUES
    ('administrative-supervisor', 'admin::projects.request-complement', 'project', NULL, '[]'),
    ('administrative-supervisor', 'admin::projects.archive', 'project', NULL, '[]'),
    ('administrative-supervisor', 'admin::projects.deny', 'project', NULL, '[]'),
    ('department-reviewer', 'admin::content.document.read', 'document', NULL, '[]'),
    ('department-reviewer', 'admin::content.document.file.read', 'document', NULL, '[]'),
    ('department-reviewer', 'admin::content.document.update', 'document', '["observacao_interna","status"]', '["admin::same-department"]'),
    ('department-reviewer', 'admin::content.document.publish', 'document', NULL, '["admin::same-department"]')
) AS d (role_code, action, subject, fields, conditions)
JOIN admin_roles r ON r.code = d.role_code
WHERE NOT EXISTS (SELECT 1 FROM admin_permissions p WHERE p.role_id = r.id AND p.action = d.action)
ON CONFLICT DO NOTHING;

-- Tira as permissões que a Up acrescentou (as dos departamentos, só do Analista).
DELETE FROM admin_permissions p USING admin_roles r
WHERE p.role_id = r.id AND r.code = 'analyst'
  AND p.action IN ('admin::settings.departments.read', 'admin::settings.departments.update', 'admin::settings.departments.delete');
DELETE FROM admin_permissions p USING admin_roles r
WHERE p.role_id = r.id AND r.code = 'department-reviewer'
  AND p.action IN ('admin::projects.read', 'admin::catalog.entity-types.read', 'admin::catalog.municipalities.read',
                   'admin::catalog.parliamentarians.read', 'admin::settings.rule-parameters.read');

-- Tira a condição de visibilidade onde ela é exatamente a que a Up colocou.
UPDATE admin_permissions p SET conditions = '[]'::jsonb, updated_at = now()
FROM admin_roles r
WHERE p.role_id = r.id AND r.code IN ('analyst', 'administrative-supervisor')
  AND p.action IN ('admin::projects.read', 'admin::projects.request-complement', 'admin::projects.archive', 'admin::projects.deny')
  AND p.conditions = '["admin::project-submitted"]'::jsonb;

-- Descrições, só onde ainda são as da Up.
UPDATE admin_roles SET description = 'Acompanha todos os projetos enviados e conduz a análise: pede complementação, arquiva ou nega.'
WHERE code = 'administrative-supervisor' AND description = 'Acompanha todos os projetos enviados, somente leitura. Recebe e-mails sobre documentação regular, enquadramento recusado e arquivamento.';
UPDATE admin_roles SET description = 'Vê todos os projetos enviados, confere a documentação, pede complementação ao proponente ou arquiva o projeto.'
WHERE code = 'analyst' AND description = 'Assume a análise dos projetos enviados, confere documentos e informações, devolve para complementação, arquiva por falta de documentos obrigatórios e administra os departamentos.';
UPDATE admin_roles SET description = 'Lê documentos; só altera status e observação interna dos documentos do próprio departamento.'
WHERE code = 'department-reviewer' AND description = 'Vê os projetos enviados do próprio departamento.';

-- Só remove o perfil Secretário se ninguém o usa (a atribuição é ON DELETE RESTRICT); as permissões saem junto (CASCADE).
DELETE FROM admin_roles r
WHERE r.code = 'state-secretary'
  AND NOT EXISTS (SELECT 1 FROM admin_users_roles ur WHERE ur.role_id = r.id);

ALTER TABLE departments DROP COLUMN is_finalistic;

SELECT pg_notify('authz_invalidate', '[]');
