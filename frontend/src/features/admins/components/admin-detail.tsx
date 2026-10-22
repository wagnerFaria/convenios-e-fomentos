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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button, buttonVariants } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FieldGroup } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan, useSession } from '@/features/auth/session-provider';
import { AuditHistory } from '@/features/audit-logs/components/audit-history';
import { auditActions } from '@/features/audit-logs/api/types';
import { PermissionSummary } from '@/features/roles/components/permission-summary';
import { ApiError } from '@/lib/api/errors';
import { errorMessage } from '@/lib/errors';
import { formatDateTime } from '@/lib/format';
import { useAppForm } from '@/lib/form';
import { adminKeys, adminPermissionsQueryOptions, adminQueryOptions } from '../api/queries';
import {
  deleteAdmins,
  setAdminActive,
  setAdminBlocked,
  setAdminRoles,
  updateAdmin
} from '../api/service';
import { adminActions, type AdminUser } from '../api/types';
import { adminDisplayName } from '../lib/display';
import { AdminStatusBadge } from './admin-status-badge';
import {
  fromDepartmentId,
  toDepartmentId,
  useDepartmentOptions,
  useRoleOptions
} from './use-options';

type Confirm = 'deactivate' | 'block' | 'delete' | null;

const confirmCopy = {
  deactivate: {
    title: 'Desativar este administrador?',
    description:
      'Ele perde o acesso na próxima requisição e mantém os perfis. Você pode reativá-lo depois.',
    label: 'Desativar'
  },
  block: {
    title: 'Bloquear este administrador?',
    description:
      'Ele perde o acesso na próxima requisição e mantém os perfis. Você pode desbloqueá-lo depois.',
    label: 'Bloquear'
  },
  delete: {
    title: 'Excluir este administrador?',
    description:
      'A exclusão não pode ser desfeita. Para suspender o acesso sem excluir, desative ou bloqueie.',
    label: 'Excluir'
  }
} as const;

const rolesSchema = z.object({
  roleIds: z.array(z.string()).min(1, 'Selecione ao menos um perfil.')
});

function RolesCard({ admin, editable }: { admin: AdminUser; editable: boolean }) {
  const qc = useQueryClient();
  const { options } = useRoleOptions();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (roleIds: string[]) => setAdminRoles(admin.id, roleIds),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success('Perfis atualizados.');
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: { roleIds: admin.roles.map((r) => r.id) },
    validators: { onSubmit: rolesSchema },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync(value.roleIds);
    }
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Perfis</CardTitle>
        <CardDescription>O acesso é a união das permissões dos perfis atribuídos.</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void form.handleSubmit();
          }}
        >
          <FieldGroup>
            <form.AppField
              name='roleIds'
              mode='array'
              children={(f) => (
                <f.CheckboxGroupField
                  label='Perfis atribuídos'
                  options={options.map((o) => ({ ...o, disabled: o.disabled || !editable }))}
                />
              )}
            />
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            {editable && (
              <form.AppForm>
                <form.SubmitButton>Salvar perfis</form.SubmitButton>
              </form.AppForm>
            )}
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  );
}

function DepartmentCard({ admin, editable }: { admin: AdminUser; editable: boolean }) {
  const qc = useQueryClient();
  const { options } = useDepartmentOptions();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (departmentId: string | null) => updateAdmin(admin.id, { departmentId }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success('Departamento atualizado.');
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: { departmentId: fromDepartmentId(admin.departmentId) },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync(toDepartmentId(value.departmentId));
    }
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Departamento</CardTitle>
        <CardDescription>
          Define o alcance das permissões com a condição &quot;mesmo departamento&quot;.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void form.handleSubmit();
          }}
        >
          <FieldGroup>
            <form.AppField
              name='departmentId'
              children={(f) => <f.SelectField label='Departamento' options={options} />}
            />
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            {editable && (
              <form.AppForm>
                <form.SubmitButton>Salvar departamento</form.SubmitButton>
              </form.AppForm>
            )}
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  );
}

