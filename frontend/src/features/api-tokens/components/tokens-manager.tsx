'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { QueryError } from '@/components/query-error';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
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
import { formatDateTime } from '@/lib/format';
import { tokenKeys, tokensQueryOptions } from '../api/queries';
import { regenerateToken, revokeToken } from '../api/service';
import {
  lifespanLabels,
  tokenActions,
  tokenStatusLabels,
  tokenTypeLabels,
  type ApiToken,
  type CreatedToken
} from '../api/types';
import { TokenCreatedDialog } from './token-created-dialog';
import { TokenFormDialog } from './token-form-dialog';

type Confirm = { kind: 'regenerate' | 'revoke'; token: ApiToken } | null;

const copy = {
  regenerate: {
    title: 'Regenerar este token?',
    description:
      'Um novo valor é gerado e o atual deixa de funcionar na hora. Integrações que o usam vão parar até serem atualizadas.',
    label: 'Regenerar'
  },
  revoke: {
    title: 'Revogar este token?',
    description: 'Ele deixa de funcionar imediatamente e não pode ser reativado.',
    label: 'Revogar'
  }
} as const;

export function TokensManager() {
  const can = useCan();
  const qc = useQueryClient();
  const [showRevoked, setShowRevoked] = useState(false);
  const tokens = useQuery(tokensQueryOptions(showRevoked));
  const [created, setCreated] = useState<CreatedToken | null>(null);
  const [editing, setEditing] = useState<ApiToken | null>(null);
  const [confirm, setConfirm] = useState<Confirm>(null);

  const act = useMutation({
    mutationFn: async (c: NonNullable<Confirm>) =>
      c.kind === 'regenerate' ? regenerateToken(c.token.id) : revokeToken(c.token.id),
    onSuccess: async (result, c) => {
      setConfirm(null);
      await qc.invalidateQueries({ queryKey: tokenKeys.all });
      if (c.kind === 'regenerate' && result) setCreated(result);
      else toast.success('Token revogado.');
    },
    onError: (err) => {
      setConfirm(null);
      toast.error(errorMessage(err));
    }
  });

  return (
    <div className='grid gap-4'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-2'>
          <Switch id='show-revoked' checked={showRevoked} onCheckedChange={setShowRevoked} />
          <Label htmlFor='show-revoked'>Mostrar revogados</Label>
        </div>
        <TokenFormDialog onCreated={setCreated} />
      </div>

      {tokens.isError ? (
        <QueryError error={tokens.error} what='os tokens' />
      ) : tokens.isLoading ? (
        <Skeleton className='h-40 w-full' />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Nome</TableHead>
              <TableHead>Tipo</TableHead>
              <TableHead>Token</TableHead>
              <TableHead>Validade</TableHead>
              <TableHead>Último uso</TableHead>
              <TableHead>Situação</TableHead>
              <TableHead className='w-px' />
            </TableRow>
          </TableHeader>
          <TableBody>
            {tokens.data?.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className='text-muted-foreground text-center'>
                  Nenhum token cadastrado.
                </TableCell>
              </TableRow>
            )}
            {tokens.data?.map((t) => (
              <TableRow key={t.id}>
                <TableCell className='font-medium'>
                  {t.name}
                  {t.description && (
                    <div className='text-muted-foreground text-xs font-normal'>{t.description}</div>
                  )}
                </TableCell>
                <TableCell>
                  {tokenTypeLabels[t.type]}
                  {t.type === 'custom' && (
                    <span className='text-muted-foreground'> ({t.permissions.length})</span>
                  )}
                </TableCell>
                <TableCell>
                  <code className='text-xs'>{t.hint}</code>
                </TableCell>
                <TableCell>
                  {t.expiresAt ? formatDateTime(t.expiresAt) : lifespanLabels.unlimited}
                </TableCell>
                <TableCell>{formatDateTime(t.lastUsedAt) || 'Nunca'}</TableCell>
                <TableCell>
                  <Badge
                    variant={
                      t.status === 'active'
                        ? 'default'
                        : t.status === 'revoked'
                          ? 'destructive'
                          : 'secondary'
                    }
                  >
                    {tokenStatusLabels[t.status]}
                  </Badge>
                </TableCell>
                <TableCell>
                  {t.status !== 'revoked' && (
                    <div className='flex justify-end gap-1'>
                      {can(tokenActions.update) && (
                        <Button size='sm' variant='ghost' onClick={() => setEditing(t)}>
                          <Icons.edit aria-hidden /> Editar
                        </Button>
                      )}
                      {can(tokenActions.regenerate) && (
                        <Button
                          size='sm'
                          variant='ghost'
                          onClick={() => setConfirm({ kind: 'regenerate', token: t })}
                        >
                          <Icons.refresh aria-hidden /> Regenerar
                        </Button>
                      )}
                      {can(tokenActions.delete) && (
                        <Button
                          size='sm'
                          variant='ghost'
                          onClick={() => setConfirm({ kind: 'revoke', token: t })}
                        >
                          <Icons.ban aria-hidden /> Revogar
                        </Button>
                      )}
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      {editing && (
        <TokenFormDialog
          key={editing.id}
          token={editing}
          open
          onOpenChange={(open) => !open && setEditing(null)}
        />
      )}
      <TokenCreatedDialog created={created} onClose={() => setCreated(null)} />
      <AlertModal
        isOpen={confirm !== null}
        onClose={() => setConfirm(null)}
        onConfirm={() => confirm && act.mutate(confirm)}
        loading={act.isPending}
        title={confirm ? copy[confirm.kind].title : ''}
        description={confirm ? copy[confirm.kind].description : ''}
        confirmLabel={confirm ? copy[confirm.kind].label : ''}
        cancelLabel='Cancelar'
      />
    </div>
  );
}
