import type { Metadata } from 'next';
import { Suspense } from 'react';
import PageContainer from '@/components/layout/page-container';
import { Skeleton } from '@/components/ui/skeleton';
import { StaffQueue } from '@/features/projects/components/staff-queue';
import { isProjectStaffSession } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Análise de projetos' };

export default async function AnalisePage() {
  return (
    <PageContainer
      access={await isProjectStaffSession()}
      pageTitle='Análise de projetos'
      pageDescription='Fila de trabalho da equipe: projetos aguardando análise, enquadramento ou parecer.'
    >
      <Suspense fallback={<Skeleton className='h-60 w-full' />}>
        <StaffQueue />
      </Suspense>
    </PageContainer>
  );
}
