import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { endUserActions } from '@/features/end-users/api/types';
import {
  EndUserRolesList,
  NewEndUserRoleDialog
} from '@/features/end-users/components/end-user-roles-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Perfis públicos' };

export default async function PerfisPublicosPage() {
  return (
    <PageContainer
      access={await hasPermission(endUserActions.roles.read)}
      pageTitle='Perfis públicos'
      pageDescription='Conjuntos de permissões na API pública para visitantes e usuários finais.'
      pageHeaderAction={<NewEndUserRoleDialog />}
    >
      <EndUserRolesList />
    </PageContainer>
  );
}
