import type { Metadata } from 'next';
import Link from 'next/link';
import { AuthShell } from '@/features/auth/components/auth-shell';
import { ForgotPasswordForm } from '@/features/auth/components/forgot-password-form';

export const metadata: Metadata = { title: 'Esqueci minha senha' };

export default function ForgotPasswordPage() {
  return (
    <AuthShell
      title='Esqueci minha senha'
      description='Informe seu e-mail para receber o link de redefinição.'
      footer={
        <Link href='/auth/sign-in' className='underline underline-offset-4'>
          Voltar ao login
        </Link>
      }
    >
      <ForgotPasswordForm />
    </AuthShell>
  );
}
