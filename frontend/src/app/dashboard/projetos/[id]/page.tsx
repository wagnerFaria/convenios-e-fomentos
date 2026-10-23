import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { projectActions } from '@/features/projects/api/types';
import { ProjectWizard } from '@/features/projects/components/project-wizard';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Projeto' };

export default async function ProjetoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer
      access={await hasPermission(projectActions.read)}
      pageTitle='Cadastro do projeto'
      pageDescription='Preencha as etapas e salve cada uma. O rascunho pode ficar incompleto e ser retomado depois.'
    >
      <ProjectWizard id={id} />
    </PageContainer>
  );
}
