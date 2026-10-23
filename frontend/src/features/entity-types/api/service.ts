import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type { EntityType, EntityTypeInput } from './types';

const base = '/admin/catalog/entity-types';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function listEntityTypes() {
  return (await apiClient<ApiResponse<EntityType[]>>(base)).data;
}

export async function createEntityType(input: EntityTypeInput) {
  return (await apiClient<ApiResponse<EntityType>>(base, { method: 'POST', ...json(input) })).data;
}

export async function updateEntityType(id: string, input: EntityTypeInput) {
  return (
    await apiClient<ApiResponse<EntityType>>(`${base}/${id}`, { method: 'PUT', ...json(input) })
  ).data;
}

export async function deleteEntityType(id: string) {
  await apiClient<void>(`${base}/${id}`, { method: 'DELETE' });
}
