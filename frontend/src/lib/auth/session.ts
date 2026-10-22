import 'server-only';
import { cache } from 'react';
import { cookies, headers } from 'next/headers';
import { redirect } from 'next/navigation';
import { ApiError } from '@/lib/api/errors';
import type { AdminMe, ApiResponse } from '@/lib/api/types';
import { backendJson } from './backend';
import { ACCESS_COOKIE } from './cookies';

/** Recusas do backend a um token válido do MT Login: a conta não pode usar o painel. */
const DENIED_CODES = new Set(['account_disabled', 'invalid_identity']);

/**
 * Administrador autenticado da requisição atual (uma consulta ao /me por requisição) ou `null`.
 * Na primeira chamada de um CPF novo o backend cria o administrador (sem roles). Conta desativada
 * ou sem CPF no token leva a /auth/denied, que não consulta a sessão (evita laço de redirecionamentos).
 * Cookies só podem ser gravados em server actions, route handlers e no proxy; por isso a
 * renovação do access token expirado acontece no proxy.ts, antes da renderização.
 */
export const getSession = cache(async (): Promise<AdminMe | null> => {
  const token = (await cookies()).get(ACCESS_COOKIE)?.value;
  if (!token) return null;
  try {
    const res = await backendJson<ApiResponse<AdminMe>>('/api/admin/auth/me', {
      token,
      headers: await headers()
    });
    return res.data;
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) return null;
    if (err instanceof ApiError && err.status === 403 && DENIED_CODES.has(err.code)) {
      redirect(`/auth/denied?code=${err.code}`);
    }
    throw err;
  }
});
