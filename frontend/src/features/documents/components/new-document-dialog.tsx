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
import { useAppForm } from '@/lib/form';
import { documentKeys } from '../api/queries';
import { createDocument } from '../api/service';
import { ApiError } from '@/lib/api/errors';

const schema = z.object({
  titulo: z.string().trim().min(1, 'Informe o título.').max(200, 'Máximo de 200 caracteres.'),
  conteudo: z.string()
});

export function NewDocumentDialog() {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const qc = useQueryClient();

  const create = useMutation({
    mutationFn: createDocument,
    onSuccess: async (doc) => {
      await qc.invalidateQueries({ queryKey: documentKeys.all });
      setOpen(false);
      router.push(`/dashboard/documentos/${doc.id}`);
    },
    onError: (err) => setError(err instanceof ApiError ? err.message : 'Erro inesperado.')
  });

  const form = useAppForm({
    defaultValues: { titulo: '', conteudo: '' },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return create.mutateAsync({ titulo: value.titulo.trim(), conteudo: value.conteudo });
    }
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo documento
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo documento</DialogTitle>
          <DialogDescription>
            Ele nasce como rascunho; depois você anexa os arquivos.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void form.handleSubmit();
          }}
        >
          <FieldGroup>
            <form.AppField
              name='titulo'
              children={(f) => <f.TextField label='Título' required />}
            />
            <form.AppField name='conteudo' children={(f) => <f.TextareaField label='Conteúdo' />} />
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            <form.AppForm>
              <form.SubmitButton>Criar documento</form.SubmitButton>
            </form.AppForm>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
