import { fullName } from '@/features/auth/full-name';
import { getSession } from '@/lib/auth/session';

export default async function OverviewPage() {
  const session = await getSession();
  return (
    <p className='text-muted-foreground'>
      Olá, {session ? fullName(session) : 'administrador'}. Os módulos de convênios e fomentos
      aparecerão aqui.
    </p>
  );
}
