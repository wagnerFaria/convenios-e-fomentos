import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { adminActions } from '@/features/admins/api/types';
import { AdminDetail } from '@/features/admins/components/admin-detail';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Administrador' };

export default async function UsuarioPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer access={await hasPermission(adminActions.read)}>
      <AdminDetail id={id} />
    </PageContainer>
  );
}
