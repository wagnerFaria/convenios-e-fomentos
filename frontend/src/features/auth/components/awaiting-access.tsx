import { Button } from '@/components/ui/button';
import { signOut } from '../actions';
import { AuthShell } from './auth-shell';

/** Quem entra pela primeira vez já tem conta, mas nenhum perfil: espera um administrador liberar. */
export function AwaitingAccess({ name }: { name: string }) {
  return (
    <AuthShell
      title='Aguardando liberação'
      description={`${name}, o seu acesso foi criado, mas ainda não há um perfil atribuído. Solicite a um administrador que libere as suas permissões e entre novamente.`}
    >
      <form action={signOut}>
        <Button type='submit' variant='outline' className='w-full'>
          Sair
        </Button>
      </form>
    </AuthShell>
  );
}
