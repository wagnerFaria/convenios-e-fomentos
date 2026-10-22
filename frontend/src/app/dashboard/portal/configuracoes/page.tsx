import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { endUserActions } from '@/features/end-users/api/types';
import { EndUserSettingsPanel } from '@/features/end-users/components/end-user-settings-form';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Configurações do portal' };

export default async function ConfiguracoesPortalPage() {
  return (
    <PageContainer
      access={await hasPermission(endUserActions.settings.read)}
      pageTitle='Configurações do portal'
      pageDescription='Cadastro, confirmação de e-mail e perfil padrão dos usuários finais.'
    >
      <EndUserSettingsPanel />
    </PageContainer>
  );
}
