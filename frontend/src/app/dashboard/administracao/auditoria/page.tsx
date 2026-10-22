import type { Metadata } from 'next';
import { Suspense } from 'react';
import PageContainer from '@/components/layout/page-container';
import { Skeleton } from '@/components/ui/skeleton';
import { auditActions } from '@/features/audit-logs/api/types';
import { AuditList } from '@/features/audit-logs/components/audit-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Auditoria' };

export default async function AuditoriaPage() {
  return (
    <PageContainer
      access={await hasPermission(auditActions.read)}
      pageTitle='Auditoria'
      pageDescription='Quem alterou o quê e quando. Clique em um registro para ver o antes e o depois.'
    >
      <Suspense fallback={<Skeleton className='h-40 w-full' />}>
        <AuditList />
      </Suspense>
    </PageContainer>
  );
}
