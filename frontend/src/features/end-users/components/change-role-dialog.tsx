'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
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
import { endUserKeys, endUserRolesQueryOptions } from '../api/queries';
import { setEndUserRole } from '../api/service';
import type { EndUser } from '../api/types';

export function ChangeRoleDialog({
  user,
  open,
  onOpenChange
}: {
  user: EndUser;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const qc = useQueryClient();
  const roles = useQuery(endUserRolesQueryOptions());
  const [error, setError] = useState<string | null>(null);
  const change = useMutation({
    mutationFn: (roleId: string) => setEndUserRole(user.id, roleId),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.all });
      toast.success('Perfil alterado.');
      onOpenChange(false);
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: { roleId: user.role.id },
    onSubmit: ({ value }) => {
      setError(null);
      return change.mutateAsync(value.roleId);
    }
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Alterar perfil</DialogTitle>
          <DialogDescription>
            {user.username} ({user.email}). As permissões na API pública mudam na próxima
            requisição.
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
              name='roleId'
              children={(f) => (
                <f.SelectField
                  label='Perfil'
                  options={(roles.data ?? []).map((r) => ({ value: r.id, label: r.name }))}
                />
              )}
            />
            {error && (
              <p role='alert' className='text-destructive text-sm'>
                {error}
              </p>
            )}
            <div className='flex justify-end gap-2'>
              <Button type='button' variant='outline' onClick={() => onOpenChange(false)}>
                Cancelar
              </Button>
              <form.AppForm>
                <form.SubmitButton>Alterar</form.SubmitButton>
              </form.AppForm>
            </div>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
