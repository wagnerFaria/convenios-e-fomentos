'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { debounce, parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { Pager } from '@/components/pager';
import { QueryError } from '@/components/query-error';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
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
import { useDebounce } from '@/hooks/use-debounce';
import { errorMessage } from '@/lib/errors';
import { formatDateTime } from '@/lib/format';
import { endUserKeys, endUserRolesQueryOptions, endUsersQueryOptions } from '../api/queries';
import { confirmEndUser, deleteEndUser, setEndUserBlocked } from '../api/service';
import { endUserActions, type EndUser } from '../api/types';
import { ChangeRoleDialog } from './change-role-dialog';
import { EndUserFormDialog } from './end-user-form-dialog';

const ALL = 'all';
const confirmedItems = [
  { value: ALL, label: 'E-mail: todos' },
  { value: 'true', label: 'E-mail confirmado' },
  { value: 'false', label: 'E-mail não confirmado' }
];
const blockedItems = [
  { value: ALL, label: 'Acesso: todos' },
  { value: 'false', label: 'Liberados' },
  { value: 'true', label: 'Bloqueados' }
];

function FilterSelect({
  items,
  value,
  onChange,
  label
}: {
  items: { value: string; label: string }[];
  value: string;
  onChange: (v: string) => void;
  label: string;
}) {
  return (
    <Select
      items={items}
      value={value || ALL}
      onValueChange={(v) => onChange(v === ALL ? '' : (v ?? ''))}
    >
      <SelectTrigger aria-label={label}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {items.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}

export function EndUsersList() {
  const can = useCan();
  const qc = useQueryClient();
  const [f, setF] = useQueryStates(
    {
      q: parseAsString.withDefault(''),
      roleId: parseAsString.withDefault(''),
      confirmed: parseAsString.withDefault(''),
      blocked: parseAsString.withDefault(''),
      page: parseAsInteger.withDefault(1)
    },
    { clearOnDefault: true, limitUrlUpdates: debounce(300) }
  );
  const search = useDebounce(f.q, 300);
  const users = useQuery({
    ...endUsersQueryOptions({
      q: search,
      roleId: f.roleId,
      confirmed: f.confirmed,
      blocked: f.blocked,
      page: f.page
    }),
    placeholderData: (prev) => prev
  });
  const roles = useQuery(endUserRolesQueryOptions());
  const roleItems = [
    { value: ALL, label: 'Todos os perfis' },
    ...(roles.data ?? []).map((r) => ({ value: r.id, label: r.name }))
  ];

  const [editing, setEditing] = useState<EndUser | null>(null);
  const [changingRole, setChangingRole] = useState<EndUser | null>(null);
  const [toDelete, setToDelete] = useState<EndUser | null>(null);

  const act = useMutation({
    mutationFn: async (a: { kind: 'block' | 'unblock' | 'confirm'; user: EndUser }) => {
      if (a.kind === 'confirm') return confirmEndUser(a.user.id);
      return setEndUserBlocked(a.user.id, a.kind === 'block');
    },
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Usuário atualizado.');
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  const remove = useMutation({
    mutationFn: deleteEndUser,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Usuário excluído.');
      setToDelete(null);
    },
    onError: (err) => {
      toast.error(errorMessage(err));
      setToDelete(null);
    }
  });

  const canUpdate = can(endUserActions.users.update);
  const canDelete = can(endUserActions.users.delete);

  return (
    <div className='grid gap-4'>
      <div className='flex flex-wrap items-center gap-2'>
        <Input
          value={f.q}
          onChange={(e) => void setF({ q: e.target.value, page: 1 })}
          placeholder='Buscar por usuário ou e-mail…'
          aria-label='Buscar usuários finais'
          className='max-w-sm'
        />
        {(roles.data?.length ?? 0) > 0 && (
          <FilterSelect
            label='Filtrar por perfil'
            items={roleItems}
            value={f.roleId}
            onChange={(roleId) => void setF({ roleId, page: 1 })}
          />
        )}
        <FilterSelect
          label='Filtrar por confirmação'
          items={confirmedItems}
          value={f.confirmed}
          onChange={(confirmed) => void setF({ confirmed, page: 1 })}
        />
        <FilterSelect
          label='Filtrar por bloqueio'
          items={blockedItems}
          value={f.blocked}
          onChange={(blocked) => void setF({ blocked, page: 1 })}
        />
      </div>

      {users.isError ? (
        <QueryError error={users.error} what='os usuários finais' />
      ) : users.isLoading ? (
        <Skeleton className='h-40 w-full' />
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Usuário</TableHead>
                <TableHead>Perfil</TableHead>
                <TableHead>Situação</TableHead>
                <TableHead>Criado em</TableHead>
                <TableHead className='w-px' />
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.data?.data.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className='text-muted-foreground text-center'>
                    Nenhum usuário encontrado.
                  </TableCell>
                </TableRow>
              )}
              {users.data?.data.map((u) => (
                <TableRow key={u.id}>
                  <TableCell className='font-medium'>
                    {u.username}
                    <div className='text-muted-foreground text-xs font-normal'>{u.email}</div>
                  </TableCell>
                  <TableCell>
                    <Badge variant='secondary'>{u.role.name}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className='flex flex-wrap gap-1'>
                      {u.blocked && <Badge variant='destructive'>Bloqueado</Badge>}
                      {!u.confirmed && <Badge variant='outline'>E-mail não confirmado</Badge>}
                      {!u.blocked && u.confirmed && <Badge>Ativo</Badge>}
                    </div>
                  </TableCell>
                  <TableCell>{formatDateTime(u.createdAt)}</TableCell>
                  <TableCell>
                    {(canUpdate || canDelete) && (
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={
                            <Button
                              variant='ghost'
                              size='icon-sm'
                              aria-label={`Ações para ${u.username}`}
                            />
                          }
                        >
                          <Icons.ellipsis aria-hidden />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align='end' className='w-auto'>
                          {canUpdate && (
                            <>
                              <DropdownMenuItem onClick={() => setEditing(u)}>
                                Editar
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => setChangingRole(u)}>
                                Alterar perfil
                              </DropdownMenuItem>
                              {!u.confirmed && (
                                <DropdownMenuItem
                                  onClick={() => act.mutate({ kind: 'confirm', user: u })}
                                >
                                  Confirmar e-mail
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuItem
                                onClick={() =>
                                  act.mutate({ kind: u.blocked ? 'unblock' : 'block', user: u })
                                }
                              >
                                {u.blocked ? 'Desbloquear' : 'Bloquear'}
                              </DropdownMenuItem>
                            </>
                          )}
                          {canUpdate && canDelete && <DropdownMenuSeparator />}
                          {canDelete && (
                            <DropdownMenuItem variant='destructive' onClick={() => setToDelete(u)}>
                              Excluir
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pager meta={users.data?.meta} onPage={(page) => void setF({ page })} />
        </>
      )}

      {editing && (
        <EndUserFormDialog
          key={editing.id}
          user={editing}
          open
          onOpenChange={(open) => !open && setEditing(null)}
        />
      )}
      {changingRole && (
        <ChangeRoleDialog
          key={changingRole.id}
          user={changingRole}
          open
          onOpenChange={(open) => !open && setChangingRole(null)}
        />
      )}
      <AlertModal
        isOpen={toDelete !== null}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && remove.mutate(toDelete.id)}
        loading={remove.isPending}
        title={`Excluir ${toDelete?.username ?? 'o usuário'}?`}
        description='A conta e as sessões são removidas. Para impedir o acesso sem excluir, bloqueie o usuário.'
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </div>
  );
}
