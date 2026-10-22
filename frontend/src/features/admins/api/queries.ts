import { queryOptions } from '@tanstack/react-query';
import { getAdmin, getAdminPermissions, listAdmins } from './service';
import type { AdminListParams } from './types';

export const adminKeys = {
  all: ['admins'] as const,
  list: (params: AdminListParams) => [...adminKeys.all, 'list', params] as const,
  detail: (id: string) => [...adminKeys.all, 'detail', id] as const,
  permissions: (id: string) => [...adminKeys.all, 'permissions', id] as const
};

export const adminsQueryOptions = (params: AdminListParams = {}) =>
  queryOptions({ queryKey: adminKeys.list(params), queryFn: () => listAdmins(params) });

export const adminQueryOptions = (id: string) =>
  queryOptions({ queryKey: adminKeys.detail(id), queryFn: () => getAdmin(id) });

export const adminPermissionsQueryOptions = (id: string) =>
  queryOptions({ queryKey: adminKeys.permissions(id), queryFn: () => getAdminPermissions(id) });
