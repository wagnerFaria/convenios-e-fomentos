import type { Metadata } from 'next';
import { Button } from '@/components/ui/button';
import { signOut } from '@/features/auth/actions';
import { AuthShell } from '@/features/auth/components/auth-shell';

export const metadata: Metadata = { title: 'Acesso negado' };

const MESSAGES: Record<string, string> = {
  account_disabled: 'O seu acesso está desativado. Procure um administrador do sistema.',
  invalid_identity: 'Não foi possível identificar o CPF da sua conta do MT Login.'
};

// Esta página não consulta a sessão de propósito: quem chega aqui já foi recusado pelo backend.
export default async function DeniedPage({
  searchParams
}: {
  searchParams: Promise<{ code?: string }>;
}) {
  const { code } = await searchParams;
  return (
    <AuthShell
      title='Acesso negado'
      description={(code && MESSAGES[code]) || 'Você não pode acessar o painel.'}
    >
      <form action={signOut}>
        <Button type='submit' variant='outline' className='w-full'>
          Sair
        </Button>
      </form>
    </AuthShell>
  );
}
