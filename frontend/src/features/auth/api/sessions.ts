import { queryOptions } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import type { AdminSession, ApiResponse } from '@/lib/api/types';

export const sessionKeys = {
  all: ['admin-sessions'] as const
};

export const sessionsQueryOptions = () =>
  queryOptions({
    queryKey: sessionKeys.all,
    queryFn: async () => (await apiClient<ApiResponse<AdminSession[]>>('/admin/auth/sessions')).data
  });

export const revokeSession = (id: string) =>
  apiClient<void>(`/admin/auth/sessions/${id}`, { method: 'DELETE' });

export const revokeOtherSessions = () =>
  apiClient<void>('/admin/auth/sessions?keepCurrent=true', { method: 'DELETE' });
