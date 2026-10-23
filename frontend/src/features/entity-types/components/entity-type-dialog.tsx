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
import { entityTypeKeys } from '../api/queries';
import { createEntityType, updateEntityType } from '../api/service';
import {
  entityTypeActions,
  natureLabels,
  type EntityNature,
  type EntityType,
  type EntityTypeInput
} from '../api/types';

const schema = z
  .object({
    code: z.string(),
    name: z.string().trim().min(1, 'Informe o nome.').max(100, 'Máximo de 100 caracteres.'),
    nature: z.enum(['public', 'private'], { error: 'Escolha a natureza.' }),
    nonProfit: z.boolean(),
    active: z.boolean(),
    sortOrder: z
      .number({ error: 'Informe a ordem.' })
      .int('Use um número inteiro.')
      .min(0, 'Use de 0 a 9999.')
      .max(9999, 'Use de 0 a 9999.')
  })
  .refine((v) => !v.nonProfit || v.nature === 'private', {
    path: ['nonProfit'],
    message: 'Só entidades privadas podem ser sem fins lucrativos.'
  });

const codeRule = /^[a-z][a-z0-9_]{1,39}$/;

function EntityTypeForm({ entityType, onDone }: { entityType?: EntityType; onDone: () => void }) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (input: EntityTypeInput) =>
      entityType ? updateEntityType(entityType.id, input) : createEntityType(input),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: entityTypeKeys.all });
      toast.success(entityType ? 'Tipo de entidade atualizado.' : 'Tipo de entidade criado.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: {
      code: entityType?.code ?? '',
      name: entityType?.name ?? '',
      nature: (entityType?.nature ?? 'private') as EntityNature,
      nonProfit: entityType?.nonProfit ?? false,
      active: entityType?.active ?? true,
      sortOrder: entityType?.sortOrder ?? 500
    },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      if (!entityType && !codeRule.test(value.code)) {
        setError(
          'O código usa de 2 a 40 caracteres: letras minúsculas, números e _, começando por letra.'
        );
        return;
      }
      return save.mutateAsync({
        ...(entityType ? {} : { code: value.code }),
        name: value.name.trim(),
        nature: value.nature,
        nonProfit: value.nonProfit,
        active: value.active,
        sortOrder: value.sortOrder
      });
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
        {!entityType && (
          <form.AppField
            name='code'
            children={(f) => (
              <f.TextField
                label='Código'
                required
                description='Identificador interno, sem acentos (ex.: coletivo_cultural). Não muda depois.'
              />
            )}
          />
        )}
        <form.AppField name='name' children={(f) => <f.TextField label='Nome' required />} />
        <form.AppField
          name='nature'
          children={(f) => (
            <f.SelectField
              label='Natureza'
              required
              description='Decide o checklist e as regras do instrumento.'
              options={Object.entries(natureLabels).map(([value, label]) => ({ value, label }))}
            />
          )}
        />
        <form.AppField
          name='nonProfit'
          children={(f) => (
            <f.SwitchField
              label='Sem fins lucrativos'
              description='Apenas entidades privadas sem fins lucrativos podem firmar Termo de Fomento.'
            />
          )}
        />
        <form.AppField
          name='sortOrder'
          children={(f) => (
            <f.TextField
              type='number'
              label='Ordem na lista'
              required
              description='Menor aparece primeiro.'
            />
          )}
        />
        <form.AppField
          name='active'
          children={(f) => (
            <f.SwitchField
              label='Ativo'
              description='Tipos inativos deixam de aparecer para novos projetos.'
            />
          )}
        />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>{entityType ? 'Salvar' : 'Criar tipo'}</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

/** Sem `entityType`, é o botão "Novo tipo"; com ele, o diálogo (controlado) de edição. */
export function EntityTypeDialog(
  props:
    | { entityType: EntityType; open: boolean; onOpenChange: (open: boolean) => void }
    | { entityType?: undefined; open?: undefined; onOpenChange?: undefined }
) {
  const can = useCan();
  const [open, setOpen] = useState(false);

  if (props.entityType) {
    return (
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Editar tipo de entidade</DialogTitle>
            <DialogDescription>{props.entityType.name}</DialogDescription>
          </DialogHeader>
          <EntityTypeForm entityType={props.entityType} onDone={() => props.onOpenChange(false)} />
        </DialogContent>
      </Dialog>
    );
  }
  if (!can(entityTypeActions.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo tipo
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo tipo de entidade</DialogTitle>
          <DialogDescription>
            Aparece na lista de tipos de quem cadastra um projeto.
          </DialogDescription>
        </DialogHeader>
        <EntityTypeForm onDone={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}
