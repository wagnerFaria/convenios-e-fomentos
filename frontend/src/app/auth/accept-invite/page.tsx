import type { Metadata } from 'next';
import { AuthShell } from '@/features/auth/components/auth-shell';
import { SetPasswordForm } from '@/features/auth/components/set-password-form';
import { acceptInvite } from '@/features/auth/actions';

export const metadata: Metadata = { title: 'Aceitar convite' };

export default async function AcceptInvitePage({
  searchParams
}: {
  searchParams: Promise<{ token?: string }>;
}) {
  const { token } = await searchParams;
  return (
    <AuthShell
      title='Aceitar convite'
      description={token ? 'Defina sua senha para ativar a conta.' : 'Link inválido ou incompleto.'}
    >
      {token && <SetPasswordForm token={token} action={acceptInvite} submitLabel='Ativar conta' />}
    </AuthShell>
  );
}
