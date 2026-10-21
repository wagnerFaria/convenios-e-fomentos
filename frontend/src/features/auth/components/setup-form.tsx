'use client';

import * as z from 'zod';
import { FieldGroup } from '@/components/ui/field';
import { useAppForm } from '@/lib/form';
import { registerFirstAdmin } from '../actions';
import { submitErrors } from '../submit-errors';
import { FormError } from './form-error';

const schema = z.object({
  firstName: z.string().trim().min(1, 'Informe o nome.'),
  lastName: z.string().trim().min(1, 'Informe o sobrenome.'),
  email: z.email('Informe um e-mail válido.'),
  password: z.string().min(8, 'A senha deve ter ao menos 8 caracteres.')
});

export function SetupForm() {
  const form = useAppForm({
    defaultValues: { firstName: '', lastName: '', email: '', password: '' },
    validators: {
      onSubmit: schema,
      onSubmitAsync: async ({ value }) => {
        const result = await registerFirstAdmin(value);
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
          name='firstName'
          children={(field) => <field.TextField label='Nome' autoComplete='given-name' required />}
        />
        <form.AppField
          name='lastName'
          children={(field) => (
            <field.TextField label='Sobrenome' autoComplete='family-name' required />
          )}
        />
        <form.AppField
          name='email'
          children={(field) => (
            <field.TextField label='E-mail' type='email' autoComplete='email' required />
          )}
        />
        <form.AppField
          name='password'
          children={(field) => (
            <field.TextField
              label='Senha'
              type='password'
              autoComplete='new-password'
              description='Mínimo de 8 caracteres.'
              required
            />
          )}
        />
        <form.AppForm>
          <FormError />
          <form.SubmitButton>Criar administrador</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}
