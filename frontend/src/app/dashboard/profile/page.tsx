import type { Metadata } from 'next';
import { redirect } from 'next/navigation';
import PageContainer from '@/components/layout/page-container';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { fullName } from '@/features/auth/full-name';
import { getSession } from '@/lib/auth/session';

export const metadata: Metadata = { title: 'Meu perfil' };

export default async function ProfilePage() {
  const session = await getSession();
  if (!session) redirect('/auth/sign-in');

  return (
    <PageContainer pageTitle='Meu perfil' pageDescription='Seus dados e perfis de acesso.'>
      <div className='grid max-w-3xl gap-4'>
        <Card>
          <CardHeader>
            <CardTitle>{fullName(session)}</CardTitle>
            <CardDescription>
              Nome, CPF e e-mail são do MT Login e se atualizam a cada acesso.
            </CardDescription>
          </CardHeader>
          <CardContent className='grid gap-4'>
            <dl className='grid gap-2 text-sm sm:grid-cols-[8rem_1fr]'>
              <dt className='text-muted-foreground'>CPF</dt>
              <dd>{session.cpf}</dd>
              <dt className='text-muted-foreground'>E-mail</dt>
              <dd>{session.email || '—'}</dd>
            </dl>
            <div className='flex flex-wrap gap-2'>
              {session.isSuperAdmin && <Badge>Super Admin</Badge>}
              {session.roles
                .filter((role) => !(session.isSuperAdmin && role.name === 'Super Admin'))
                .map((role) => (
                  <Badge key={role.id} variant='secondary'>
                    {role.name}
                  </Badge>
                ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}
