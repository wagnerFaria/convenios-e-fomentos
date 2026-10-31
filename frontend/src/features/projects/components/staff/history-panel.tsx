'use client';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { formatDateTime } from '@/lib/format';
import { framingDecisionLabels, type Framing, type Project } from '../../api/types';

function FramingRow({ f }: { f: Framing }) {
  return (
    <li className='space-y-1 rounded-md border p-3 text-sm'>
      <div className='flex flex-wrap items-center gap-2'>
        <span className='font-medium'>
          Indicação nº {f.round}: {f.department.name}
        </span>
        <Badge variant={f.decision === 'refused' ? 'destructive' : 'secondary'}>
          {framingDecisionLabels[f.decision]}
        </Badge>
      </div>
      <p className='text-muted-foreground text-xs'>
        Indicado por {f.indicatedBy || '—'} em {formatDateTime(f.indicatedAt)}
        {f.decidedAt && (
          <>
            {' · '}decidido por {f.decidedBy || '—'} em {formatDateTime(f.decidedAt)}
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

/** Histórico do projeto para a equipe: rodadas de enquadramento e o pedido de complementação mais recente. */
export function HistoryPanel({ project }: { project: Project }) {
  const framings = project.framings ?? [];
  const c = project.complement;
  return (
    <div className='max-w-3xl space-y-4'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Enquadramento</CardTitle>
          <CardDescription>
            Cada vez que o Analista indica um departamento, abre-se uma indicação para o Revisor do
            Departamento decidir. Não há limite de recusas.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {framings.length === 0 ? (
            <p className='text-muted-foreground text-sm'>Nenhum departamento foi indicado ainda.</p>
          ) : (
            <ul className='space-y-2'>
              {framings.map((f) => (
                <FramingRow key={f.round} f={f} />
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Pedido de complementação</CardTitle>
        </CardHeader>
        <CardContent className='space-y-2 text-sm'>
          {!c ? (
            <p className='text-muted-foreground'>Nenhuma complementação foi pedida.</p>
          ) : (
            <>
              <p>
                Pedido nº {c.number} em {formatDateTime(c.requestedAt)}, com prazo até{' '}
                {formatDateTime(c.dueAt)}.{' '}
                <Badge variant='outline'>
                  {
                    {
                      open: 'Aberto',
                      answered: 'Respondido',
                      expired: 'Vencido',
                      cancelled: 'Cancelado'
                    }[c.status]
                  }
                </Badge>
              </p>
              {c.message && <p className='whitespace-pre-line'>{c.message}</p>}
              <ul className='list-disc space-y-1 pl-5'>
                {c.items.map((it) => (
                  <li key={`${it.kind}-${it.target}`}>
                    <span className='font-medium'>{it.label}</span>
                    {it.note && <span className='text-muted-foreground'>: {it.note}</span>}
                  </li>
                ))}
              </ul>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
