'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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
import { entityTypeKeys, entityTypesQueryOptions } from '../api/queries';
import { deleteEntityType } from '../api/service';
import { entityTypeActions, natureLabels, type EntityType } from '../api/types';
import { EntityTypeDialog } from './entity-type-dialog';

export function EntityTypesList() {
  const can = useCan();
  const qc = useQueryClient();
  const types = useQuery(entityTypesQueryOptions());
  const [editing, setEditing] = useState<EntityType | null>(null);
  const [toDelete, setToDelete] = useState<EntityType | null>(null);

  const remove = useMutation({
    mutationFn: deleteEntityType,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: entityTypeKeys.all });
      toast.success('Tipo de entidade excluído.');
      setToDelete(null);
    },
    onError: (err) => {
      toast.error(errorMessage(err));
      setToDelete(null);
    }
  });

  if (types.isError) return <QueryError error={types.error} what='os tipos de entidade' />;
  if (types.isLoading) return <Skeleton className='h-40 w-full' />;

  return (
    <>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Tipo</TableHead>
            <TableHead>Natureza</TableHead>
            <TableHead>Situação</TableHead>
            <TableHead className='w-px' />
          </TableRow>
        </TableHeader>
        <TableBody>
          {types.data?.map((t) => (
            <TableRow key={t.id} className={t.active ? undefined : 'text-muted-foreground'}>
              <TableCell>
                <div className='font-medium'>{t.name}</div>
                <div className='text-muted-foreground font-mono text-xs'>{t.code}</div>
              </TableCell>
              <TableCell>
                {natureLabels[t.nature]}
                {t.nonProfit && ', sem fins lucrativos'}
                {t.isOther && (
                  <span className='text-muted-foreground block text-xs'>
                    O proponente informa o tipo por extenso e a natureza.
                  </span>
                )}
              </TableCell>
              <TableCell>
                <div className='flex flex-wrap gap-1'>
                  <Badge variant={t.active ? 'secondary' : 'outline'}>
                    {t.active ? 'Ativo' : 'Inativo'}
                  </Badge>
                  {t.isSystem && <Badge variant='outline'>Do sistema</Badge>}
                </div>
              </TableCell>
              <TableCell>
                <div className='flex justify-end gap-1'>
                  {can(entityTypeActions.update) && (
                    <Button size='sm' variant='ghost' onClick={() => setEditing(t)}>
                      <Icons.edit aria-hidden /> Editar
                    </Button>
                  )}
                  {can(entityTypeActions.delete) && (
                    <Button
                      size='sm'
                      variant='ghost'
                      disabled={t.isSystem}
                      title={t.isSystem ? 'Tipo do sistema não pode ser excluído.' : undefined}
                      onClick={() => setToDelete(t)}
                    >
                      <Icons.trash aria-hidden /> Excluir
                    </Button>
                  )}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {editing && (
        <EntityTypeDialog
          key={editing.id}
          entityType={editing}
          open
          onOpenChange={(open) => !open && setEditing(null)}
        />
      )}
      <AlertModal
        isOpen={toDelete !== null}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && remove.mutate(toDelete.id)}
        loading={remove.isPending}
        title={`Excluir o tipo "${toDelete?.name ?? ''}"?`}
        description='Se houver projetos usando o tipo, a exclusão é recusada; desative-o em vez de excluir.'
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </>
  );
}
