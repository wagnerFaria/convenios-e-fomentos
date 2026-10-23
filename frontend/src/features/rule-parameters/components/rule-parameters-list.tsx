'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { QueryError } from '@/components/query-error';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import {
  ruleParameterKeys,
  ruleParameterKeysQueryOptions,
  ruleParametersQueryOptions
} from '../api/queries';
import { deleteRuleParameter } from '../api/service';
import { ruleParameterActions, type RuleParameter, type RuleParameterKey } from '../api/types';
import { formatDate, summarize, todayInCuiaba } from '../format';
import { RuleParameterDialog } from './rule-parameter-dialog';

export function RuleParametersList() {
  const can = useCan();
  const qc = useQueryClient();
  const keys = useQuery(ruleParameterKeysQueryOptions());
  const params = useQuery(ruleParametersQueryOptions());
  const [editing, setEditing] = useState<RuleParameter | null>(null);
  const [toDelete, setToDelete] = useState<RuleParameter | null>(null);

  const remove = useMutation({
    mutationFn: deleteRuleParameter,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ruleParameterKeys.all });
      toast.success('Versão excluída.');
      setToDelete(null);
    },
    onError: (err) => {
      toast.error(errorMessage(err));
      setToDelete(null);
    }
  });

  if (keys.isError) return <QueryError error={keys.error} what='as chaves de parâmetros' />;
  if (params.isError) return <QueryError error={params.error} what='os parâmetros' />;
  if (keys.isLoading || params.isLoading) return <Skeleton className='h-60 w-full' />;

  const today = todayInCuiaba();
  const editingKey = editing ? keys.data?.find((k) => k.key === editing.key) : undefined;

  return (
    <>
      <div className='flex flex-col gap-4'>
        {keys.data?.map((k: RuleParameterKey) => {
          const versions = (params.data ?? []).filter((p) => p.key === k.key);
          return (
            <Card key={k.key}>
              <CardHeader>
                <div className='flex flex-wrap items-start justify-between gap-2'>
                  <div>
                    <CardTitle className='font-mono text-base'>{k.key}</CardTitle>
                    <CardDescription>{k.description}</CardDescription>
                    <p className='text-muted-foreground mt-1 text-xs'>Origem: {k.source}</p>
                  </div>
                  <RuleParameterDialog paramKey={k} />
                </div>
              </CardHeader>
              <CardContent>
                {versions.length === 0 && (
                  <p className='text-muted-foreground text-sm'>Nenhuma versão cadastrada.</p>
                )}
                <ul className='divide-y'>
                  {versions.map((p) => {
                    const current =
                      p.validFrom <= today && (p.validTo === null || p.validTo >= today);
                    return (
                      <li
                        key={p.id}
                        className='flex flex-wrap items-start justify-between gap-2 py-3'
                      >
                        <div className='space-y-1'>
                          <div className='flex flex-wrap items-center gap-2 text-sm font-medium'>
                            {formatDate(p.validFrom)} até{' '}
                            {p.validTo ? formatDate(p.validTo) : 'sem fim'}
                            {current && <Badge>Vigente hoje</Badge>}
                          </div>
                          {summarize(k.key, p.value).map((line) => (
                            <div key={line} className='text-sm'>
                              {line}
                            </div>
                          ))}
                          {p.note && <div className='text-muted-foreground text-xs'>{p.note}</div>}
                        </div>
                        <div className='flex gap-1'>
                          {can(ruleParameterActions.update) && (
                            <Button size='sm' variant='ghost' onClick={() => setEditing(p)}>
                              <Icons.edit aria-hidden /> Editar
                            </Button>
                          )}
                          {can(ruleParameterActions.delete) && (
                            <Button size='sm' variant='ghost' onClick={() => setToDelete(p)}>
                              <Icons.trash aria-hidden /> Excluir
                            </Button>
                          )}
                        </div>
                      </li>
                    );
                  })}
                </ul>
              </CardContent>
            </Card>
          );
        })}
      </div>
      {editing && editingKey && (
        <RuleParameterDialog
          key={editing.id}
          paramKey={editingKey}
          parameter={editing}
          open
          onOpenChange={(open) => !open && setEditing(null)}
        />
      )}
      <AlertModal
        isOpen={toDelete !== null}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && remove.mutate(toDelete.id)}
        loading={remove.isPending}
        title='Excluir esta versão?'
        description='Datas sem versão vigente deixam a regra sem valor naquele período. A exclusão não pode ser desfeita.'
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </>
  );
}
