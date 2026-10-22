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
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { useCan } from '@/features/auth/session-provider';
import { ApiActionPicker } from '@/features/roles/components/api-action-picker';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { tokenKeys } from '../api/queries';
import { createToken, updateToken } from '../api/service';
import {
  lifespanLabels,
  tokenActions,
  tokenTypeDescriptions,
  tokenTypeLabels,
  type ApiToken,
  type CreatedToken,
  type TokenLifespan,
  type TokenType
} from '../api/types';

const typeOptions = (Object.keys(tokenTypeLabels) as TokenType[]).map((value) => ({
  value,
  label: `${tokenTypeLabels[value]}: ${tokenTypeDescriptions[value]}`
}));
const lifespanOptions = (Object.keys(lifespanLabels) as TokenLifespan[]).map((value) => ({
  value,
  label: lifespanLabels[value]
}));

const schema = z
  .object({
    name: z.string().trim().min(1, 'Informe o nome.').max(100, 'Máximo de 100 caracteres.'),
    description: z.string().max(500, 'Máximo de 500 caracteres.'),
    type: z.enum(['read-only', 'full-access', 'custom']),
    lifespan: z.enum(['7d', '30d', '90d', 'unlimited']),
    permissions: z.array(z.string())
  })
  .refine((v) => v.type !== 'custom' || v.permissions.length > 0, {
    path: ['permissions'],
    message: 'Escolha ao menos uma ação.'
  });

const isCreated = (t: ApiToken | CreatedToken): t is CreatedToken => 'token' in t;

function TokenForm({
  token,
  onDone,
  onCreated
}: {
  token?: ApiToken;
  onDone: () => void;
  onCreated: (t: CreatedToken) => void;
}) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: async (v: z.infer<typeof schema>) => {
      const input = {
        name: v.name.trim(),
        description: v.description,
        type: v.type,
        permissions: v.type === 'custom' ? v.permissions : []
      };
      return token ? updateToken(token.id, input) : createToken({ ...input, lifespan: v.lifespan });
    },
    onSuccess: async (saved) => {
      await qc.invalidateQueries({ queryKey: tokenKeys.all });
      if (isCreated(saved)) onCreated(saved);
      else toast.success('Token atualizado.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });

  const form = useAppForm({
    defaultValues: {
      name: token?.name ?? '',
      description: token?.description ?? '',
      type: (token?.type ?? 'read-only') as TokenType,
      lifespan: (token?.lifespan ?? '30d') as TokenLifespan,
      permissions: token?.permissions ?? ([] as string[])
    },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync(value);
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
        <form.AppField name='description' children={(f) => <f.TextareaField label='Descrição' />} />
        <form.AppField
          name='type'
          children={(f) => <f.RadioGroupField label='Tipo de acesso' options={typeOptions} />}
        />
        <form.Subscribe selector={(s) => s.values.type}>
          {(type) =>
            type === 'custom' && (
              <form.Field
                name='permissions'
                children={(field) => {
                  const invalid = field.state.meta.isTouched && !field.state.meta.isValid;
                  return (
                    <Field data-invalid={invalid}>
                      <FieldLabel>Ações permitidas *</FieldLabel>
                      <ApiActionPicker
                        value={field.state.value}
                        onChange={(v) => field.handleChange(v)}
                      />
                      {invalid && <FieldError errors={field.state.meta.errors} />}
                    </Field>
                  );
                }}
              />
            )
          }
        </form.Subscribe>
        {!token && (
          <form.AppField
            name='lifespan'
            children={(f) => (
              <f.SelectField
                label='Validade'
                description='Depois de criado, o prazo não muda; regenere o token para renová-lo.'
                options={lifespanOptions}
              />
            )}
          />
        )}
        {token && <FieldDescription>A validade não pode ser alterada.</FieldDescription>}
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>{token ? 'Salvar' : 'Criar token'}</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

/** Sem `token`, é o botão "Novo token"; com ele, o diálogo (controlado) de edição. */
export function TokenFormDialog(
  props:
    | {
        token: ApiToken;
        open: boolean;
        onOpenChange: (open: boolean) => void;
        onCreated?: undefined;
      }
    | {
        token?: undefined;
        open?: undefined;
        onOpenChange?: undefined;
        onCreated: (t: CreatedToken) => void;
      }
) {
  const can = useCan();
  const [open, setOpen] = useState(false);

  if (props.token) {
    return (
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent className='max-h-[90vh] overflow-y-auto'>
          <DialogHeader>
            <DialogTitle>Editar token</DialogTitle>
            <DialogDescription>
              Alterações valem na próxima requisição com o token.
            </DialogDescription>
          </DialogHeader>
          <TokenForm
            token={props.token}
            onDone={() => props.onOpenChange(false)}
            onCreated={() => {}}
          />
        </DialogContent>
      </Dialog>
    );
  }
  if (!can(tokenActions.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo token
      </DialogTrigger>
      <DialogContent className='max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle>Novo token de API</DialogTitle>
          <DialogDescription>
            Para integrações sistema a sistema com a API pública. O valor aparece uma única vez.
          </DialogDescription>
        </DialogHeader>
        <TokenForm onDone={() => setOpen(false)} onCreated={props.onCreated} />
      </DialogContent>
    </Dialog>
  );
}
