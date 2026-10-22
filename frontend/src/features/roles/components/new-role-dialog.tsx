'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import * as z from 'zod';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { FieldGroup } from '@/components/ui/field';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { roleKeys } from '../api/queries';
import { createRole } from '../api/service';
import { roleActions } from '../api/types';

export const roleSchema = z.object({
  name: z.string().trim().min(1, 'Informe o nome.').max(100, 'Máximo de 100 caracteres.'),
  description: z.string().max(500, 'Máximo de 500 caracteres.')
});

export function NewRoleDialog() {
  const can = useCan();
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const qc = useQueryClient();

  const create = useMutation({
    mutationFn: createRole,
    onSuccess: async (role) => {
      await qc.invalidateQueries({ queryKey: roleKeys.all });
      setOpen(false);
      router.push(`/dashboard/administracao/perfis/${role.id}`);
    },
    onError: (err) => setError(errorMessage(err))
  });

  const form = useAppForm({
    defaultValues: { name: '', description: '' },
    validators: { onSubmit: roleSchema },
    onSubmit: ({ value }) => {
      setError(null);
      return create.mutateAsync({ name: value.name.trim(), description: value.description });
    }
  });

  if (!can(roleActions.create)) return null;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo perfil
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo perfil</DialogTitle>
          <DialogDescription>
            Ele nasce sem permissões; você as define na próxima tela.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void form.handleSubmit();
          }}
        >
          <FieldGroup>
            <form.AppField name='name' children={(f) => <f.TextField label='Nome' required />} />
            <form.AppField
              name='description'
              children={(f) => <f.TextareaField label='Descrição' />}
            />
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            <form.AppForm>
              <form.SubmitButton>Criar perfil</form.SubmitButton>
            </form.AppForm>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
