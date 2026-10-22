import type { Metadata } from 'next';
import { Suspense } from 'react';
import PageContainer from '@/components/layout/page-container';
import { Skeleton } from '@/components/ui/skeleton';
import { endUserActions } from '@/features/end-users/api/types';
import { EndUserFormDialog } from '@/features/end-users/components/end-user-form-dialog';
import { EndUsersList } from '@/features/end-users/components/end-users-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Usuários finais' };

export default async function UsuariosFinaisPage() {
  return (
    <PageContainer
      access={await hasPermission(endUserActions.users.read)}
      pageTitle='Usuários finais'
      pageDescription='Contas do portal público e da API pública.'
      pageHeaderAction={<EndUserFormDialog />}
    >
      <Suspense fallback={<Skeleton className='h-40 w-full' />}>
        <EndUsersList />
      </Suspense>
    </PageContainer>
  );
}
