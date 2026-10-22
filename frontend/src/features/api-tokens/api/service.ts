import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type { ApiToken, CreatedToken, TokenInput, TokenLifespan } from './types';

const base = '/admin/api-tokens';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function listTokens(includeRevoked: boolean) {
  const qs = includeRevoked ? '?includeRevoked=true' : '';
  return (await apiClient<ApiResponse<ApiToken[]>>(`${base}${qs}`)).data;
}

export async function createToken(input: TokenInput & { lifespan: TokenLifespan }) {
  return (await apiClient<ApiResponse<CreatedToken>>(base, { method: 'POST', ...json(input) }))
    .data;
}

export async function updateToken(id: string, input: TokenInput) {
  return (
    await apiClient<ApiResponse<ApiToken>>(`${base}/${id}`, { method: 'PUT', ...json(input) })
  ).data;
}

export async function regenerateToken(id: string) {
  return (
    await apiClient<ApiResponse<CreatedToken>>(`${base}/${id}/regenerate`, { method: 'POST' })
  ).data;
}

export async function revokeToken(id: string) {
  await apiClient<void>(`${base}/${id}`, { method: 'DELETE' });
}
