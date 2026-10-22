import { apiErrorFromResponse } from '@/lib/api/errors';

/** Todas as chamadas do navegador passam pelo proxy autenticado do Next (cookies httpOnly). */
const BASE_URL = '/bff';

export async function apiClient<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...options.headers
    }
  });

  if (!res.ok) {
    const error = await apiErrorFromResponse(res);
    // Sessão encerrada de vez (o proxy já limpou os cookies): volta ao login.
    if (error.status === 401 && typeof window !== 'undefined') {
      window.location.assign('/auth/sign-in');
    }
    throw error;
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}
