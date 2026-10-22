import type { Metadata } from 'next';
import { Suspense } from 'react';
import PageContainer from '@/components/layout/page-container';
import { Skeleton } from '@/components/ui/skeleton';
import { adminActions } from '@/features/admins/api/types';
import { AdminsList } from '@/features/admins/components/admins-list';
import { NewAdminDialog } from '@/features/admins/components/new-admin-dialog';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Usuários' };

export default async function UsuariosPage() {
  return (
    <PageContainer
      access={await hasPermission(adminActions.read)}
      pageTitle='Usuários'
      pageDescription='Administradores do painel, seus perfis e quem aguarda liberação de acesso.'
      pageHeaderAction={<NewAdminDialog />}
    >
      <Suspense fallback={<Skeleton className='h-40 w-full' />}>
        <AdminsList />
      </Suspense>
    </PageContainer>
  );
}
