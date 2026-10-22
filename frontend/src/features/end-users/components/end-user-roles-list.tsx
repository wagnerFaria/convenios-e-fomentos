'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { QueryError } from '@/components/query-error';
import { Badge } from '@/components/ui/badge';
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
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { endUserKeys, endUserRolesQueryOptions } from '../api/queries';
import { createEndUserRole, deleteEndUserRole } from '../api/service';
import { endUserActions, type EndUserRole } from '../api/types';

export const roleSchema = z.object({
  name: z.string().trim().min(1, 'Informe o nome.').max(100, 'Máximo de 100 caracteres.'),
  description: z.string().max(500, 'Máximo de 500 caracteres.')
});

export function NewEndUserRoleDialog() {
  const can = useCan();
  const router = useRouter();
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const create = useMutation({
    mutationFn: createEndUserRole,
    onSuccess: async (role) => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      setOpen(false);
      router.push(`/dashboard/portal/perfis/${role.id}`);
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

  if (!can(endUserActions.roles.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo perfil
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo perfil público</DialogTitle>
          <DialogDescription>
            Ele nasce sem permissões na API pública; você as define na próxima tela.
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

const roleTypeNote: Record<string, string> = {
  public: 'Visitantes sem login',
  authenticated: 'Usuários logados (padrão)'
};

export function EndUserRolesList() {
  const can = useCan();
  const qc = useQueryClient();
  const roles = useQuery(endUserRolesQueryOptions());
  const [toDelete, setToDelete] = useState<EndUserRole | null>(null);
  const remove = useMutation({
    mutationFn: deleteEndUserRole,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Perfil excluído. Os usuários dele passaram para o perfil Public.');
      setToDelete(null);
    },
    onError: (err) => {
      toast.error(errorMessage(err));
      setToDelete(null);
    }
  });

  if (roles.isError) return <QueryError error={roles.error} what='os perfis públicos' />;
  if (roles.isLoading) return <Skeleton className='h-40 w-full' />;

  return (
    <>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Perfil</TableHead>
            <TableHead>Descrição</TableHead>
            <TableHead>Usuários</TableHead>
            <TableHead className='w-px' />
          </TableRow>
        </TableHeader>
        <TableBody>
          {roles.data?.map((r) => (
            <TableRow key={r.id}>
              <TableCell className='font-medium'>
                {can(endUserActions.roles.read) ? (
                  <Link href={`/dashboard/portal/perfis/${r.id}`} className='hover:underline'>
                    {r.name}
                  </Link>
                ) : (
                  r.name
                )}{' '}
                {r.isSystem && <Badge variant='secondary'>Sistema</Badge>}
                {roleTypeNote[r.type] && (
                  <div className='text-muted-foreground text-xs font-normal'>
                    {roleTypeNote[r.type]}
                  </div>
                )}
              </TableCell>
              <TableCell className='text-muted-foreground max-w-md truncate'>
                {r.description || '—'}
              </TableCell>
              <TableCell className='tabular-nums'>{r.userCount}</TableCell>
              <TableCell>
                {can(endUserActions.roles.delete) && !r.isSystem && (
                  <Button size='sm' variant='ghost' onClick={() => setToDelete(r)}>
                    <Icons.trash aria-hidden /> Excluir
                  </Button>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <AlertModal
        isOpen={toDelete !== null}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && remove.mutate(toDelete.id)}
        loading={remove.isPending}
        title={`Excluir o perfil "${toDelete?.name ?? ''}"?`}
        description='Os usuários que o têm passam para o perfil Public. O perfil padrão das configurações não pode ser excluído.'
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </>
  );
}
