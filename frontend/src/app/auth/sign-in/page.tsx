import type { Metadata } from 'next';
import { redirect } from 'next/navigation';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { buttonVariants } from '@/components/ui/button';
import { AuthShell } from '@/features/auth/components/auth-shell';
import { safeRedirect } from '@/features/auth/action-result';
import { getSession } from '@/lib/auth/session';
import { cn } from '@/lib/utils';

export const metadata: Metadata = { title: 'Entrar' };

const ERRORS: Record<string, string> = {
  access_denied: 'O login foi cancelado. Tente novamente quando quiser.',
  expired: 'A tentativa de login expirou. Tente novamente.',
  provider_error: 'O MT Login não concluiu o login. Tente novamente.',
  unavailable: 'Não foi possível contatar o MT Login. Tente novamente em instantes.',
  token_too_large: 'A resposta do MT Login excede o limite de armazenamento. Avise o suporte.',
  config: 'O login pelo MT Login não está configurado neste ambiente.'
};

export default async function SignInPage({
  searchParams
}: {
  searchParams: Promise<{ redirect?: string; error?: string }>;
}) {
  const { redirect: target, error } = await searchParams;
  if (await getSession()) redirect(safeRedirect(target));

  // Link comum (não <Link>): o /auth/login é um route handler que redireciona para fora do app.
  const loginHref = `/auth/login${target ? `?redirect=${encodeURIComponent(target)}` : ''}`;

  return (
    <AuthShell
      title='Entrar'
      description='Acesse o painel administrativo com a sua conta do MT Login.'
      footer='O acesso é feito pelo MT Login. Se for o seu primeiro acesso, peça a um administrador que libere o seu perfil.'
    >
      {error && (
        <Alert variant='destructive'>
          <AlertDescription>{ERRORS[error] ?? ERRORS.unavailable}</AlertDescription>
        </Alert>
      )}
      <a href={loginHref} className={cn(buttonVariants({ size: 'lg' }), 'w-full')}>
        Entrar com MT Login
      </a>
    </AuthShell>
  );
}
