import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { endUserActions } from '@/features/end-users/api/types';
import { EmailTemplatesPanel } from '@/features/end-users/components/email-templates-panel';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Templates de e-mail' };

export default async function TemplatesEmailPage() {
  return (
    <PageContainer
      access={await hasPermission(endUserActions.templates.read)}
      pageTitle='Templates de e-mail'
      pageDescription='Mensagens enviadas aos usuários finais (confirmação de e-mail e redefinição de senha).'
    >
      <EmailTemplatesPanel />
    </PageContainer>
  );
}
