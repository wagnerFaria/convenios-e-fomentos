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
import { FieldGroup } from '@/components/ui/field';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { departmentKeys } from '../api/queries';
import { createDepartment, updateDepartment } from '../api/service';
import { departmentActions, type Department } from '../api/types';

const schema = z.object({
  name: z.string().trim().min(1, 'Informe o nome.').max(100, 'Máximo de 100 caracteres.')
});

function DepartmentForm({ department, onDone }: { department?: Department; onDone: () => void }) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (name: string) =>
      department ? updateDepartment(department.id, { name }) : createDepartment({ name }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: departmentKeys.all });
      toast.success(department ? 'Departamento renomeado.' : 'Departamento criado.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: { name: department?.name ?? '' },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync(value.name.trim());
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
        <form.AppField name='name' children={(f) => <f.TextField label='Nome' required />} />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>{department ? 'Salvar' : 'Criar departamento'}</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

/** Sem `department`, é o botão "Novo departamento"; com ele, o diálogo (controlado) de renomear. */
export function DepartmentDialog(
  props:
    | { department: Department; open: boolean; onOpenChange: (open: boolean) => void }
    | {
        department?: undefined;
        open?: undefined;
        onOpenChange?: undefined;
      }
) {
  const can = useCan();
  const [open, setOpen] = useState(false);

  if (props.department) {
    return (
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Renomear departamento</DialogTitle>
            <DialogDescription>
              Administradores e documentos vinculados continuam nele.
            </DialogDescription>
          </DialogHeader>
          <DepartmentForm department={props.department} onDone={() => props.onOpenChange(false)} />
        </DialogContent>
      </Dialog>
    );
  }
  if (!can(departmentActions.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>
        <Icons.add aria-hidden /> Novo departamento
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo departamento</DialogTitle>
          <DialogDescription>
            Departamentos delimitam o alcance das permissões com a condição &quot;mesmo
            departamento&quot;.
          </DialogDescription>
        </DialogHeader>
        <DepartmentForm onDone={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}
