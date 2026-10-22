import { apiClient } from '@/lib/api-client';
import type { Paginated } from '@/lib/api/types';
import type { AuditListParams, AuditLog } from './types';

export async function listAuditLogs(params: AuditListParams = {}) {
  const qs = new URLSearchParams({ pageSize: String(params.pageSize ?? 25) });
  for (const key of ['action', 'entity', 'entityId', 'actorId', 'from', 'to'] as const) {
    if (params[key]) qs.set(key, params[key]);
  }
  if (params.page) qs.set('page', String(params.page));
  return apiClient<Paginated<AuditLog>>(`/admin/audit-logs?${qs}`);
}
