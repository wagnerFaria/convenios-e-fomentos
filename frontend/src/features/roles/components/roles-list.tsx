'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { QueryError } from '@/components/query-error';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
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
import { roleKeys, rolesQueryOptions } from '../api/queries';
import { deleteRole, duplicateRole } from '../api/service';
import { roleActions, type RoleSummary } from '../api/types';

export function RolesList() {
  const can = useCan();
  const router = useRouter();
  const qc = useQueryClient();
  const roles = useQuery(rolesQueryOptions());
  const [toDelete, setToDelete] = useState<RoleSummary | null>(null);

  const duplicate = useMutation({
    mutationFn: duplicateRole,
    onSuccess: async (role) => {
      await qc.invalidateQueries({ queryKey: roleKeys.all });
      toast.success(`Perfil "${role.name}" criado.`);
      router.push(`/dashboard/administracao/perfis/${role.id}`);
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  const remove = useMutation({
    mutationFn: deleteRole,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: roleKeys.all });
      toast.success('Perfil excluído.');
      setToDelete(null);
    },
    onError: (err) => {
      toast.error(errorMessage(err));
      setToDelete(null);
    }
  });

  if (roles.isError) return <QueryError error={roles.error} what='os perfis' />;
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
          {roles.data?.length === 0 && (
            <TableRow>
              <TableCell colSpan={4} className='text-muted-foreground text-center'>
                Nenhum perfil cadastrado.
              </TableCell>
            </TableRow>
          )}
          {roles.data?.map((r) => (
            <TableRow key={r.id}>
              <TableCell className='font-medium'>
                {can(roleActions.read) ? (
                  <Link
                    href={`/dashboard/administracao/perfis/${r.id}`}
                    className='hover:underline'
                  >
                    {r.name}
                  </Link>
                ) : (
                  r.name
                )}{' '}
                {r.isSuperAdmin ? (
                  <Badge>Acesso total</Badge>
                ) : (
                  r.isSystem && <Badge variant='secondary'>Sistema</Badge>
                )}
              </TableCell>
              <TableCell className='text-muted-foreground max-w-md truncate'>
                {r.description || '—'}
              </TableCell>
              <TableCell>
                <Link
                  href={`/dashboard/administracao/usuarios?roleId=${r.id}`}
                  className='tabular-nums hover:underline'
                >
                  {r.userCount}
                </Link>
              </TableCell>
              <TableCell>
                <div className='flex justify-end gap-1'>
                  {can(roleActions.create) && !r.isSuperAdmin && (
                    <Button
                      size='sm'
                      variant='ghost'
                      disabled={duplicate.isPending}
                      onClick={() => duplicate.mutate(r.id)}
                    >
                      <Icons.copy aria-hidden /> Duplicar
                    </Button>
                  )}
                  {can(roleActions.delete) && !r.isSuperAdmin && (
                    <Button size='sm' variant='ghost' onClick={() => setToDelete(r)}>
                      <Icons.trash aria-hidden /> Excluir
                    </Button>
                  )}
                </div>
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
        description='O perfil e suas permissões serão removidos. Perfis ainda atribuídos a administradores não podem ser excluídos.'
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </>
  );
}
