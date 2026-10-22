'use client';

import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { QueryError } from '@/components/query-error';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan } from '@/features/auth/session-provider';
import { formatDateTime } from '@/lib/format';
import { auditLogsQueryOptions } from '../api/queries';
import { auditActions, type AuditLog } from '../api/types';
import { auditActionLabel } from '../lib/labels';
import { actorLabel, AuditLogSheet } from './audit-log-sheet';

/** Últimas alterações de uma entidade; some para quem não pode ler a auditoria. */
export function AuditHistory({ entity, entityId }: { entity: string; entityId: string }) {
  const can = useCan();
  const [open, setOpen] = useState<AuditLog | null>(null);
  const logs = useQuery({
    ...auditLogsQueryOptions({ entity, entityId, pageSize: 10 }),
    enabled: can(auditActions.read)
  });

  if (!can(auditActions.read)) return null;
  if (logs.isLoading) return <Skeleton className='h-24 w-full' />;
  if (logs.isError) return <QueryError error={logs.error} what='o histórico' />;
  const rows = logs.data?.data ?? [];
  if (rows.length === 0) {
    return <p className='text-muted-foreground text-sm'>Nenhuma alteração registrada.</p>;
  }
  return (
    <>
      <ul className='grid gap-1.5'>
        {rows.map((l) => (
          <li key={l.id}>
            <button
              type='button'
              className='hover:bg-muted flex w-full flex-wrap items-baseline gap-x-3 rounded-md px-2 py-1 text-left text-sm'
              onClick={() => setOpen(l)}
            >
              <span className='text-muted-foreground tabular-nums'>
                {formatDateTime(l.createdAt)}
              </span>
              <span className='font-medium'>{auditActionLabel(l.action)}</span>
              <span className='text-muted-foreground'>por {actorLabel(l)}</span>
            </button>
          </li>
        ))}
      </ul>
      <AuditLogSheet log={open} onClose={() => setOpen(null)} />
    </>
  );
}
