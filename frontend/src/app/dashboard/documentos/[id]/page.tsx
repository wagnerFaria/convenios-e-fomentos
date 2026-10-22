import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { DocumentDetail } from '@/features/documents/components/document-detail';

export const metadata: Metadata = { title: 'Documento' };

export default async function DocumentoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer>
      <DocumentDetail id={id} />
    </PageContainer>
  );
}
