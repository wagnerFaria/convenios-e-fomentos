'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import { QueryError } from '@/components/query-error';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FieldGroup } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { endUserKeys, endUserRolesQueryOptions, endUserSettingsQueryOptions } from '../api/queries';
import { updateEndUserSettings } from '../api/service';
import { endUserActions, type EndUserSettings } from '../api/types';

const schema = z.object({
  defaultRoleId: z.string().min(1, 'Escolha o perfil padrão.'),
  oneAccountPerEmail: z.boolean(),
  registrationEnabled: z.boolean(),
  emailConfirmationRequired: z.boolean(),
  resetPasswordUrl: z.string(),
  emailConfirmationRedirectUrl: z.string(),
  registerAllowedFields: z.array(z.string())
});

function SettingsForm({ settings }: { settings: EndUserSettings }) {
  const can = useCan();
  const qc = useQueryClient();
  const roles = useQuery(endUserRolesQueryOptions());
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: updateEndUserSettings,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.settings() });
      toast.success('Configurações salvas.');
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: settings,
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync(value);
    }
  });
  const editable = can(endUserActions.settings.update);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.AppField
          name='defaultRoleId'
          children={(f) => (
            <f.SelectField
              label='Perfil padrão'
              description='Atribuído a quem se cadastra. Não pode ser excluído enquanto for o padrão.'
              options={(roles.data ?? []).map((r) => ({ value: r.id, label: r.name }))}
            />
          )}
        />
        <form.AppField
          name='registrationEnabled'
          children={(f) => (
            <f.SwitchField
              label='Cadastro público habilitado'
              description='Permite que qualquer pessoa crie uma conta.'
            />
          )}
        />
        <form.AppField
          name='oneAccountPerEmail'
          children={(f) => (
            <f.SwitchField
              label='Uma conta por e-mail'
              description='Impede contas diferentes com o mesmo e-mail.'
            />
          )}
        />
        <form.AppField
          name='emailConfirmationRequired'
          children={(f) => (
            <f.SwitchField
              label='Exigir confirmação de e-mail'
              description='Só entra depois de confirmar o e-mail.'
            />
          )}
        />
        <form.AppField
          name='resetPasswordUrl'
          children={(f) => (
            <f.TextField
              label='URL de redefinição de senha'
              type='url'
              description='Página do seu site que recebe o código enviado por e-mail.'
              placeholder='https://…'
            />
          )}
        />
        <form.AppField
          name='emailConfirmationRedirectUrl'
          children={(f) => (
            <f.TextField
              label='URL de redirecionamento após confirmar o e-mail'
              type='url'
              placeholder='https://…'
            />
          )}
        />
        <form.AppField
          name='registerAllowedFields'
          mode='array'
          children={(f) => (
            <f.TagsField
              label='Campos extras aceitos no cadastro'
              description='Ficam no perfil do usuário. Nomes reservados (como role ou password) são recusados.'
              placeholder='Digite o nome do campo e pressione Enter'
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
            <form.SubmitButton>Salvar configurações</form.SubmitButton>
          </form.AppForm>
        )}
      </FieldGroup>
    </form>
  );
}

export function EndUserSettingsPanel() {
  const settings = useQuery(endUserSettingsQueryOptions());
  if (settings.isLoading) return <Skeleton className='h-60 w-full max-w-2xl' />;
  if (settings.isError || !settings.data) {
    return <QueryError error={settings.error} what='as configurações' />;
  }
  return (
    <Card className='max-w-2xl'>
      <CardHeader>
        <CardTitle>Cadastro e acesso</CardTitle>
        <CardDescription>Como os usuários finais se cadastram e entram.</CardDescription>
      </CardHeader>
      <CardContent>
        <SettingsForm key={JSON.stringify(settings.data)} settings={settings.data} />
      </CardContent>
    </Card>
  );
}
