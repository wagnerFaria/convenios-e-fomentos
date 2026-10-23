'use client';

import { Icons } from '@/components/icons';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { sectionLabels, type Issue, type SectionKey } from '../api/types';

/** Pendências do projeto, agrupadas por etapa; clicar leva à etapa. */
export function IssuesPanel({
  issues,
  onGo
}: {
  issues: Issue[];
  onGo: (section: SectionKey) => void;
}) {
  const blocks = issues.filter((i) => i.severity === 'block');
  const warns = issues.filter((i) => i.severity === 'warn');
  const sections = (Object.keys(sectionLabels) as SectionKey[]).filter((s) =>
    issues.some((i) => i.section === s)
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          {blocks.length === 0 ? (
            <Icons.circleCheck className='size-4 text-green-600' aria-hidden />
          ) : (
            <Icons.alertCircle className='size-4' aria-hidden />
          )}
          Pendências
          {blocks.length > 0 && <Badge variant='secondary'>{blocks.length}</Badge>}
        </CardTitle>
        <CardDescription>
          {blocks.length === 0
            ? 'Nada impede o envio nas etapas já preenchidas.'
            : 'O que falta para o projeto poder ser enviado. O rascunho pode ser salvo incompleto.'}
        </CardDescription>
      </CardHeader>
      {sections.length > 0 && (
        <CardContent className='space-y-4'>
          {sections.map((s) => (
            <section key={s} aria-label={sectionLabels[s]}>
              <button
                type='button'
                className='text-sm font-medium underline-offset-4 hover:underline'
                onClick={() => onGo(s)}
              >
                {sectionLabels[s]}
              </button>
              <ul className='mt-1 space-y-1'>
                {issues
                  .filter((i) => i.section === s)
                  .map((i) => (
                    <li
                      key={`${i.field}-${i.code}`}
                      className='text-muted-foreground flex gap-2 text-sm'
                    >
                      <span aria-hidden>{i.severity === 'block' ? '•' : '⚠'}</span>
                      <span>
                        {i.severity === 'warn' && <span className='sr-only'>Aviso: </span>}
                        {i.message}
                      </span>
                    </li>
                  ))}
              </ul>
            </section>
          ))}
          {warns.length > 0 && blocks.length === 0 && (
            <p className='text-muted-foreground text-xs'>Avisos não impedem o envio.</p>
          )}
        </CardContent>
      )}
    </Card>
  );
}
