'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { LoadingButton } from '@/components/ui/loading-button';
import { useCan } from '@/features/auth/session-provider';
import { projectKeys } from '../../api/queries';
import { archiveProject, startAnalysis } from '../../api/service';
import { projectActions, type Project } from '../../api/types';
import { AssignDepartmentDialog, ComplementDialog, NoteDialog } from './dialogs';
import { reportStaffError } from './errors';

function StartAnalysisButton({ project }: { project: Project }) {
  const qc = useQueryClient();
  const run = useMutation({
    mutationFn: () => startAnalysis(project.id),
    onSuccess: async (saved) => {
      qc.setQueryData(projectKeys.detail(project.id), saved);
      await qc.invalidateQueries({ queryKey: projectKeys.lists() });
      toast.success('Análise iniciada. Você é o analista responsável por este projeto.');
    },
    onError: async (err) => {
      if (reportStaffError(err).reload)
        await qc.invalidateQueries({ queryKey: projectKeys.detail(project.id) });
    }
  });
  return (
    <LoadingButton loading={run.isPending} onClick={() => run.mutate()}>
      <Icons.circleCheck aria-hidden /> Iniciar análise
    </LoadingButton>
  );
}

/** Ações do Analista, só as que cabem na etapa e na conferência; o servidor decide e a mensagem do erro explica. */
function AnalystActions({ project }: { project: Project }) {
  const can = useCan();
  if (project.status !== 'submitted') return null;
  const stage = project.stage;
  const framingPending = project.framing?.decision === 'pending';
  const working = stage === 'awaiting_checklist' || stage === 'in_analysis';
  const review = project.review;

  const showAssign =
    can(projectActions.assignDepartment) && stage === 'in_analysis' && !framingPending;
  const readyToAssign = !!review?.allApproved;
  const missing = review
    ? `${review.summary.pending} pendente(s) e ${review.summary.blockers} impedimento(s)`
    : '';

  return (
    <>
      {can(projectActions.startAnalysis) && stage === 'awaiting_checklist' && (
        <StartAnalysisButton project={project} />
      )}
      {showAssign && (
        <AssignDepartmentDialog
          project={project}
          button={{
            label: 'Indicar departamento',
            disabled: !readyToAssign,
            title: readyToAssign ? undefined : `Aprove todos os itens da conferência (${missing}).`
          }}
        />
      )}
      {can(projectActions.requestComplement) && working && !framingPending && (
        <ComplementDialog
          project={project}
          button={{ label: 'Devolver para complementação', variant: 'outline' }}
        />
      )}
      {can(projectActions.archive) && working && !framingPending && (
        <NoteDialog
          project={project}
          title='Arquivar por falta de documentos obrigatórios'
          description='O projeto é arquivado e deixa de andar (Portaria 71, art. 1º, §2º). O proponente, o Revisor do Departamento, o Analista e os Supervisores são avisados por e-mail.'
          warning='O arquivamento é definitivo. Se o objetivo é pedir ajustes, use "Devolver para complementação".'
          label='Motivo (obrigatório)'
          confirmLabel='Arquivar projeto'
          success='Projeto arquivado.'
          destructive
          submit={(note) => archiveProject(project.id, note)}
          button={{ label: 'Arquivar por falta de documentos', variant: 'outline' }}
        />
      )}
    </>
  );
}

/** Barra de ações da equipe sobre o projeto, conforme o perfil, a etapa e a situação do enquadramento. */
export function StaffActions({ project }: { project: Project }) {
  const actions = <AnalystActions project={project} />;
  return <div className='flex flex-wrap items-center gap-2'>{actions}</div>;
}
