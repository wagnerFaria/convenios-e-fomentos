import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type { Department } from './types';

const base = '/admin/departments';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

/** `finalistic: true` devolve só as áreas finalísticas (as que podem ser indicadas para enquadrar um projeto). */
export async function listDepartments({ finalistic }: { finalistic?: boolean } = {}) {
  const query = finalistic ? '?finalistic=true' : '';
  return (await apiClient<ApiResponse<Department[]>>(`${base}${query}`)).data;
}

export interface DepartmentInput {
  name: string;
  isFinalistic: boolean;
}

export async function createDepartment(input: DepartmentInput) {
  return (await apiClient<ApiResponse<Department>>(base, { method: 'POST', ...json(input) })).data;
}

export async function updateDepartment(id: string, input: DepartmentInput) {
  return (
    await apiClient<ApiResponse<Department>>(`${base}/${id}`, { method: 'PUT', ...json(input) })
  ).data;
}

export async function deleteDepartment(id: string) {
  await apiClient<void>(`${base}/${id}`, { method: 'DELETE' });
}
