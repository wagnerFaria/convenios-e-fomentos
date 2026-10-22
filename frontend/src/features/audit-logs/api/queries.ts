import { queryOptions } from '@tanstack/react-query';
import { listAuditLogs } from './service';
import type { AuditListParams } from './types';

export const auditKeys = {
  all: ['audit-logs'] as const,
  list: (params: AuditListParams) => [...auditKeys.all, 'list', params] as const
};

export const auditLogsQueryOptions = (params: AuditListParams = {}) =>
  queryOptions({ queryKey: auditKeys.list(params), queryFn: () => listAuditLogs(params) });
