'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { useCan } from '@/features/auth/session-provider';
import { formatCpf, isValidCpf } from '@/lib/cpf';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { adminKeys } from '../api/queries';
import { registerAdmin } from '../api/service';
import { adminActions } from '../api/types';
import { toDepartmentId, useDepartmentOptions, useRoleOptions, NO_DEPARTMENT } from './use-options';

const schema = z.object({
  cpf: z.string().refine(isValidCpf, 'CPF inválido.'),
  roleIds: z.array(z.string()).min(1, 'Selecione ao menos um perfil.'),
  departmentId: z.string()
});

/** Pré-cadastro pelo CPF: nome e e-mail vêm do MT Login quando a pessoa entrar pela primeira vez. */
export function NewAdminDialog() {
  const can = useCan();
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const qc = useQueryClient();
  const { options: roleOptions, isLoading: loadingRoles } = useRoleOptions();
  const { options: deptOptions } = useDepartmentOptions();

  const register = useMutation({
    mutationFn: registerAdmin,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success('Administrador pré-cadastrado. O acesso é liberado no primeiro login.');
      setOpen(false);
    },
    onError: (err) => setError(errorMessage(err))
  });

  const form = useAppForm({
    defaultValues: { cpf: '', roleIds: [] as string[], departmentId: NO_DEPARTMENT },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return register.mutateAsync({
        cpf: value.cpf,
        roleIds: value.roleIds,
        departmentId: toDepartmentId(value.departmentId)
      });
    }
  });

  if (!can(adminActions.create)) return null;

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) {
          form.reset();
          setError(null);
        }
      }}
    >
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Pré-cadastrar
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Pré-cadastrar administrador</DialogTitle>
          <DialogDescription>
            Informe o CPF e os perfis. Nome e e-mail vêm do MT Login no primeiro acesso.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void form.handleSubmit();
          }}
        >
          <FieldGroup>
            <form.Field
              name='cpf'
              children={(field) => {
                const invalid = field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={invalid}>
                    <FieldLabel htmlFor='cpf'>CPF *</FieldLabel>
                    <Input
                      id='cpf'
                      name='cpf'
                      inputMode='numeric'
                      autoComplete='off'
                      placeholder='000.000.000-00'
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(formatCpf(e.target.value))}
                      aria-invalid={invalid}
                    />
                    <FieldDescription>A pessoa será identificada por este CPF.</FieldDescription>
                    {invalid && <FieldError errors={field.state.meta.errors} />}
                  </Field>
                );
              }}
            />
            <form.AppField
              name='roleIds'
              mode='array'
              children={(f) => (
                <f.CheckboxGroupField
                  label='Perfis'
                  required
                  description={
                    loadingRoles
                      ? 'Carregando perfis…'
                      : 'O acesso é a união das permissões dos perfis.'
                  }
                  options={roleOptions}
                />
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
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            <form.AppForm>
              <form.SubmitButton>Pré-cadastrar</form.SubmitButton>
            </form.AppForm>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
