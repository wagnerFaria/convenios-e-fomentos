import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type { Parliamentarian, ParliamentarianInput } from './types';

const base = '/admin/catalog/parliamentarians';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function listParliamentarians() {
  return (await apiClient<ApiResponse<Parliamentarian[]>>(base)).data;
}

export async function createParliamentarian(input: ParliamentarianInput) {
  return (await apiClient<ApiResponse<Parliamentarian>>(base, { method: 'POST', ...json(input) }))
    .data;
}

export async function updateParliamentarian(id: string, input: ParliamentarianInput) {
  return (
    await apiClient<ApiResponse<Parliamentarian>>(`${base}/${id}`, {
      method: 'PUT',
      ...json(input)
    })
  ).data;
}

export async function deleteParliamentarian(id: string) {
  await apiClient<void>(`${base}/${id}`, { method: 'DELETE' });
}
