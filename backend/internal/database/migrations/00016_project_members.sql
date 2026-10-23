-- +goose Up

-- Compartilhamento do projeto por CPF. O titular (projects.created_by_id) convida pelo CPF, com nível de leitura ou de edição. O acesso vale
-- para quem tem conta com aquele CPF (admins.cpf) e é conferido a cada requisição, então um convite feito antes do primeiro login já vale
-- assim que a pessoa entra, e a revogação vale na hora. Não há aceite: accepted_at marca só a primeira vez que o convidado abriu o projeto.
CREATE TABLE project_members (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    cpf           text NOT NULL CHECK (cpf ~ '^[0-9]{11}$'),
    -- Como o titular identifica a pessoa (a API não revela o nome de quem tem conta).
    label         text NOT NULL DEFAULT '' CHECK (char_length(label) <= 100),
    level         text NOT NULL CHECK (level IN ('read', 'edit')),
    status        text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    invited_by_id uuid NOT NULL REFERENCES admins (id),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    accepted_at   timestamptz,
    revoked_at    timestamptz,
    revoked_by_id uuid REFERENCES admins (id),
    CONSTRAINT project_members_unique UNIQUE (project_id, cpf)
);
CREATE INDEX project_members_cpf_idx ON project_members (cpf) WHERE status = 'active';

-- A role Proponente passa a enxergar também os projetos em que é membro: leitura para qualquer nível, alteração e envio para quem tem
-- edição. Excluir, desistir e gerir o acesso continuam só do titular. Só mexe em permissão que ainda é "só o criador".
UPDATE admin_permissions p SET conditions = p.conditions || '["admin::project-member"]'::jsonb, updated_at = now()
FROM admin_roles r
WHERE p.role_id = r.id AND r.code = 'proponent' AND p.action = 'admin::projects.read'
  AND p.conditions = '["admin::is-creator"]'::jsonb;
UPDATE admin_permissions p SET conditions = p.conditions || '["admin::project-editor"]'::jsonb, updated_at = now()
FROM admin_roles r
WHERE p.role_id = r.id AND r.code = 'proponent' AND p.action IN ('admin::projects.update', 'admin::projects.submit')
  AND p.conditions = '["admin::is-creator"]'::jsonb;
INSERT INTO admin_permissions (role_id, action, subject, conditions)
SELECT r.id, 'admin::projects.share', 'project', '["admin::is-creator"]'::jsonb FROM admin_roles r WHERE r.code = 'proponent'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM admin_permissions p USING admin_roles r WHERE p.role_id = r.id AND r.code = 'proponent' AND p.action = 'admin::projects.share';
UPDATE admin_permissions p SET conditions = '["admin::is-creator"]'::jsonb
FROM admin_roles r
WHERE p.role_id = r.id AND r.code = 'proponent' AND p.action IN ('admin::projects.read', 'admin::projects.update', 'admin::projects.submit');
DROP TABLE project_members;
