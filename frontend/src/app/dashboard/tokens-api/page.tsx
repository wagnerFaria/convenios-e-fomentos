import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { tokenActions } from '@/features/api-tokens/api/types';
import { TokensManager } from '@/features/api-tokens/components/tokens-manager';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Tokens de API' };

export default async function TokensApiPage() {
  return (
    <PageContainer
      access={await hasPermission(tokenActions.read)}
      pageTitle='Tokens de API'
      pageDescription='Credenciais para integrações sistema a sistema com a API pública.'
    >
      <TokensManager />
    </PageContainer>
  );
}
