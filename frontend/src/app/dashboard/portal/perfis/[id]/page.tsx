import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { endUserActions } from '@/features/end-users/api/types';
import { EndUserRoleDetail } from '@/features/end-users/components/end-user-role-detail';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Perfil público' };

export default async function PerfilPublicoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer access={await hasPermission(endUserActions.roles.read)}>
      <EndUserRoleDetail id={id} />
    </PageContainer>
  );
}
