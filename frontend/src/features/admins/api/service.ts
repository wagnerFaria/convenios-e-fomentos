import { apiClient } from '@/lib/api-client';
import type { ApiResponse, Paginated } from '@/lib/api/types';
import type { AdminListParams, AdminPermissions, AdminUser, RegisterAdminInput } from './types';

const base = '/admin/users';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });
const post = (path: string) => apiClient<ApiResponse<AdminUser>>(path, { method: 'POST' });

export async function listAdmins(params: AdminListParams = {}) {
  const qs = new URLSearchParams({
    pageSize: String(params.pageSize ?? 25),
    sort: 'createdAt:desc'
  });
  if (params.q) qs.set('q', params.q);
  if (params.status) qs.set('status', params.status);
  if (params.roleId) qs.set('roleId', params.roleId);
  if (params.page) qs.set('page', String(params.page));
  return apiClient<Paginated<AdminUser>>(`${base}?${qs}`);
}

export async function getAdmin(id: string) {
  return (await apiClient<ApiResponse<AdminUser>>(`${base}/${id}`)).data;
}

/** Pré-cadastro pelo CPF: nome e e-mail vêm do MT Login no primeiro acesso. */
export async function registerAdmin(input: RegisterAdminInput) {
  return (await apiClient<ApiResponse<AdminUser>>(base, { method: 'POST', ...json(input) })).data;
}

export async function updateAdmin(id: string, input: { departmentId: string | null }) {
  return (
    await apiClient<ApiResponse<AdminUser>>(`${base}/${id}`, { method: 'PUT', ...json(input) })
  ).data;
}

export async function setAdminActive(id: string, active: boolean) {
  return (await post(`${base}/${id}/${active ? 'activate' : 'deactivate'}`)).data;
}

export async function setAdminBlocked(id: string, blocked: boolean) {
  return (await post(`${base}/${id}/${blocked ? 'block' : 'unblock'}`)).data;
}

/** Substitui o conjunto de roles (libera o acesso de quem aguarda). */
export async function setAdminRoles(id: string, roleIds: string[]) {
  return (
    await apiClient<ApiResponse<AdminUser>>(`${base}/${id}/roles`, {
      method: 'PUT',
      ...json({ roleIds })
    })
  ).data;
}

export async function deleteAdmins(ids: string[]) {
  await apiClient<void>(`${base}/batch-delete`, { method: 'POST', ...json({ ids }) });
}

export async function getAdminPermissions(id: string) {
  return (await apiClient<ApiResponse<AdminPermissions>>(`${base}/${id}/permissions`)).data;
}
