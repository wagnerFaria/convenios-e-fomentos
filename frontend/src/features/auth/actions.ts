'use server';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { applyCookies, clearedSessionCookies } from '@/lib/auth/cookies';
import { OidcConfigError, endSessionUrl } from '@/lib/auth/oidc';

/**
 * Encerra a sessão: limpa os cookies e leva o usuário ao logout do MT Login, que encerra também o
 * SSO (sem isso, "Entrar" logaria de novo sozinho) e volta para a tela de entrada.
 */
export async function signOut(): Promise<void> {
  applyCookies(await cookies(), clearedSessionCookies());
  let target = '/auth/sign-in';
  try {
    target = endSessionUrl();
  } catch (err) {
    if (!(err instanceof OidcConfigError)) throw err;
  }
  redirect(target);
}
