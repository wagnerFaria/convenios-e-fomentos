import { fullName } from '@/features/auth/full-name';
import { getSession } from '@/lib/auth/session';

export default async function OverviewPage() {
  const session = await getSession();
  const isProponent = session?.roles.some((r) => r.code === 'proponent') ?? false;
  return (
    <p className='text-muted-foreground'>
      Olá, {session ? fullName(session) : 'bem-vindo'}.{' '}
      {isProponent
        ? 'Comece pelo seu perfil de proponente e depois cadastre o seu projeto de convênio ou de termo de fomento em "Meus projetos".'
        : 'Os módulos de convênios e fomentos aparecerão aqui.'}
    </p>
  );
}
