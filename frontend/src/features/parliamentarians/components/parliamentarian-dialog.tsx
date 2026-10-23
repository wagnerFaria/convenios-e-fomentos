'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
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
import { parliamentarianKeys } from '../api/queries';
import { createParliamentarian, updateParliamentarian } from '../api/service';
import {
  chamberLabels,
  parliamentarianActions,
  type Chamber,
  type Parliamentarian,
  type ParliamentarianInput
} from '../api/types';

const schema = z.object({
  name: z.string().trim().min(1, 'Informe o nome.').max(120, 'Máximo de 120 caracteres.'),
  chamber: z.enum(['state', 'federal'], { error: 'Escolha a casa legislativa.' }),
  active: z.boolean()
});

function ParliamentarianForm({
  parliamentarian,
  onDone
}: {
  parliamentarian?: Parliamentarian;
  onDone: () => void;
}) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (input: ParliamentarianInput) =>
      parliamentarian
        ? updateParliamentarian(parliamentarian.id, input)
        : createParliamentarian(input),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: parliamentarianKeys.all });
      toast.success(parliamentarian ? 'Parlamentar atualizado.' : 'Parlamentar criado.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: {
      name: parliamentarian?.name ?? '',
      chamber: (parliamentarian?.chamber ?? 'state') as Chamber,
      active: parliamentarian?.active ?? true
    },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync({ ...value, name: value.name.trim() });
    }
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.AppField name='name' children={(f) => <f.TextField label='Nome' required />} />
        <form.AppField
          name='chamber'
          children={(f) => (
            <f.SelectField
              label='Casa legislativa'
              required
              options={Object.entries(chamberLabels).map(([value, label]) => ({ value, label }))}
            />
          )}
        />
        <form.AppField
          name='active'
          children={(f) => (
            <f.SwitchField
              label='Ativo'
              description='Parlamentares inativos deixam de aparecer ao cadastrar emendas.'
            />
          )}
        />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>{parliamentarian ? 'Salvar' : 'Criar parlamentar'}</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

/** Sem `parliamentarian`, é o botão "Novo parlamentar"; com ele, o diálogo (controlado) de edição. */
export function ParliamentarianDialog(
  props:
    | { parliamentarian: Parliamentarian; open: boolean; onOpenChange: (open: boolean) => void }
    | { parliamentarian?: undefined; open?: undefined; onOpenChange?: undefined }
) {
  const can = useCan();
  const [open, setOpen] = useState(false);

  if (props.parliamentarian) {
    return (
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Editar parlamentar</DialogTitle>
            <DialogDescription>{props.parliamentarian.name}</DialogDescription>
          </DialogHeader>
          <ParliamentarianForm
            parliamentarian={props.parliamentarian}
            onDone={() => props.onOpenChange(false)}
          />
        </DialogContent>
      </Dialog>
    );
  }
  if (!can(parliamentarianActions.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo parlamentar
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo parlamentar</DialogTitle>
          <DialogDescription>
            Aparece na lista de deputados ao cadastrar uma emenda parlamentar.
          </DialogDescription>
        </DialogHeader>
        <ParliamentarianForm onDone={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}
