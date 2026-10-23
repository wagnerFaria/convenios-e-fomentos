'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LoadingButton } from '@/components/ui/loading-button';
import { Progress } from '@/components/ui/progress';
import { Skeleton } from '@/components/ui/skeleton';
import { QueryError } from '@/components/query-error';
import { FieldGroup } from '@/components/ui/field';
import { useAppForm } from '@/lib/form';
import { checklistQueryOptions, projectKeys } from '../api/queries';
import { saveDocumentOptions } from '../api/service';
import type { Access } from '../access';
import type { ChecklistItem, Project } from '../api/types';
import { ChecklistItemCard } from './checklist-item';
import { ServerErrors } from './server-errors';

/** Situações informadas pelo proponente que acrescentam documentos ao checklist. */
function OptionsForm({ project, readOnly }: { project: Project; readOnly: boolean }) {
  const qc = useQueryClient();
  const o = project.documentOptions;
  const fomento = project.institution.instrument === 'fomento';
  const save = useMutation({
    mutationFn: (v: typeof o) =>
      saveDocumentOptions(project.id, { version: project.version, ...v }),
    onSuccess: async (saved) => {
      qc.setQueryData(projectKeys.detail(project.id), saved);
      await qc.invalidateQueries({ queryKey: projectKeys.checklist(project.id) });
      toast.success('Situações salvas. O checklist foi atualizado.');
    }
  });
  const form = useAppForm({
    defaultValues: { ...o },
    onSubmit: ({ value }) => save.mutateAsync(value).catch(() => undefined)
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Situações que pedem documentos adicionais</CardTitle>
          <CardDescription>
            Marque o que se aplica ao projeto; o checklist abaixo se ajusta.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <form.AppField
              name='exclusiveSupplier'
              children={(f) => (
                <f.SwitchField
                  label='Contratação por exclusividade'
                  description='Fornecedor ou artista exclusivo: pede a documentação de exclusividade.'
                />
              )}
            />
            {fomento && (
              <form.AppField
                name='publicSpaceUse'
                children={(f) => (
                  <f.SwitchField
                    label='Atividade em espaço público'
                    description='Pede a autorização de uso do espaço.'
                  />
                )}
              />
            )}
            <form.AppField
              name='reducedDeadline'
              children={(f) => (
                <f.SwitchField
                  label='Pedido de redução do prazo de tramitação'
                  description='Pede a justificativa (modelo da SEDEC).'
                />
              )}
            />
            <ServerErrors error={save.error} />
            <div>
              <LoadingButton type='submit' size='sm' loading={save.isPending} disabled={readOnly}>
                Salvar situações
              </LoadingButton>
            </div>
          </FieldGroup>
        </CardContent>
      </Card>
    </form>
  );
}

export function DocumentsStep({ project, access }: { project: Project; access: Access }) {
  const checklist = useQuery(checklistQueryOptions(project.id));

  if (checklist.isError)
    return <QueryError error={checklist.error} what='o checklist de documentos' />;
  if (checklist.isLoading || !checklist.data) return <Skeleton className='h-80 w-full' />;
  const { summary, items, instrument } = checklist.data;
  const delivered = summary.done + summary.justified;
  const groups = Array.from(new Set(items.map((i) => i.group)));
  const byGroup = (g: string): ChecklistItem[] => items.filter((i) => i.group === g);

  return (
    <div className='space-y-6'>
      {!instrument ? (
        <Alert>
          <AlertDescription>
            Escolha o instrumento na etapa Instituição para ver o checklist de documentos.
          </AlertDescription>
        </Alert>
      ) : (
        <>
          <div className='space-y-2'>
            <div className='flex flex-wrap items-baseline justify-between gap-2'>
              <p className='text-sm font-medium'>
                {delivered} de {summary.total} documentos entregues
                {summary.justified > 0 && (
                  <span className='text-muted-foreground font-normal'>
                    {' '}
                    ({summary.justified} com justificativa de não envio)
                  </span>
                )}
              </p>
              <p className='text-muted-foreground text-sm'>{summary.missing} pendentes</p>
            </div>
            <Progress
              value={summary.total ? Math.round((delivered / summary.total) * 100) : 0}
              aria-label='Documentos entregues'
            />
            <p className='text-muted-foreground text-xs'>
              Aceitamos PDF, DOCX, PNG, JPEG e WebP, com até 10 arquivos por documento. Os modelos
              são os publicados pela SEDEC.
            </p>
          </div>
          <OptionsForm
            key={project.version}
            project={project}
            readOnly={
              !(
                access.draft ||
                access.complement?.items.some((i) => i.target === 'document-options')
              )
            }
          />
          {groups.map((g) => (
            <section key={g} aria-labelledby={`group-${g}`} className='space-y-3'>
              <h3 id={`group-${g}`} className='text-base font-semibold'>
                {g}
              </h3>
              <ul className='space-y-3'>
                {byGroup(g).map((item) => (
                  <ChecklistItemCard
                    key={`${item.code}-${item.amendmentId}-${item.justification}`}
                    projectId={project.id}
                    item={item}
                    editable={access.document(item)}
                  />
                ))}
              </ul>
            </section>
          ))}
        </>
      )}
    </div>
  );
}
