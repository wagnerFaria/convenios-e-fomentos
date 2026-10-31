'use client';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { framingDecisionLabels, type Framing, type Project } from '../../api/types';
import { formatDateTimeCuiaba } from '../../format';

function FramingRow({ f, current }: { f: Framing; current: boolean }) {
  return (
    <li className='space-y-1 rounded-md border p-3 text-sm'>
      <div className='flex flex-wrap items-center gap-2'>
        <span className='font-medium'>
          Indicação nº {f.round}: {f.department.name}
        </span>
        <Badge variant={f.decision === 'refused' ? 'destructive' : 'secondary'}>
          {framingDecisionLabels[f.decision]}
        </Badge>
        {current && <Badge variant='outline'>Em curso</Badge>}
      </div>
      <p className='text-muted-foreground text-xs'>
        Indicado por {f.indicatedBy || '—'} em {formatDateTimeCuiaba(f.indicatedAt)}
        {f.decidedAt && (
          <>
            {' · '}decidido por {f.decidedBy || '—'} em {formatDateTimeCuiaba(f.decidedAt)}
          </>
        )}
      </p>
      {f.justification && (
        <p className='whitespace-pre-line'>
          <span className='font-medium'>Justificativa: </span>
          {f.justification}
        </p>
      )}
    </li>
  );
}

/** Enquadramento: a indicação em curso e o histórico de todas as rodadas, da mais nova para a mais antiga. */
export function FramingPanel({ project }: { project: Project }) {
  const framings = project.framings ?? [];
  const ordered = framings.toReversed();
  return (
    <div className='max-w-3xl space-y-4'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Enquadramento no departamento</CardTitle>
          <CardDescription>
            O Analista indica um departamento da área finalística e o Revisor do Departamento
            confirma ou recusa. A primeira decisão vale, e não há limite de recusas: ao recusar, o
            projeto volta ao Analista, que indica outro departamento.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {ordered.length === 0 ? (
            <p className='text-muted-foreground text-sm'>Nenhum departamento foi indicado ainda.</p>
          ) : (
            <ul className='space-y-2'>
              {ordered.map((f) => (
                <FramingRow key={f.round} f={f} current={project.framing?.round === f.round} />
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
