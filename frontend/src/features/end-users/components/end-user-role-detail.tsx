'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import { Icons } from '@/components/icons';
import { QueryError } from '@/components/query-error';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan } from '@/features/auth/session-provider';
import { ApiActionPicker } from '@/features/roles/components/api-action-picker';
import { ApiError } from '@/lib/api/errors';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { endUserKeys, endUserRoleQueryOptions } from '../api/queries';
import { updateEndUserRole } from '../api/service';
import { endUserActions, type EndUserRoleDetail } from '../api/types';
import { roleSchema } from './end-user-roles-list';

function RoleForm({ role, editable }: { role: EndUserRoleDetail; editable: boolean }) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (v: { name: string; description: string; permissions: string[] }) =>
      updateEndUserRole(role.id, v),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Perfil salvo.');
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: {
      name: role.name,
      description: role.description,
      permissions: role.permissions.map((p) => p.action)
    },
    validators: { onSubmit: roleSchema.extend({ permissions: z.array(z.string()) }) },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync({ ...value, name: value.name.trim() });
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
          name='name'
          children={(f) => <f.TextField label='Nome' required disabled={!editable} />}
        />
        <form.AppField
          name='description'
          children={(f) => <f.TextareaField label='Descrição' disabled={!editable} />}
        />
        <form.Field
          name='permissions'
          children={(field) => (
            <Field>
              <FieldLabel>Ações permitidas na API pública</FieldLabel>
              <ApiActionPicker
                value={field.state.value}
                onChange={(v) => field.handleChange(v)}
                disabled={!editable}
              />
            </Field>
          )}
        />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        {editable && (
          <form.AppForm>
            <form.SubmitButton>Salvar</form.SubmitButton>
          </form.AppForm>
        )}
      </FieldGroup>
    </form>
  );
}

export function EndUserRoleDetail({ id }: { id: string }) {
  const can = useCan();
  const role = useQuery(endUserRoleQueryOptions(id));

  if (role.isLoading) return <Skeleton className='h-40 w-full max-w-3xl' />;
  if (role.isError || !role.data) {
    return role.error instanceof ApiError && role.error.status === 404 ? (
      <Alert variant='destructive' className='max-w-3xl'>
        <AlertDescription>Perfil não encontrado.</AlertDescription>
      </Alert>
    ) : (
      <QueryError error={role.error} what='o perfil' />
    );
  }
  const r = role.data;
  return (
    <div className='grid max-w-3xl gap-4'>
      <div>
        <Link
          href='/dashboard/portal/perfis'
          className={buttonVariants({ variant: 'ghost', size: 'sm' })}
        >
          <Icons.chevronLeft aria-hidden /> Perfis públicos
        </Link>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className='flex flex-wrap items-center gap-2'>
            {r.name}
            {r.isSystem && <Badge variant='secondary'>Sistema</Badge>}
          </CardTitle>
          <CardDescription>
            {r.userCount === 1 ? '1 usuário tem' : `${r.userCount} usuários têm`} este perfil. As
            permissões valem para a API pública.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <RoleForm key={r.updatedAt} role={r} editable={can(endUserActions.roles.update)} />
        </CardContent>
      </Card>
    </div>
  );
}
