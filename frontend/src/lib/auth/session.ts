import 'server-only';
import { cache } from 'react';
import { cookies, headers } from 'next/headers';
import { ApiError } from '@/lib/api/errors';
import type { AdminMe, ApiResponse } from '@/lib/api/types';
import { backendJson } from './backend';
import { ACCESS_COOKIE } from './cookies';

/**
 * Administrador autenticado da requisição atual (uma consulta ao /me por requisição) ou `null`.
 * Cookies só podem ser gravados em server actions, route handlers e no proxy; por isso a
 * renovação do access token expirado acontece no proxy.ts, antes da renderização.
 */
export const getSession = cache(async (): Promise<AdminMe | null> => {
  const token = (await cookies()).get(ACCESS_COOKIE)?.value;
  if (!token) return null;
  try {
    const res = await backendJson<ApiResponse<AdminMe>>('/admin/auth/me', {
      token,
      headers: await headers()
    });
    return res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) return null;
    throw err;
  }
});
