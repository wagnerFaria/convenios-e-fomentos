'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Icons } from '@/components/icons';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LoadingButton } from '@/components/ui/loading-button';
import { Progress } from '@/components/ui/progress';
import { useCan } from '@/features/auth/session-provider';
import { projectKeys } from '../../api/queries';
import { getFileDownload, reviewItem } from '../../api/service';
import {
  projectActions,
  reviewKindLabels,
  type Project,
  type ReviewFile,
  type ReviewItem,
  type ReviewKind
} from '../../api/types';
import { reportStaffError } from './errors';
import { formatDateTimeCuiaba } from '../../format';

/** Aba do cadastro que mostra cada seção da conferência. */
const sectionTab: Record<string, string> = {
  proponent: 'proponente',
  institution: 'instituicao',
  proposal: 'proposta',
  funding: 'recursos',
  'document-options': 'documentos'
};

/** O Analista pode aprovar e desfazer quando o projeto está em análise e nenhum enquadramento aguarda decisão. */
export function canReviewNow(project: Project, can: (action: string) => boolean): boolean {
  return (
    can(projectActions.reviewItem) &&
    project.status === 'submitted' &&
    project.stage === 'in_analysis' &&
    project.framing?.decision !== 'pending'
  );
}

function FileLine({ projectId, file }: { projectId: string; file: ReviewFile }) {
  const qc = useQueryClient();
  const open = useMutation({
    mutationFn: async () => (await getFileDownload(projectId, file.id)).url,
    onSuccess: async (url) => {
      const a = document.createElement('a');
      a.href = url;
      a.rel = 'noopener';
      a.click();
      // O servidor registrou a abertura: recarrega para o selo "Aberto" aparecer e liberar a aprovação.
      await qc.invalidateQueries({ queryKey: projectKeys.detail(projectId) });
    },
    onError: (err) => reportStaffError(err)
  });
  return (
    <li className='flex flex-wrap items-center justify-between gap-2 text-sm'>
      <span className='flex min-w-0 items-center gap-2'>
        <Icons.paperclip className='text-muted-foreground size-4 shrink-0' aria-hidden />
        <span className='truncate'>{file.filename}</span>
        {file.opened ? (
          <Badge variant='secondary'>Aberto</Badge>
        ) : (
          <Badge variant='outline'>Não aberto</Badge>
        )}
      </span>
      <LoadingButton
        size='sm'
        variant='outline'
        loading={open.isPending}
        onClick={() => open.mutate()}
        aria-label={`Abrir ${file.filename}`}
      >
        <Icons.eye aria-hidden /> Abrir
      </LoadingButton>
    </li>
  );
}

function ReviewRow({
  project,
  item,
  canReview,
  onGoTab
}: {
  project: Project;
  item: ReviewItem;
  canReview: boolean;
  onGoTab: (tab: string) => void;
}) {
  const qc = useQueryClient();
  const approved = item.status === 'approved';
  const allOpened = item.files.every((f) => f.opened);
  const change = useMutation({
    mutationFn: (value: boolean) =>
      reviewItem(project.id, { kind: item.kind, target: item.target, approved: value }),
    onSuccess: (saved) => qc.setQueryData(projectKeys.detail(project.id), saved),
    onError: async (err) => {
      if (reportStaffError(err).reload)
        await qc.invalidateQueries({ queryKey: projectKeys.detail(project.id) });
    }
  });
  const tab = item.kind === 'section' ? sectionTab[item.target] : undefined;

  return (
    <li className='space-y-2 rounded-md border p-3'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex min-w-0 items-center gap-2'>
          {approved ? (
            <Icons.circleCheck className='size-4 shrink-0 text-green-600' aria-hidden />
          ) : (
            <Icons.circle className='text-muted-foreground size-4 shrink-0' aria-hidden />
          )}
          <span className='font-medium'>{item.label}</span>
          {item.blocking ? (
            <Badge variant='destructive'>Impedimento</Badge>
          ) : approved ? (
            <Badge variant='secondary'>Aprovado</Badge>
          ) : (
            <Badge variant='outline'>Pendente</Badge>
          )}
        </div>
        <div className='flex gap-1'>
          {tab && (
            <Button size='sm' variant='ghost' onClick={() => onGoTab(tab)}>
              Ver dados
            </Button>
          )}
          {canReview && !item.blocking && (
            <LoadingButton
              size='sm'
              variant={approved ? 'outline' : 'default'}
              loading={change.isPending}
              disabled={!approved && !allOpened}
              onClick={() => change.mutate(!approved)}
            >
              {approved ? 'Desfazer' : 'Aprovar'}
            </LoadingButton>
          )}
        </div>
      </div>
      {item.blocking && item.reason && (
        <p className='text-sm'>{item.reason} Devolva o item ao proponente ou arquive o projeto.</p>
      )}
      {canReview && !item.blocking && !approved && !allOpened && (
        <p className='text-muted-foreground text-xs'>Abra todos os arquivos para poder aprovar.</p>
      )}
      {approved && item.reviewedBy && (
        <p className='text-muted-foreground text-xs'>
          Aprovado por {item.reviewedBy}
          {item.reviewedAt ? ` em ${formatDateTimeCuiaba(item.reviewedAt)}` : ''}
        </p>
      )}
      {item.files.length > 0 && (
        <ul className='space-y-1.5 border-t pt-2'>
          {item.files.map((f) => (
            <FileLine key={f.id} projectId={project.id} file={f} />
          ))}
        </ul>
      )}
    </li>
  );
}

