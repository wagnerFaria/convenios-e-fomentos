import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { ProfileForm } from '@/features/proponent-profile/components/profile-form';
import { projectActions } from '@/features/projects/api/types';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Perfil de proponente' };

export default async function PerfilProponentePage() {
  return (
    <PageContainer
      access={await hasPermission(projectActions.profileRead)}
      pageTitle='Perfil de proponente'
      pageDescription='Seus contatos e cargo. Cada projeto novo começa com uma cópia destes dados; mudar o perfil depois não altera projetos já criados.'
    >
      <ProfileForm />
    </PageContainer>
  );
}
