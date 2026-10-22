import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { roleActions } from '@/features/roles/api/types';
import { NewRoleDialog } from '@/features/roles/components/new-role-dialog';
import { RolesList } from '@/features/roles/components/roles-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Perfis e permissões' };

export default async function PerfisPage() {
  return (
    <PageContainer
      access={await hasPermission(roleActions.read)}
      pageTitle='Perfis e permissões'
      pageDescription='Conjuntos de permissões que são atribuídos aos administradores.'
      pageHeaderAction={<NewRoleDialog />}
    >
      <RolesList />
    </PageContainer>
  );
}
