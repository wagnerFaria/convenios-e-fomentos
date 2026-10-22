'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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
import { endUserKeys, endUserRolesQueryOptions } from '../api/queries';
import { createEndUser, updateEndUser } from '../api/service';
import { endUserActions, type EndUser } from '../api/types';

const DEFAULT_ROLE = 'default';

const baseShape = {
  username: z.string().trim().min(3, 'Mínimo de 3 caracteres.').max(50, 'Máximo de 50 caracteres.'),
  email: z.string().trim().email('E-mail inválido.').max(255, 'Máximo de 255 caracteres.')
};
const createSchema = z.object({
  ...baseShape,
  password: z.string().min(1, 'Informe a senha.'),
  roleId: z.string(),
  confirmed: z.boolean(),
  blocked: z.boolean()
});
const editSchema = z.object({ ...baseShape, password: z.string() });

function CreateForm({ onDone }: { onDone: () => void }) {
  const qc = useQueryClient();
  const roles = useQuery(endUserRolesQueryOptions());
  const [error, setError] = useState<string | null>(null);
  const create = useMutation({
    mutationFn: createEndUser,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Usuário criado.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: {
      username: '',
      email: '',
      password: '',
      roleId: DEFAULT_ROLE,
      confirmed: true,
      blocked: false
    },
    validators: { onSubmit: createSchema },
    onSubmit: ({ value }) => {
      setError(null);
      return create.mutateAsync({
        ...value,
        roleId: value.roleId === DEFAULT_ROLE ? undefined : value.roleId
      });
    }
  });
  const roleOptions = [
    { value: DEFAULT_ROLE, label: 'Perfil padrão das configurações' },
    ...(roles.data ?? []).map((r) => ({ value: r.id, label: r.name }))
  ];

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.AppField
          name='username'
          children={(f) => <f.TextField label='Nome de usuário' required autoComplete='off' />}
        />
        <form.AppField
          name='email'
          children={(f) => <f.TextField label='E-mail' type='email' required autoComplete='off' />}
        />
        <form.AppField
          name='password'
          children={(f) => (
            <f.TextField label='Senha' type='password' required autoComplete='new-password' />
          )}
        />
        <form.AppField
          name='roleId'
          children={(f) => <f.SelectField label='Perfil' options={roleOptions} />}
        />
        <form.AppField
          name='confirmed'
          children={(f) => (
            <f.SwitchField
              label='E-mail confirmado'
              description='Dispensa a confirmação por e-mail.'
            />
          )}
        />
        <form.AppField
          name='blocked'
          children={(f) => <f.SwitchField label='Bloqueado' description='Não consegue entrar.' />}
        />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>Criar usuário</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

function EditForm({ user, onDone }: { user: EndUser; onDone: () => void }) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (v: { username: string; email: string; password: string }) =>
      updateEndUser(user.id, {
        username: v.username.trim(),
        email: v.email.trim(),
        password: v.password || undefined
      }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Usuário atualizado.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: { username: user.username, email: user.email, password: '' },
    validators: { onSubmit: editSchema },
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
        <form.AppField
          name='username'
          children={(f) => <f.TextField label='Nome de usuário' required autoComplete='off' />}
        />
        <form.AppField
          name='email'
          children={(f) => <f.TextField label='E-mail' type='email' required autoComplete='off' />}
        />
        <form.AppField
          name='password'
          children={(f) => (
            <f.TextField
              label='Nova senha'
              type='password'
              autoComplete='new-password'
              description='Deixe em branco para manter. Ao trocar, as sessões ativas são encerradas.'
            />
          )}
        />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>Salvar</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

/** Sem `user`, é o botão "Novo usuário"; com ele, o diálogo (controlado) de edição. */
export function EndUserFormDialog(
  props:
    | { user: EndUser; open: boolean; onOpenChange: (open: boolean) => void }
    | { user?: undefined; open?: undefined; onOpenChange?: undefined }
) {
  const can = useCan();
  const [open, setOpen] = useState(false);

  if (props.user) {
    return (
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Editar usuário</DialogTitle>
            <DialogDescription>{props.user.email}</DialogDescription>
          </DialogHeader>
          <EditForm user={props.user} onDone={() => props.onOpenChange(false)} />
        </DialogContent>
      </Dialog>
    );
  }
  if (!can(endUserActions.users.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo usuário
      </DialogTrigger>
      <DialogContent className='max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle>Novo usuário final</DialogTitle>
          <DialogDescription>
            Conta do portal público, criada por um administrador.
          </DialogDescription>
        </DialogHeader>
        <CreateForm onDone={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}