export function AdminDetail({ id }: { id: string }) {
  const can = useCan();
  const session = useSession();
  const router = useRouter();
  const qc = useQueryClient();
  const admin = useQuery(adminQueryOptions(id));
  const perms = useQuery(adminPermissionsQueryOptions(id));
  const [confirm, setConfirm] = useState<Confirm>(null);

  const act = useMutation({
    mutationFn: async (kind: 'activate' | 'unblock' | NonNullable<Confirm>) => {
      switch (kind) {
        case 'activate':
          return setAdminActive(id, true);
        case 'deactivate':
          return setAdminActive(id, false);
        case 'block':
          return setAdminBlocked(id, true);
        case 'unblock':
          return setAdminBlocked(id, false);
        case 'delete':
          return deleteAdmins([id]);
      }
    },
    onSuccess: async (_, kind) => {
      setConfirm(null);
      if (kind === 'delete') {
        await qc.invalidateQueries({ queryKey: adminKeys.all });
        toast.success('Administrador excluído.');
        router.push('/dashboard/administracao/usuarios');
        return;
      }
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success('Situação atualizada.');
    },
    onError: (err) => {
      setConfirm(null);
      toast.error(errorMessage(err));
    }
  });

  if (admin.isLoading) return <Skeleton className='h-40 w-full max-w-3xl' />;
  if (admin.isError || !admin.data) {
    const notFound = admin.error instanceof ApiError && admin.error.status === 404;
    return notFound ? (
      <Alert variant='destructive' className='max-w-3xl'>
        <AlertDescription>Administrador não encontrado.</AlertDescription>
      </Alert>
    ) : (
      <QueryError error={admin.error} what='o administrador' />
    );
  }

  const a = admin.data;
  const isSelf = a.id === session.id;
  const canUpdate = can(adminActions.update);
  // Quem não é Super Admin não altera a própria conta (o backend também recusa).
  const editable = canUpdate && (!isSelf || session.isSuperAdmin);

  return (
    <div className='grid max-w-3xl gap-4'>
      <div>
        <Link
          href='/dashboard/administracao/usuarios'
          className={buttonVariants({ variant: 'ghost', size: 'sm' })}
        >
          <Icons.chevronLeft aria-hidden /> Usuários
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className='flex flex-wrap items-center gap-2'>
            {adminDisplayName(a)}
            {isSelf && <span className='text-muted-foreground text-sm font-normal'>(você)</span>}
            <AdminStatusBadge status={a.status} />
          </CardTitle>
          <CardDescription>{a.email || 'E-mail definido no primeiro acesso'}</CardDescription>
        </CardHeader>
        <CardContent className='grid gap-4'>
          <dl className='grid grid-cols-2 gap-3 text-sm sm:grid-cols-3'>
            <div>
              <dt className='text-muted-foreground'>CPF</dt>
              <dd className='tabular-nums'>{a.cpf}</dd>
            </div>
            <div>
              <dt className='text-muted-foreground'>Último acesso</dt>
              <dd>{formatDateTime(a.lastLoginAt) || 'Nunca'}</dd>
            </div>
            <div>
              <dt className='text-muted-foreground'>Cadastrado em</dt>
              <dd>{formatDateTime(a.createdAt)}</dd>
            </div>
          </dl>
          {canUpdate && (
            <div className='flex flex-wrap items-center gap-2'>
              {a.isActive ? (
                <Button
                  variant='outline'
                  disabled={isSelf}
                  onClick={() => setConfirm('deactivate')}
                >
                  <Icons.circleX aria-hidden /> Desativar
                </Button>
              ) : (
                <Button
                  variant='outline'
                  disabled={act.isPending}
                  onClick={() => act.mutate('activate')}
                >
                  <Icons.check aria-hidden /> Ativar
                </Button>
              )}
              {a.blocked ? (
                <Button
                  variant='outline'
                  disabled={act.isPending}
                  onClick={() => act.mutate('unblock')}
                >
                  <Icons.lockOpen aria-hidden /> Desbloquear
                </Button>
              ) : (
                <Button variant='outline' disabled={isSelf} onClick={() => setConfirm('block')}>
                  <Icons.ban aria-hidden /> Bloquear
                </Button>
              )}
              {can(adminActions.delete) && (
                <Button
                  variant='destructive'
                  disabled={isSelf}
                  onClick={() => setConfirm('delete')}
                >
                  <Icons.trash aria-hidden /> Excluir
                </Button>
              )}
            </div>
          )}
          {isSelf && (
            <p className='text-muted-foreground text-sm'>
              Esta é a sua conta: você não pode desativá-la, bloqueá-la ou excluí-la.
              {!session.isSuperAdmin && ' Também não pode alterar seus próprios perfis.'}
            </p>
          )}
        </CardContent>
      </Card>

      <RolesCard key={`r-${a.updatedAt}-${a.roles.length}`} admin={a} editable={editable} />
      <DepartmentCard key={`d-${a.updatedAt}`} admin={a} editable={editable} />

      <Card>
        <CardHeader>
          <CardTitle>Permissões efetivas</CardTitle>
          <CardDescription>
            O que este administrador consegue fazer hoje, somando todos os perfis.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {perms.isLoading ? (
            <Skeleton className='h-24 w-full' />
          ) : perms.isError || !perms.data ? (
            <QueryError error={perms.error} what='as permissões' />
          ) : (
            <PermissionSummary
              permissions={perms.data.permissions}
              isSuperAdmin={perms.data.isSuperAdmin}
            />
          )}
        </CardContent>
      </Card>

      {can(auditActions.read) && (
        <Card>
          <CardHeader>
            <CardTitle>Histórico</CardTitle>
            <CardDescription>Últimas alterações nesta conta.</CardDescription>
          </CardHeader>
          <CardContent>
            <AuditHistory entity='admin' entityId={a.id} />
          </CardContent>
        </Card>
      )}

      <AlertModal
        isOpen={confirm !== null}
        onClose={() => setConfirm(null)}
        onConfirm={() => confirm && act.mutate(confirm)}
        loading={act.isPending}
        title={confirm ? confirmCopy[confirm].title : ''}
        description={confirm ? confirmCopy[confirm].description : ''}
        confirmLabel={confirm ? confirmCopy[confirm].label : ''}
        cancelLabel='Cancelar'
      />
    </div>
  );
}
