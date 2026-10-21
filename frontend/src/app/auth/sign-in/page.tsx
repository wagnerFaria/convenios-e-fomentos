import type { Metadata } from 'next';
import Link from 'next/link';
import { redirect } from 'next/navigation';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AuthShell } from '@/features/auth/components/auth-shell';
import { SignInForm } from '@/features/auth/components/sign-in-form';
import { safeRedirect } from '@/features/auth/action-result';
import { getSession } from '@/lib/auth/session';

export const metadata: Metadata = { title: 'Entrar' };

export default async function SignInPage({
  searchParams
}: {
  searchParams: Promise<{ redirect?: string; reset?: string }>;
}) {
  const { redirect: target, reset } = await searchParams;
  if (await getSession()) redirect(safeRedirect(target));

  return (
    <AuthShell
      title='Entrar'
      description='Acesse o painel administrativo.'
      footer={
        <Link href='/auth/setup' className='underline underline-offset-4'>
          Primeiro acesso? Configure o administrador
        </Link>
      }
    >
      {reset && (
        <Alert>
          <AlertDescription>Senha redefinida. Entre com a nova senha.</AlertDescription>
        </Alert>
      )}
      <SignInForm redirectTo={target} />
    </AuthShell>
  );
}
