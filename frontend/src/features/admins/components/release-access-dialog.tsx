'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { FieldGroup } from '@/components/ui/field';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { adminKeys } from '../api/queries';
import { setAdminRoles, updateAdmin } from '../api/service';
import type { AdminUser } from '../api/types';
import { adminDisplayName } from '../lib/display';
import {
  fromDepartmentId,
  NO_DEPARTMENT,
  toDepartmentId,
  useDepartmentOptions,
  useRoleOptions
} from './use-options';
import { REVIEWER_ROLE_CODE, ReviewerDepartmentWarning } from './reviewer-warning';

const schema = z.object({
  roleIds: z.array(z.string()).min(1, 'Selecione ao menos um perfil.'),
  departmentId: z.string()
});

/** Libera o acesso de quem entrou pelo MT Login e aguarda um perfil: define perfis e departamento. */
export function ReleaseAccessDialog({
  admin,
  open,
  onOpenChange
}: {
  admin: AdminUser;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const { options: roleOptions, roles } = useRoleOptions();
  const reviewerRoleIds = new Set(
    roles.filter((r) => r.code === REVIEWER_ROLE_CODE).map((r) => r.id)
  );
  const { options: deptOptions } = useDepartmentOptions();

  const release = useMutation({
    mutationFn: async (v: { roleIds: string[]; departmentId: string | null }) => {
      await setAdminRoles(admin.id, v.roleIds);
      if (v.departmentId !== admin.departmentId) {
        await updateAdmin(admin.id, { departmentId: v.departmentId });
      }
    },
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success(`Acesso de ${adminDisplayName(admin)} liberado.`);
      onOpenChange(false);
    },
    onError: (err) => setError(errorMessage(err))
  });

  const form = useAppForm({
    defaultValues: { roleIds: [] as string[], departmentId: fromDepartmentId(admin.departmentId) },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return release.mutateAsync({
        roleIds: value.roleIds,
        departmentId: toDepartmentId(value.departmentId)
      });
    }
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Liberar acesso</DialogTitle>
          <DialogDescription>
            {adminDisplayName(admin)} ({admin.email}) já entrou, mas ainda não tem perfil. Escolha o
            que ele poderá fazer.
          </DialogDescription>
        </DialogHeader>
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
                <f.CheckboxGroupField label='Perfis' required options={roleOptions} />
              )}
            />
            <form.AppField
              name='departmentId'
              children={(f) => (
                <f.SelectField
                  label='Departamento'
                  placeholder='Sem departamento'
                  options={deptOptions}
                />
              )}
            />
            <form.Subscribe selector={(s) => s.values}>
              {(values) => (
                <ReviewerDepartmentWarning
                  show={
                    values.roleIds.some((id) => reviewerRoleIds.has(id)) &&
                    values.departmentId === NO_DEPARTMENT
                  }
                />
              )}
            </form.Subscribe>
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            <form.AppForm>
              <form.SubmitButton>Liberar acesso</form.SubmitButton>
            </form.AppForm>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
