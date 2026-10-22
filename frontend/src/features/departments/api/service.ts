import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type { Department } from './types';

const base = '/admin/departments';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function listDepartments() {
  return (await apiClient<ApiResponse<Department[]>>(base)).data;
}

export async function createDepartment(input: { name: string }) {
  return (await apiClient<ApiResponse<Department>>(base, { method: 'POST', ...json(input) })).data;
}

export async function updateDepartment(id: string, input: { name: string }) {
  return (
    await apiClient<ApiResponse<Department>>(`${base}/${id}`, { method: 'PUT', ...json(input) })
  ).data;
}

export async function deleteDepartment(id: string) {
  await apiClient<void>(`${base}/${id}`, { method: 'DELETE' });
}
