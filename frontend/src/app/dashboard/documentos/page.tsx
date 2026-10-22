import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { DocumentsList } from '@/features/documents/components/documents-list';
import { NewDocumentDialog } from '@/features/documents/components/new-document-dialog';

export const metadata: Metadata = { title: 'Documentos' };

export default function DocumentosPage() {
  return (
    <PageContainer
      pageTitle='Documentos'
      pageDescription='Documentos e seus arquivos anexados.'
      pageHeaderAction={<NewDocumentDialog />}
    >
      <DocumentsList />
    </PageContainer>
  );
}
