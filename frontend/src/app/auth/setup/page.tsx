import type { Metadata } from 'next';
import Link from 'next/link';
import { AuthShell } from '@/features/auth/components/auth-shell';
import { SetupForm } from '@/features/auth/components/setup-form';

export const metadata: Metadata = { title: 'Configuração inicial' };

export default function SetupPage() {
  return (
    <AuthShell
      title='Configuração inicial'
      description='Crie o primeiro Super Admin. Disponível apenas enquanto não houver administradores.'
      footer={
        <Link href='/auth/sign-in' className='underline underline-offset-4'>
          Voltar ao login
        </Link>
      }
    >
      <SetupForm />
    </AuthShell>
  );
}
