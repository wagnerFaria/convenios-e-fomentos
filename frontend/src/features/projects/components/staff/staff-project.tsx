'use client';

import { parseAsString, useQueryState } from 'nuqs';
import { useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useCan } from '@/features/auth/session-provider';
import { accessOf } from '../../access';
import {
  framingStatusLabels,
  instrumentLabels,
  projectActions,
  stageLabels,
  statusLabels,
  type Project
} from '../../api/types';
import { DocumentsStep } from '../documents-step';
import { FundingStep } from '../funding-step';
import { InstitutionStep } from '../institution-step';
import { ProponentStep } from '../proponent-step';
import { ProposalStep } from '../proposal-step';
import { HistoryPanel } from './history-panel';
import { ReviewPanel } from './review-panel';
import { StaffActions } from './staff-actions';
import { StaffBanner } from './staff-banner';

// A equipe só lê: os formulários do cadastro abrem em modo somente leitura e nada é salvo.
const noop = () => undefined;

interface TabSpec {
  key: string;
  label: string;
}

const registrationTabs: TabSpec[] = [
  { key: 'proponente', label: 'Proponente' },
  { key: 'instituicao', label: 'Instituição' },
  { key: 'proposta', label: 'Proposta' },
  { key: 'recursos', label: 'Recursos' },
  { key: 'documentos', label: 'Documentos' }
];

/**
 * O projeto visto pela equipe (quem não é titular nem membro): cabeçalho com a situação, abas somente leitura do cadastro e dos
 * documentos, e o histórico. As ações e a conferência entram nos painéis ao lado.
 */
export function StaffProjectView({ project }: { project: Project }) {
  const can = useCan();
  const [tab, setTab] = useQueryState('aba', parseAsString.withDefault(''));
  // Relógio da tela, fixado na montagem (a leitura da hora não pode acontecer a cada renderização).
  const [now] = useState(() => Date.now());
  const access = accessOf(project, false, now);

  // A conferência é a primeira aba de quem confere; os demais perfis começam pelo cadastro.
  const reviewTab: TabSpec[] = project.review ? [{ key: 'conferencia', label: 'Conferência' }] : [];
  const tabs: TabSpec[] = [
    ...reviewTab,
    ...registrationTabs,
    { key: 'historico', label: 'Histórico' }
  ];
  const first = project.review && can(projectActions.reviewItem) ? 'conferencia' : 'proponente';
  const current = tabs.find((t) => t.key === tab) ?? tabs.find((t) => t.key === first) ?? tabs[0];
  const pending = project.review?.summary.pending ?? 0;
  const common = { project, onSaved: noop, onDirty: noop, readOnly: true };

  return (
    <div className='space-y-4'>
      <div className='flex flex-wrap items-center gap-2'>
        <h2 className='text-lg font-semibold'>{project.proposal.title || 'Projeto sem título'}</h2>
        <Badge variant={project.status === 'archived' ? 'outline' : 'secondary'}>
          {statusLabels[project.status]}
        </Badge>
        {project.institution.instrument && (
          <Badge variant='outline'>{instrumentLabels[project.institution.instrument]}</Badge>
        )}
        {project.status === 'submitted' && project.stage && (
          <Badge variant='outline'>{stageLabels[project.stage]}</Badge>
        )}
        {project.framing && (
          <Badge variant='outline'>
            {
              framingStatusLabels[
                project.framing.decision === 'confirmed' ? 'confirmed' : 'pending'
              ]
            }
          </Badge>
        )}
        {project.code && (
          <span className='text-muted-foreground font-mono text-sm'>{project.code}</span>
        )}
      </div>
      <p className='text-muted-foreground text-sm'>
        {project.institution.name || 'Instituição não informada'}
        {project.framing && <> · Departamento: {project.framing.department.name}</>}
        {project.analyst && <> · Analista: {project.analyst.name}</>}
      </p>
      <StaffBanner project={project} />
      <StaffActions project={project} />

      <Tabs value={current.key} onValueChange={(v) => void setTab(v)}>
        <TabsList className='h-auto flex-wrap justify-start'>
          {tabs.map((t) => (
            <TabsTrigger key={t.key} value={t.key}>
              {t.label}
              {t.key === 'conferencia' && pending > 0 && (
                <Badge
                  variant='secondary'
                  className='ml-1 tabular-nums'
                  aria-label={`${pending} pendentes`}
                >
                  {pending}
                </Badge>
              )}
            </TabsTrigger>
          ))}
        </TabsList>
        <TabsContent value='conferencia'>
          <ReviewPanel project={project} onGoTab={(t) => void setTab(t)} />
        </TabsContent>
        <TabsContent value='proponente' className='max-w-3xl'>
          <ProponentStep key={project.version} {...common} />
        </TabsContent>
        <TabsContent value='instituicao' className='max-w-3xl'>
          <InstitutionStep key={project.version} {...common} />
        </TabsContent>
        <TabsContent value='proposta' className='max-w-3xl'>
          <ProposalStep key={project.version} {...common} />
        </TabsContent>
        <TabsContent value='recursos' className='max-w-3xl'>
          <FundingStep key={project.version} {...common} />
        </TabsContent>
        <TabsContent value='documentos' className='max-w-3xl'>
          <DocumentsStep project={project} access={access} />
        </TabsContent>
        <TabsContent value='historico'>
          <HistoryPanel project={project} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
