import { apiClient } from '@/lib/api-client';
import type { ApiResponse, Paginated } from '@/lib/api/types';
import type {
  EmailTemplate,
  EmailTemplateInput,
  EndUser,
  EndUserCreateInput,
  EndUserListParams,
  EndUserRole,
  EndUserRoleDetail,
  EndUserRoleInput,
  EndUserSettings,
  EndUserUpdateInput
} from './types';

const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });
const data = async <T>(p: Promise<ApiResponse<T>>) => (await p).data;

// ---- usuários finais ----
const users = '/admin/end-users';

export async function listEndUsers(params: EndUserListParams = {}) {
  const qs = new URLSearchParams({ pageSize: '25', sort: 'createdAt:desc' });
  for (const key of ['q', 'roleId', 'confirmed', 'blocked'] as const) {
    if (params[key]) qs.set(key, params[key]);
  }
  if (params.page) qs.set('page', String(params.page));
  return apiClient<Paginated<EndUser>>(`${users}?${qs}`);
}

export const createEndUser = (input: EndUserCreateInput) =>
  data(apiClient<ApiResponse<EndUser>>(users, { method: 'POST', ...json(input) }));

export const updateEndUser = (id: string, input: EndUserUpdateInput) =>
  data(apiClient<ApiResponse<EndUser>>(`${users}/${id}`, { method: 'PUT', ...json(input) }));

export const deleteEndUser = (id: string) =>
  apiClient<void>(`${users}/${id}`, { method: 'DELETE' });

export const setEndUserBlocked = (id: string, blocked: boolean) =>
  data(
    apiClient<ApiResponse<EndUser>>(`${users}/${id}/${blocked ? 'block' : 'unblock'}`, {
      method: 'POST'
    })
  );

export const confirmEndUser = (id: string) =>
  data(apiClient<ApiResponse<EndUser>>(`${users}/${id}/confirm`, { method: 'POST' }));

export const setEndUserRole = (id: string, roleId: string) =>
  data(
    apiClient<ApiResponse<EndUser>>(`${users}/${id}/role`, { method: 'PUT', ...json({ roleId }) })
  );

// ---- perfis públicos ----
const roles = '/admin/end-user-roles';

export const listEndUserRoles = () => data(apiClient<ApiResponse<EndUserRole[]>>(roles));

export const getEndUserRole = (id: string) =>
  data(apiClient<ApiResponse<EndUserRoleDetail>>(`${roles}/${id}`));

export const createEndUserRole = (input: EndUserRoleInput) =>
  data(apiClient<ApiResponse<EndUserRoleDetail>>(roles, { method: 'POST', ...json(input) }));

export const updateEndUserRole = (id: string, input: EndUserRoleInput) =>
  data(
    apiClient<ApiResponse<EndUserRoleDetail>>(`${roles}/${id}`, { method: 'PUT', ...json(input) })
  );

export const deleteEndUserRole = (id: string) =>
  apiClient<void>(`${roles}/${id}`, { method: 'DELETE' });

// ---- configurações ----
const settings = '/admin/end-user-settings';

export const getEndUserSettings = () => data(apiClient<ApiResponse<EndUserSettings>>(settings));

export const updateEndUserSettings = (input: EndUserSettings) =>
  data(apiClient<ApiResponse<EndUserSettings>>(settings, { method: 'PUT', ...json(input) }));

// ---- templates de e-mail ----
const templates = '/admin/email-templates';

export const listEmailTemplates = () => data(apiClient<ApiResponse<EmailTemplate[]>>(templates));

export const updateEmailTemplate = (key: string, input: EmailTemplateInput) =>
  data(
    apiClient<ApiResponse<EmailTemplate>>(`${templates}/${key}`, { method: 'PUT', ...json(input) })
  );
