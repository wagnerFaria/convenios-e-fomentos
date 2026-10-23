import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { projectActions } from '@/features/projects/api/types';
import { NewProjectButton } from '@/features/projects/components/new-project-button';
import { ProjectsList } from '@/features/projects/components/projects-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Meus projetos' };

export default async function ProjetosPage() {
  return (
    <PageContainer
      access={await hasPermission(projectActions.read)}
      pageTitle='Meus projetos'
      pageDescription='Projetos de convênio e de termo de fomento que você cadastrou ou está cadastrando.'
      pageHeaderAction={<NewProjectButton />}
    >
      <ProjectsList />
    </PageContainer>
  );
}
