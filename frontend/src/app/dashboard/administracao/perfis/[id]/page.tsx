import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { roleActions } from '@/features/roles/api/types';
import { RoleDetail } from '@/features/roles/components/role-detail';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Perfil' };

export default async function PerfilPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer access={await hasPermission(roleActions.read)}>
      <RoleDetail id={id} />
    </PageContainer>
  );
}
