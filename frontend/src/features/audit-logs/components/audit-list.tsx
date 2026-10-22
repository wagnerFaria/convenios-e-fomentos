'use client';

import { useQuery } from '@tanstack/react-query';
import { parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import { useState } from 'react';
import { Pager } from '@/components/pager';
import { QueryError } from '@/components/query-error';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { formatDateTime } from '@/lib/format';
import { auditLogsQueryOptions } from '../api/queries';
import type { AuditLog } from '../api/types';
import {
  auditActionLabel,
  auditActionLabels,
  auditEntityLabel,
  auditEntityLabels
} from '../lib/labels';
import { actorLabel, AuditLogSheet } from './audit-log-sheet';

const ALL = 'all';
const actionItems = [
  { value: ALL, label: 'Todas as ações' },
  ...Object.entries(auditActionLabels).map(([value, label]) => ({ value, label }))
];
const entityItems = [
  { value: ALL, label: 'Todas as entidades' },
  ...Object.entries(auditEntityLabels).map(([value, label]) => ({ value, label }))
];

/** Início do dia (local) em RFC 3339. */
const startOfDay = (d: string) => (d ? new Date(`${d}T00:00:00`).toISOString() : undefined);
/** Fim do dia (local) em RFC 3339. */
const endOfDay = (d: string) => (d ? new Date(`${d}T23:59:59.999`).toISOString() : undefined);

function FilterSelect({
  items,
  value,
  onChange,
  label
}: {
  items: { value: string; label: string }[];
  value: string;
  onChange: (v: string) => void;
  label: string;
}) {
  return (
    <Select
      items={items}
      value={value || ALL}
      onValueChange={(v) => onChange(v === ALL ? '' : (v ?? ''))}
    >
      <SelectTrigger aria-label={label}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {items.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}

export function AuditList() {
  const [f, setF] = useQueryStates(
    {
      action: parseAsString.withDefault(''),
      entity: parseAsString.withDefault(''),
      from: parseAsString.withDefault(''),
      to: parseAsString.withDefault(''),
      page: parseAsInteger.withDefault(1)
    },
    { clearOnDefault: true }
  );
  const [open, setOpen] = useState<AuditLog | null>(null);
  const logs = useQuery({
    ...auditLogsQueryOptions({
      action: f.action || undefined,
      entity: f.entity || undefined,
      from: startOfDay(f.from),
      to: endOfDay(f.to),
      page: f.page
    }),
    placeholderData: (prev) => prev
  });

  return (
    <div className='grid gap-4'>
      <div className='flex flex-wrap items-center gap-2'>
        <FilterSelect
          label='Filtrar por ação'
          items={actionItems}
          value={f.action}
          onChange={(action) => void setF({ action, page: 1 })}
        />
        <FilterSelect
          label='Filtrar por entidade'
          items={entityItems}
          value={f.entity}
          onChange={(entity) => void setF({ entity, page: 1 })}
        />
        <div className='text-muted-foreground flex items-center gap-1.5 text-sm'>
          <label htmlFor='audit-from'>De</label>
          <Input
            id='audit-from'
            type='date'
            value={f.from}
            max={f.to || undefined}
            onChange={(e) => void setF({ from: e.target.value, page: 1 })}
            className='w-auto'
          />
        </div>
        <div className='text-muted-foreground flex items-center gap-1.5 text-sm'>
          <label htmlFor='audit-to'>até</label>
          <Input
            id='audit-to'
            type='date'
            value={f.to}
            min={f.from || undefined}
            onChange={(e) => void setF({ to: e.target.value, page: 1 })}
            className='w-auto'
          />
        </div>
      </div>

      {logs.isError ? (
        <QueryError error={logs.error} what='o registro de auditoria' />
      ) : logs.isLoading ? (
        <Skeleton className='h-40 w-full' />
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Data e hora</TableHead>
                <TableHead>Ator</TableHead>
                <TableHead>Ação</TableHead>
                <TableHead>Entidade</TableHead>
                <TableHead>IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.data?.data.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className='text-muted-foreground text-center'>
                    Nenhum registro encontrado.
                  </TableCell>
                </TableRow>
              )}
              {logs.data?.data.map((l) => (
                <TableRow key={l.id} className='cursor-pointer' onClick={() => setOpen(l)}>
                  <TableCell className='whitespace-nowrap tabular-nums'>
                    {formatDateTime(l.createdAt)}
                  </TableCell>
                  <TableCell>{actorLabel(l)}</TableCell>
                  <TableCell>
                    <button
                      type='button'
                      className='text-left font-medium hover:underline'
                      onClick={(e) => {
                        e.stopPropagation();
                        setOpen(l);
                      }}
                    >
                      {auditActionLabel(l.action)}
                    </button>
                  </TableCell>
                  <TableCell>{auditEntityLabel(l.entity)}</TableCell>
                  <TableCell className='text-muted-foreground'>{l.ip ?? '—'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pager meta={logs.data?.meta} onPage={(page) => void setF({ page })} />
        </>
      )}
      <AuditLogSheet log={open} onClose={() => setOpen(null)} />
    </div>
  );
}
