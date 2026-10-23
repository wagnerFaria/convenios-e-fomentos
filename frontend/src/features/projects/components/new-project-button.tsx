'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { LoadingButton } from '@/components/ui/loading-button';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { projectKeys } from '../api/queries';
import { createProject } from '../api/service';
import { projectActions } from '../api/types';

export function NewProjectButton() {
  const can = useCan();
  const router = useRouter();
  const qc = useQueryClient();
  const create = useMutation({
    mutationFn: createProject,
    onSuccess: async (project) => {
      await qc.invalidateQueries({ queryKey: projectKeys.lists() });
      router.push(`/dashboard/projetos/${project.id}`);
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  if (!can(projectActions.create)) return null;
  return (
    <LoadingButton loading={create.isPending} onClick={() => create.mutate()}>
      <Icons.add aria-hidden /> Novo projeto
    </LoadingButton>
  );
}
