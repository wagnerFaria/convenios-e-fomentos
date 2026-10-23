import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type { RuleParameter, RuleParameterInput, RuleParameterKey } from './types';

const base = '/admin/rule-parameters';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function listRuleParameters() {
  return (await apiClient<ApiResponse<RuleParameter[]>>(base)).data;
}

export async function listRuleParameterKeys() {
  return (await apiClient<ApiResponse<RuleParameterKey[]>>(`${base}/keys`)).data;
}

export async function createRuleParameter(input: RuleParameterInput) {
  return (await apiClient<ApiResponse<RuleParameter>>(base, { method: 'POST', ...json(input) }))
    .data;
}

export async function updateRuleParameter(id: string, input: RuleParameterInput) {
  return (
    await apiClient<ApiResponse<RuleParameter>>(`${base}/${id}`, { method: 'PUT', ...json(input) })
  ).data;
}

export async function deleteRuleParameter(id: string) {
  await apiClient<void>(`${base}/${id}`, { method: 'DELETE' });
}
