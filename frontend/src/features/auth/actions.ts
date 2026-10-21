'use server';

import { cookies, headers } from 'next/headers';
import { redirect } from 'next/navigation';
import { ApiError } from '@/lib/api/errors';
import type { ApiResponse, TokenPair } from '@/lib/api/types';
import { backendJson } from '@/lib/auth/backend';
import {
  ACCESS_COOKIE,
  applyCookies,
  clearedSessionCookies,
  sessionCookies
} from '@/lib/auth/cookies';
import { safeRedirect, type ActionFailure, type ActionResult } from './action-result';

function failure(err: unknown): ActionFailure {
  if (err instanceof ApiError) {
    return { ok: false, code: err.code, message: err.message, fields: err.fieldErrors };
  }
  return {
    ok: false,
    code: 'network_error',
    message: 'Não foi possível contatar o servidor. Tente novamente.',
    fields: []
  };
}

async function startSession(pair: TokenPair) {
  applyCookies(await cookies(), sessionCookies(pair));
}

export async function signIn(input: {
  email: string;
  password: string;
  redirectTo?: string | null;
}): Promise<ActionResult> {
  try {
    const res = await backendJson<ApiResponse<TokenPair>>('/admin/auth/login', {
      method: 'POST',
      body: { email: input.email, password: input.password },
      headers: await headers()
    });
    await startSession(res.data);
  } catch (err) {
    return failure(err);
  }
  redirect(safeRedirect(input.redirectTo));
}

export async function signOut(): Promise<void> {
  const jar = await cookies();
  const token = jar.get(ACCESS_COOKIE)?.value;
  if (token) {
    // Melhor esforço: se o access token já expirou, a sessão cai no vencimento do refresh.
    await backendJson('/admin/auth/logout', {
      method: 'POST',
      token,
      headers: await headers()
    }).catch(() => undefined);
  }
  applyCookies(jar, clearedSessionCookies());
  redirect('/auth/sign-in');
}

export async function registerFirstAdmin(input: {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
}): Promise<ActionResult> {
  try {
    await backendJson('/admin/auth/register-first', {
      method: 'POST',
      body: input,
      headers: await headers()
    });
    const res = await backendJson<ApiResponse<TokenPair>>('/admin/auth/login', {
      method: 'POST',
      body: { email: input.email, password: input.password },
      headers: await headers()
    });
    await startSession(res.data);
  } catch (err) {
    // Configuração inicial já concluída: segue para o login.
    if (err instanceof ApiError && err.code === 'setup_completed') redirect('/auth/sign-in');
    return failure(err);
  }
  redirect('/dashboard/overview');
}

export async function forgotPassword(input: { email: string }): Promise<ActionResult> {
  try {
    await backendJson('/admin/auth/forgot-password', {
      method: 'POST',
      body: input,
      headers: await headers()
    });
    return { ok: true };
  } catch (err) {
    return failure(err);
  }
}

export async function resetPassword(input: {
  token: string;
  password: string;
}): Promise<ActionResult> {
  try {
    await backendJson('/admin/auth/reset-password', {
      method: 'POST',
      body: input,
      headers: await headers()
    });
  } catch (err) {
    return failure(err);
  }
  redirect('/auth/sign-in?reset=1');
}

export async function acceptInvite(input: {
  token: string;
  password: string;
}): Promise<ActionResult> {
  try {
    const res = await backendJson<ApiResponse<TokenPair>>('/admin/auth/accept-invite', {
      method: 'POST',
      body: input,
      headers: await headers()
    });
    await startSession(res.data);
  } catch (err) {
    return failure(err);
  }
  redirect('/dashboard/overview');
}
