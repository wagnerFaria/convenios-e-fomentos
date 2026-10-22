'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import { QueryError } from '@/components/query-error';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FieldGroup } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { emailTemplatesQueryOptions, endUserKeys } from '../api/queries';
import { updateEmailTemplate } from '../api/service';
import { endUserActions, templateLabels, type EmailTemplate } from '../api/types';

const schema = z.object({
  senderName: z.string().max(100, 'Máximo de 100 caracteres.'),
  senderEmail: z.string().trim().email('E-mail inválido.'),
  replyTo: z.string(),
  subject: z.string().trim().min(1, 'Informe o assunto.').max(300, 'Máximo de 300 caracteres.'),
  bodyHtml: z.string().trim().min(1, 'Informe o corpo da mensagem.')
});

const variables = [
  ['{{.User.Username}}', 'nome de usuário'],
  ['{{.User.Email}}', 'e-mail do usuário'],
  ['{{.URL}}', 'link de ação (confirmar e-mail ou redefinir senha)'],
  ['{{.Token}}', 'código de uso único'],
  ['{{.ServerURL}}', 'endereço da API']
] as const;

function TemplateForm({ template }: { template: EmailTemplate }) {
  const can = useCan();
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (v: z.infer<typeof schema>) => updateEmailTemplate(template.key, v),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: endUserKeys.templates() });
      toast.success('Template salvo.');
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: {
      senderName: template.senderName,
      senderEmail: template.senderEmail,
      replyTo: template.replyTo,
      subject: template.subject,
      bodyHtml: template.bodyHtml
    },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      return save.mutateAsync(value);
    }
  });
  const editable = can(endUserActions.templates.update);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <FieldGroup>
        <div className='grid gap-4 sm:grid-cols-2'>
          <form.AppField
            name='senderName'
            children={(f) => <f.TextField label='Nome do remetente' disabled={!editable} />}
          />
          <form.AppField
            name='senderEmail'
            children={(f) => (
              <f.TextField label='E-mail do remetente' type='email' required disabled={!editable} />
            )}
          />
        </div>
        <form.AppField
          name='replyTo'
          children={(f) => <f.TextField label='Responder para' type='email' disabled={!editable} />}
        />
        <form.AppField
          name='subject'
          children={(f) => <f.TextField label='Assunto' required disabled={!editable} />}
        />
        <form.AppField
          name='bodyHtml'
          children={(f) => (
            <f.TextareaField
              label='Corpo (HTML)'
              required
              rows={12}
              className='font-mono text-xs'
              disabled={!editable}
            />
          )}
        />
        <div className='text-muted-foreground grid gap-1 text-sm'>
          <span className='font-medium'>Variáveis disponíveis</span>
          {variables.map(([name, desc]) => (
            <span key={name}>
              <code>{name}</code> — {desc}
            </span>
          ))}
        </div>
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        {editable && (
          <form.AppForm>
            <form.SubmitButton>Salvar template</form.SubmitButton>
          </form.AppForm>
        )}
      </FieldGroup>
    </form>
  );
}

export function EmailTemplatesPanel() {
  const templates = useQuery(emailTemplatesQueryOptions());
  if (templates.isLoading) return <Skeleton className='h-60 w-full max-w-3xl' />;
  if (templates.isError || !templates.data) {
    return <QueryError error={templates.error} what='os templates' />;
  }
  const first = templates.data[0]?.key;
  return (
    <Card className='max-w-3xl'>
      <CardHeader>
        <CardTitle>E-mails enviados aos usuários</CardTitle>
        <CardDescription>
          O template é validado ao salvar (sintaxe e variáveis) antes de entrar em uso.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue={first}>
          <TabsList>
            {templates.data.map((t) => (
              <TabsTrigger key={t.key} value={t.key}>
                {templateLabels[t.key] ?? t.key}
              </TabsTrigger>
            ))}
          </TabsList>
          {templates.data.map((t) => (
            <TabsContent key={t.key} value={t.key} className='pt-4'>
              <TemplateForm key={t.updatedAt} template={t} />
            </TabsContent>
          ))}
        </Tabs>
      </CardContent>
    </Card>
  );
}
