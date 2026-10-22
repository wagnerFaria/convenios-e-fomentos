'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { debounce, parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { Pager } from '@/components/pager';
import { QueryError } from '@/components/query-error';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
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
import { useCan, useSession } from '@/features/auth/session-provider';
import { useDebounce } from '@/hooks/use-debounce';
import { errorMessage } from '@/lib/errors';
import { formatDateTime } from '@/lib/format';
import { adminKeys, adminsQueryOptions } from '../api/queries';
import { deleteAdmins } from '../api/service';
import { adminActions, adminStatusLabels, type AdminStatus, type AdminUser } from '../api/types';
import { adminDisplayName } from '../lib/display';
import { AdminStatusBadge } from './admin-status-badge';
import { ReleaseAccessDialog } from './release-access-dialog';
import { useDepartmentOptions, useRoleOptions } from './use-options';

const ALL = 'all';
const statusItems = [
  { value: ALL, label: 'Todos os status' },
  ...(Object.keys(adminStatusLabels) as AdminStatus[]).map((s) => ({
    value: s,
    label: adminStatusLabels[s]
  }))
];

export function AdminsList() {
  const can = useCan();
  const session = useSession();
  const qc = useQueryClient();
  const [filters, setFilters] = useQueryStates(
    {
      q: parseAsString.withDefault(''),
      status: parseAsString.withDefault(''),
      roleId: parseAsString.withDefault(''),
      page: parseAsInteger.withDefault(1)
    },
    { clearOnDefault: true, limitUrlUpdates: debounce(300) }
  );
  const search = useDebounce(filters.q, 300);
  const params = {
    q: search,
    status: filters.status || undefined,
    roleId: filters.roleId || undefined,
    page: filters.page
  };
  const admins = useQuery({ ...adminsQueryOptions(params), placeholderData: (prev) => prev });
  const awaiting = useQuery({
    ...adminsQueryOptions({ status: 'awaiting', pageSize: 1 }),
    enabled: can(adminActions.read)
  });
  const { roles, options: roleOptions } = useRoleOptions();
  const { departments } = useDepartmentOptions();
  const deptName = (id: string | null) => departments.find((d) => d.id === id)?.name ?? '—';

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [releasing, setReleasing] = useState<AdminUser | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const remove = useMutation({
    mutationFn: deleteAdmins,
    onSuccess: async (_, ids) => {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success(
        ids.length === 1 ? 'Administrador excluído.' : `${ids.length} administradores excluídos.`
      );
      setSelected(new Set());
      setConfirmDelete(false);
    },
    onError: (err) => {
      toast.error(errorMessage(err));
      setConfirmDelete(false);
    }
  });

  const rows = admins.data?.data ?? [];
  const selectable = rows.filter((a) => a.id !== session.id);
  const allSelected = selectable.length > 0 && selectable.every((a) => selected.has(a.id));
  const toggle = (id: string, on: boolean) =>
    setSelected((cur) => {
      const next = new Set(cur);
      if (on) next.add(id);
      else next.delete(id);
      return next;
    });
  const awaitingCount = awaiting.data?.meta?.total ?? 0;
  const roleItems = [
    { value: ALL, label: 'Todos os perfis' },
    ...roleOptions.map((o) => ({ value: o.value, label: o.label }))
  ];

  return (
    <div className='grid gap-4'>
      {awaitingCount > 0 && filters.status !== 'awaiting' && (
        <Alert>
          <Icons.userCheck aria-hidden />
          <AlertDescription className='flex flex-wrap items-center justify-between gap-2'>
            <span>
              {awaitingCount === 1
                ? '1 pessoa entrou e aguarda liberação de acesso.'
                : `${awaitingCount} pessoas entraram e aguardam liberação de acesso.`}
            </span>
            <Button
              size='sm'
              variant='outline'
              onClick={() => void setFilters({ status: 'awaiting', page: 1 })}
            >
              Ver quem aguarda
            </Button>
          </AlertDescription>
        </Alert>
      )}

      <div className='flex flex-wrap items-center gap-2'>
        <Input
          value={filters.q}
          onChange={(e) => void setFilters({ q: e.target.value, page: 1 })}
          placeholder='Buscar por nome, e-mail ou CPF…'
          aria-label='Buscar administradores'
          className='max-w-sm'
        />
        <Select
          items={statusItems}
          value={filters.status || ALL}
          onValueChange={(v) => void setFilters({ status: v === ALL ? '' : (v ?? ''), page: 1 })}
        >
          <SelectTrigger aria-label='Filtrar por situação'>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              {statusItems.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
        {roles.length > 0 && (
          <Select
            items={roleItems}
            value={filters.roleId || ALL}
            onValueChange={(v) => void setFilters({ roleId: v === ALL ? '' : (v ?? ''), page: 1 })}
          >
            <SelectTrigger aria-label='Filtrar por perfil'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {roleItems.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        )}
        {can(adminActions.delete) && selected.size > 0 && (
          <Button variant='destructive' onClick={() => setConfirmDelete(true)}>
            <Icons.trash aria-hidden /> Excluir selecionados ({selected.size})
          </Button>
        )}
      </div>

      {admins.isError ? (
        <QueryError error={admins.error} what='os administradores' />
      ) : admins.isLoading ? (
        <Skeleton className='h-40 w-full' />
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                {can(adminActions.delete) && (
                  <TableHead className='w-8'>
                    <Checkbox
                      aria-label='Selecionar todos'
                      checked={allSelected}
                      onCheckedChange={(on) =>
                        setSelected(on ? new Set(selectable.map((a) => a.id)) : new Set())
                      }
                    />
                  </TableHead>
                )}
                <TableHead>Nome</TableHead>
                <TableHead>CPF</TableHead>
                <TableHead>Perfis</TableHead>
                <TableHead>Departamento</TableHead>
                <TableHead>Situação</TableHead>
                <TableHead>Último acesso</TableHead>
                <TableHead className='w-px' />
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8} className='text-muted-foreground text-center'>
                    Nenhum administrador encontrado.
                  </TableCell>
                </TableRow>
              )}
              {rows.map((a) => (
                <TableRow key={a.id}>
                  {can(adminActions.delete) && (
                    <TableCell>
                      <Checkbox
                        aria-label={`Selecionar ${adminDisplayName(a)}`}
                        disabled={a.id === session.id}
                        checked={selected.has(a.id)}
                        onCheckedChange={(on) => toggle(a.id, !!on)}
                      />
                    </TableCell>
                  )}
                  <TableCell className='font-medium'>
                    <Link
                      href={`/dashboard/administracao/usuarios/${a.id}`}
                      className='hover:underline'
                    >
                      {adminDisplayName(a)}
                    </Link>
                    {a.id === session.id && <span className='text-muted-foreground'> (você)</span>}
                    <div className='text-muted-foreground text-xs font-normal'>
                      {a.email || '—'}
                    </div>
                  </TableCell>
                  <TableCell className='tabular-nums'>{a.cpf}</TableCell>
                  <TableCell>
                    <div className='flex flex-wrap gap-1'>
                      {a.roles.length === 0 && <span className='text-muted-foreground'>—</span>}
                      {a.roles.map((r) => (
                        <Badge key={r.id} variant='secondary'>
                          {r.name}
                        </Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell>{deptName(a.departmentId)}</TableCell>
                  <TableCell>
                    <AdminStatusBadge status={a.status} />
                  </TableCell>
                  <TableCell>{formatDateTime(a.lastLoginAt) || 'Nunca'}</TableCell>
                  <TableCell>
                    {a.status === 'awaiting' && can(adminActions.update) && (
                      <Button size='sm' onClick={() => setReleasing(a)}>
                        Liberar acesso
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pager meta={admins.data?.meta} onPage={(page) => void setFilters({ page })} />
        </>
      )}

      {releasing && (
        <ReleaseAccessDialog
          key={releasing.id}
          admin={releasing}
          open
          onOpenChange={(open) => !open && setReleasing(null)}
        />
      )}
      <AlertModal
        isOpen={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        onConfirm={() => remove.mutate([...selected])}
        loading={remove.isPending}
        title={`Excluir ${selected.size} administrador(es)?`}
        description='Eles perdem o acesso na hora. A exclusão não pode ser desfeita; para suspender o acesso sem excluir, desative ou bloqueie.'
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </div>
  );
}
