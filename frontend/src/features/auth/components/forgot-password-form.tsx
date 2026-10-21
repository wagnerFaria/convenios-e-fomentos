'use client';

import { useState } from 'react';
import * as z from 'zod';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { FieldGroup } from '@/components/ui/field';
import { useAppForm } from '@/lib/form';
import { forgotPassword } from '../actions';
import { submitErrors } from '../submit-errors';
import { FormError } from './form-error';

const schema = z.object({ email: z.email('Informe um e-mail válido.') });

export function ForgotPasswordForm() {
  const [sent, setSent] = useState(false);
  const form = useAppForm({
    defaultValues: { email: '' },
    validators: {
      onSubmit: schema,
      onSubmitAsync: async ({ value }) => {
        const result = await forgotPassword(value);
        if (result.ok) setSent(true);
        else return submitErrors(result);
      }
    }
  });

  if (sent) {
    return (
      <Alert>
        <AlertDescription>
          Se o e-mail estiver cadastrado, enviaremos as instruções para redefinir a senha.
        </AlertDescription>
      </Alert>
    );
  }

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
          children={(field) => <field.TextField label='E-mail' type='email' required />}
        />
        <form.AppForm>
          <FormError />
          <form.SubmitButton>Enviar instruções</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}
