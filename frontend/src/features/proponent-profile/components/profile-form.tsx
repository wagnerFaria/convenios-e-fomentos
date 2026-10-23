'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { QueryError } from '@/components/query-error';
import { Skeleton } from '@/components/ui/skeleton';
import { LoadingButton } from '@/components/ui/loading-button';
import { useCan } from '@/features/auth/session-provider';
import { ContactEditor } from '@/features/projects/components/contact-editor';
import { contactValues } from '@/features/projects/components/proponent-step';
import { ServerErrors } from '@/features/projects/components/server-errors';
import { profileQueryOptions, projectKeys } from '@/features/projects/api/queries';
import { saveProfile } from '@/features/projects/api/service';
import { projectActions, type Profile } from '@/features/projects/api/types';
import { formatCpf } from '@/features/projects/format';
import { useAppForm } from '@/lib/form';

function Form({ profile }: { profile: Profile }) {
  const can = useCan();
  const qc = useQueryClient();
  const save = useMutation({
    mutationFn: saveProfile,
    onSuccess: (saved) => {
      qc.setQueryData(projectKeys.profile(), saved);
      toast.success('Perfil salvo.');
    }
  });
  const form = useAppForm({
    defaultValues: { contact: contactValues(profile) },
    onSubmit: ({ value }) => save.mutateAsync(value.contact).catch(() => undefined)
  });

  return (
    <form
      className='max-w-3xl space-y-6'
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <div className='bg-muted/40 rounded-lg border p-3 text-sm'>
        <p className='font-medium'>{profile.name}</p>
        <p className='text-muted-foreground'>
          CPF {formatCpf(profile.cpf)} · dados do seu login no MT Login
        </p>
      </div>
      <form.Field
        name='contact'
        children={(f) => <ContactEditor value={f.state.value} onChange={f.handleChange} />}
      />
      <ServerErrors error={save.error} />
      <LoadingButton
        type='submit'
        loading={save.isPending}
        disabled={!can(projectActions.profileUpdate)}
      >
        Salvar perfil
      </LoadingButton>
    </form>
  );
}

export function ProfileForm() {
  const profile = useQuery(profileQueryOptions());
  if (profile.isError) return <QueryError error={profile.error} what='o perfil' />;
  if (profile.isLoading || !profile.data) return <Skeleton className='h-80 w-full' />;
  // A chave refaz o formulário com o que o servidor normalizou depois de salvar.
  return <Form key={JSON.stringify(profile.data)} profile={profile.data} />;
}
