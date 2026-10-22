import type { TokenPair } from '@/lib/api/types';

export const ACCESS_COOKIE = 'cf_access';
export const REFRESH_COOKIE = 'cf_refresh';

/** Antecedência com que o cookie de acesso expira, para renovar antes do 401. */
const ACCESS_SKEW_SECONDS = 30;
/** Validade do cookie do refresh token quando o MT Login não informa (`refresh_expires_in`). */
const REFRESH_DEFAULT_AGE_SECONDS = 60 * 60 * 24 * 30;
/** Navegadores descartam cookies acima de ~4096 bytes (nome + valor + atributos). */
const MAX_COOKIE_VALUE = 3800;

/** Cookie do login em andamento (state, nonce e PKCE), válido só entre a ida e a volta do MT Login. */
export const LOGIN_FLOW_COOKIE = 'cf_oidc';
const LOGIN_FLOW_MAX_AGE_SECONDS = 600;

interface CookieOptions {
  httpOnly: true;
  sameSite: 'lax';
  secure: boolean;
  path: '/';
  maxAge: number;
}

function base(maxAge: number): CookieOptions {
  return {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge
  };
}

export interface CookieToSet {
  name: string;
  value: string;
  options: CookieOptions;
}

/** O token não cabe num cookie: o navegador o descartaria em silêncio e o login "não pegaria". */
export class TokenTooLargeError extends Error {
  constructor() {
    super('Token do MT Login grande demais para um cookie.');
    this.name = 'TokenTooLargeError';
  }
}

export function sessionCookies(pair: TokenPair): CookieToSet[] {
  if (
    pair.accessToken.length > MAX_COOKIE_VALUE ||
    (pair.refreshToken?.length ?? 0) > MAX_COOKIE_VALUE
  ) {
    throw new TokenTooLargeError();
  }
  return [
    {
      name: ACCESS_COOKIE,
      value: pair.accessToken,
      options: base(Math.max(pair.expiresIn - ACCESS_SKEW_SECONDS, 1))
    },
    pair.refreshToken
      ? {
          name: REFRESH_COOKIE,
          value: pair.refreshToken,
          options: base(
            pair.refreshExpiresIn && pair.refreshExpiresIn > 0
              ? pair.refreshExpiresIn
              : REFRESH_DEFAULT_AGE_SECONDS
          )
        }
      : // Sem refresh token no MT Login: limpa um eventual cookie antigo.
        { name: REFRESH_COOKIE, value: '', options: base(0) }
  ];
}

export function loginFlowCookie(value: string): CookieToSet {
  return { name: LOGIN_FLOW_COOKIE, value, options: base(LOGIN_FLOW_MAX_AGE_SECONDS) };
}

export function clearedLoginFlowCookie(): CookieToSet {
  return { name: LOGIN_FLOW_COOKIE, value: '', options: base(0) };
}

export function clearedSessionCookies(): CookieToSet[] {
  return [ACCESS_COOKIE, REFRESH_COOKIE].map((name) => ({ name, value: '', options: base(0) }));
}

interface CookieWriter {
  set(name: string, value: string, options: CookieOptions): unknown;
}

export function applyCookies(target: CookieWriter, cookies: CookieToSet[]) {
  for (const c of cookies) target.set(c.name, c.value, c.options);
}
