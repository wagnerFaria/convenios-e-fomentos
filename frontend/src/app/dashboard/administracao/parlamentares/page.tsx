import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { parliamentarianActions } from '@/features/parliamentarians/api/types';
import { ParliamentarianDialog } from '@/features/parliamentarians/components/parliamentarian-dialog';
import { ParliamentariansList } from '@/features/parliamentarians/components/parliamentarians-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Parlamentares' };

export default async function ParlamentaresPage() {
  return (
    <PageContainer
      access={await hasPermission(parliamentarianActions.read)}
      pageTitle='Parlamentares'
      pageDescription='Deputados que destinam emendas. O proponente escolhe nesta lista ao cadastrar uma emenda parlamentar.'
      pageHeaderAction={<ParliamentarianDialog />}
    >
      <ParliamentariansList />
    </PageContainer>
  );
}