/** Conferência da análise: o Analista abre cada documento e aprova item a item; os demais perfis acompanham. */
export function ReviewPanel({
  project,
  onGoTab
}: {
  project: Project;
  onGoTab: (tab: string) => void;
}) {
  const can = useCan();
  const review = project.review;
  if (!review) return null;
  const canReview = canReviewNow(project, can);
  const { total, approved, blockers } = review.summary;
  const percent = total ? Math.round((approved / total) * 100) : 0;

  const blocking = review.items.filter((i) => i.blocking);
  const kinds: ReviewKind[] = ['section', 'document', 'extra'];

  let hint = 'Somente leitura.';
  if (canReview) {
    hint =
      'Abra cada arquivo antes de aprovar o documento. Aprove todos os itens para poder indicar o departamento.';
  } else if (project.framing?.decision === 'pending') {
    hint = 'A conferência fica travada enquanto o departamento não decide o enquadramento.';
  } else if (project.status === 'submitted' && project.stage === 'awaiting_checklist') {
    hint = 'Inicie a análise para poder aprovar os itens.';
  } else if (project.stage === 'awaiting_proponent') {
    hint = 'O projeto está com o proponente; a conferência volta quando ele reenviar.';
  }

  return (
    <div className='max-w-3xl space-y-4'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Conferência</CardTitle>
          <CardDescription>{hint}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-2'>
          <p className='text-sm font-medium'>
            {approved} de {total} itens aprovados
            {blockers > 0 && (
              <span className='text-destructive'>
                {' '}
                · {blockers} impedimento{blockers > 1 ? 's' : ''}
              </span>
            )}
          </p>
          <Progress value={percent} aria-label='Itens aprovados' />
          {review.allApproved && (
            <p className='text-sm font-medium text-green-700'>
              Tudo aprovado: o departamento já pode ser indicado.
            </p>
          )}
        </CardContent>
      </Card>

      {blocking.length > 0 && (
        <section aria-labelledby='review-blockers' className='space-y-2'>
          <h3 id='review-blockers' className='text-base font-semibold'>
            Impedimentos
          </h3>
          <ul className='space-y-2'>
            {blocking.map((item) => (
              <ReviewRow
                key={`${item.kind}-${item.target}`}
                project={project}
                item={item}
                canReview={canReview}
                onGoTab={onGoTab}
              />
            ))}
          </ul>
        </section>
      )}

      {kinds.map((kind) => {
        const items = review.items.filter((i) => i.kind === kind && !i.blocking);
        if (items.length === 0) return null;
        return (
          <section key={kind} aria-labelledby={`review-${kind}`} className='space-y-2'>
            <h3 id={`review-${kind}`} className='text-base font-semibold'>
              {reviewKindLabels[kind]}
            </h3>
            <ul className='space-y-2'>
              {items.map((item) => (
                <ReviewRow
                  key={`${item.kind}-${item.target}`}
                  project={project}
                  item={item}
                  canReview={canReview}
                  onGoTab={onGoTab}
                />
              ))}
            </ul>
          </section>
        );
      })}
    </div>
  );
}
