'use client';

import { Icons } from '@/components/icons';
import { Alert, AlertDescription } from '@/components/ui/alert';

/** Código do perfil de sistema "Revisor do Departamento" (criado pela migração 00022 do backend). */
export const REVIEWER_ROLE_CODE = 'department-reviewer';

/**
 * O Revisor do Departamento só enxerga os projetos do departamento dele: sem departamento no cadastro, a API nega tudo (403) e a
 * pessoa não vê projeto algum. Avisa quem está atribuindo o perfil.
 */
export function ReviewerDepartmentWarning({ show }: { show: boolean }) {
  if (!show) return null;
  return (
    <Alert>
      <Icons.alertCircle aria-hidden />
      <AlertDescription>
        O Revisor do Departamento só enxerga os projetos do departamento dele. Sem departamento, não
        enxerga nenhum projeto: escolha o departamento.
      </AlertDescription>
    </Alert>
  );
}
