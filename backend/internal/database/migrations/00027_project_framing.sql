-- +goose Up

-- Enquadramento: depois da conferência, o Analista indica um departamento finalístico e o Revisor daquele departamento confirma ou
-- recusa. Cada indicação é uma rodada; a história fica em project_framings.
ALTER TABLE projects ADD COLUMN framing_status text CHECK (framing_status IN ('pending', 'confirmed'));

-- Confirmar leva o projeto a "Aguardando parecer".
ALTER TABLE projects DROP CONSTRAINT projects_stage_check;
ALTER TABLE projects ADD CONSTRAINT projects_stage_check
    CHECK (stage IN ('awaiting_checklist', 'in_analysis', 'awaiting_proponent', 'awaiting_opinion'));

-- Enquadramento em andamento ou confirmado sempre tem departamento. Se um departamento fosse excluído por fora do serviço (que
-- confere), o banco recusaria em vez de deixar o projeto sem área.
ALTER TABLE projects ADD CONSTRAINT projects_framing_consistency CHECK (framing_status IS NULL OR department_id IS NOT NULL);

CREATE TABLE project_framings (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    round           integer NOT NULL,
    -- O nome é copiado: o histórico sobrevive a renomeações e exclusões do departamento.
    department_id   uuid REFERENCES departments (id) ON DELETE SET NULL,
    department_name text NOT NULL,
    indicated_by_id uuid REFERENCES admins (id) ON DELETE SET NULL,
    indicated_at    timestamptz NOT NULL DEFAULT now(),
    decision        text NOT NULL DEFAULT 'pending' CHECK (decision IN ('pending', 'confirmed', 'refused')),
    decided_by_id   uuid REFERENCES admins (id) ON DELETE SET NULL,
    decided_at      timestamptz,
    justification   text NOT NULL DEFAULT '',
    UNIQUE (project_id, round),
    CONSTRAINT project_framings_decided_consistency CHECK ((decision = 'pending') = (decided_at IS NULL)),
    CONSTRAINT project_framings_refusal_justified CHECK (decision <> 'refused' OR justification <> '')
);
-- No máximo uma rodada pendente por projeto.
CREATE UNIQUE INDEX project_framings_pending_key ON project_framings (project_id) WHERE decision = 'pending';

-- Quem indica o departamento (Analista) e quem confirma ou recusa (Revisor do Departamento, só do próprio departamento).
INSERT INTO admin_permissions (role_id, action, subject, conditions)
SELECT r.id, d.action, 'project', d.conditions::jsonb
FROM (VALUES
    ('analyst', 'admin::projects.assign-department', '["admin::project-submitted"]'),
    ('department-reviewer', 'admin::projects.review-framing', '["admin::project-submitted-same-department"]')
) AS d (role_code, action, conditions)
JOIN admin_roles r ON r.code = d.role_code
WHERE NOT EXISTS (SELECT 1 FROM admin_permissions p WHERE p.role_id = r.id AND p.action = d.action)
ON CONFLICT DO NOTHING;

UPDATE admin_roles SET description = 'Vê os projetos enviados do próprio departamento e confirma ou recusa o enquadramento.'
WHERE code = 'department-reviewer' AND description = 'Vê os projetos enviados do próprio departamento.';

-- O cache de permissões não tem prazo de validade (ver a 00024): avisa as instâncias em execução.
SELECT pg_notify('authz_invalidate', '[]');

-- +goose Down
DELETE FROM admin_permissions p USING admin_roles r
WHERE p.role_id = r.id AND p.action IN ('admin::projects.assign-department', 'admin::projects.review-framing')
  AND r.code IN ('analyst', 'department-reviewer');
UPDATE admin_roles SET description = 'Vê os projetos enviados do próprio departamento.'
WHERE code = 'department-reviewer' AND description = 'Vê os projetos enviados do próprio departamento e confirma ou recusa o enquadramento.';
SELECT pg_notify('authz_invalidate', '[]');

DROP TABLE project_framings;
ALTER TABLE projects DROP CONSTRAINT projects_framing_consistency;
-- Falha se houver projeto em "Aguardando parecer": não se desce com dado que não cabe.
ALTER TABLE projects DROP CONSTRAINT projects_stage_check;
ALTER TABLE projects ADD CONSTRAINT projects_stage_check CHECK (stage IN ('awaiting_checklist', 'in_analysis', 'awaiting_proponent'));
ALTER TABLE projects DROP COLUMN framing_status;
