'use client';

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { formatDateTime } from '@/lib/format';
import type { AuditLog } from '../api/types';
import { auditActionLabel, auditEntityLabel } from '../lib/labels';
import { PayloadDiff } from './payload-diff';

export const actorLabel = (l: Pick<AuditLog, 'actorType' | 'actorName' | 'actorId'>) =>
  l.actorName ??
  (l.actorType === 'system'
    ? 'Sistema'
    : l.actorType === 'admin'
      ? 'Administrador removido'
      : l.actorType);

/** Detalhe de um registro de auditoria, com o antes/depois da alteração. */
export function AuditLogSheet({ log, onClose }: { log: AuditLog | null; onClose: () => void }) {
  return (
    <Sheet open={log !== null} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className='w-full overflow-y-auto sm:max-w-2xl'>
        {log && (
          <>
            <SheetHeader>
              <SheetTitle>{auditActionLabel(log.action)}</SheetTitle>
              <SheetDescription>
                {formatDateTime(log.createdAt)} · {actorLabel(log)}
                {log.ip ? ` · IP ${log.ip}` : ''}
              </SheetDescription>
            </SheetHeader>
            <div className='grid gap-4 px-4 pb-4'>
              <dl className='grid grid-cols-[6rem_1fr] gap-y-1 text-sm'>
                <dt className='text-muted-foreground'>Entidade</dt>
                <dd>{auditEntityLabel(log.entity)}</dd>
                <dt className='text-muted-foreground'>Identificador</dt>
                <dd className='font-mono text-xs break-all'>{log.entityId ?? '—'}</dd>
                <dt className='text-muted-foreground'>Ação</dt>
                <dd className='font-mono text-xs'>{log.action}</dd>
              </dl>
              <PayloadDiff before={log.payload?.before} after={log.payload?.after} />
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  );
}
