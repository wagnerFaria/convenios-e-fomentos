import type { Metadata } from 'next';
import Link from 'next/link';
import { AuthShell } from '@/features/auth/components/auth-shell';
import { SetPasswordForm } from '@/features/auth/components/set-password-form';
import { resetPassword } from '@/features/auth/actions';

export const metadata: Metadata = { title: 'Redefinir senha' };

export default async function ResetPasswordPage({
  searchParams
}: {
  searchParams: Promise<{ token?: string }>;
}) {
  const { token } = await searchParams;
  return (
    <AuthShell
      title='Redefinir senha'
      description={token ? 'Escolha uma nova senha.' : 'Link inválido ou incompleto.'}
      footer={
        <Link href='/auth/sign-in' className='underline underline-offset-4'>
          Voltar ao login
        </Link>
      }
    >
      {token && (
        <SetPasswordForm token={token} action={resetPassword} submitLabel='Redefinir senha' />
      )}
    </AuthShell>
  );
}
