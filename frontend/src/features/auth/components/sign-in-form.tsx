'use client';

import Link from 'next/link';
import * as z from 'zod';
import { FieldGroup } from '@/components/ui/field';
import { useAppForm } from '@/lib/form';
import { signIn } from '../actions';
import { submitErrors } from '../submit-errors';
import { FormError } from './form-error';

const schema = z.object({
  email: z.email('Informe um e-mail válido.'),
  password: z.string().min(1, 'Informe a senha.')
});

export function SignInForm({ redirectTo }: { redirectTo?: string }) {
  const form = useAppForm({
    defaultValues: { email: '', password: '' },
    validators: {
      onSubmit: schema,
      onSubmitAsync: async ({ value }) => {
        const result = await signIn({ ...value, redirectTo });
        if (result && !result.ok) return submitErrors(result);
      }
    }
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.AppField
          name='email'
          children={(field) => (
            <field.TextField label='E-mail' type='email' autoComplete='username' required />
          )}
        />
        <form.AppField
          name='password'
          children={(field) => (
            <field.TextField
              label='Senha'
              type='password'
              autoComplete='current-password'
              required
            />
          )}
        />
        <form.AppForm>
          <FormError />
          <form.SubmitButton>Entrar</form.SubmitButton>
        </form.AppForm>
        <Link
          href='/auth/forgot-password'
          className='text-muted-foreground hover:text-foreground text-center text-sm underline underline-offset-4'
        >
          Esqueci minha senha
        </Link>
      </FieldGroup>
    </form>
  );
}
