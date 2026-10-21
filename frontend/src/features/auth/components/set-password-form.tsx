'use client';

import * as z from 'zod';
import { FieldGroup } from '@/components/ui/field';
import { useAppForm } from '@/lib/form';
import type { ActionResult } from '../action-result';
import { submitErrors } from '../submit-errors';
import { FormError } from './form-error';

const schema = z
  .object({
    password: z.string().min(8, 'A senha deve ter ao menos 8 caracteres.'),
    confirm: z.string()
  })
  .refine((v) => v.password === v.confirm, {
    path: ['confirm'],
    message: 'As senhas não conferem.'
  });

/** Definição de nova senha por token de uso único (redefinição e aceite de convite). */
export function SetPasswordForm({
  token,
  action,
  submitLabel
}: {
  token: string;
  action: (input: { token: string; password: string }) => Promise<ActionResult>;
  submitLabel: string;
}) {
  const form = useAppForm({
    defaultValues: { password: '', confirm: '' },
    validators: {
      onSubmit: schema,
      onSubmitAsync: async ({ value }) => {
        const result = await action({ token, password: value.password });
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
          name='password'
          children={(field) => (
            <field.TextField
              label='Nova senha'
              type='password'
              autoComplete='new-password'
              required
            />
          )}
        />
        <form.AppField
          name='confirm'
          children={(field) => (
            <field.TextField
              label='Confirmar senha'
              type='password'
              autoComplete='new-password'
              required
            />
          )}
        />
        <form.AppForm>
          <FormError />
          <form.SubmitButton>{submitLabel}</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}
