import type { TokenPair } from '@/lib/api/types';

export const ACCESS_COOKIE = 'cf_access';
export const REFRESH_COOKIE = 'cf_refresh';

/** Antecedência com que o cookie de acesso expira, para renovar antes do 401. */
const ACCESS_SKEW_SECONDS = 30;
/** Teto do refresh token no backend (REFRESH_TOKEN_MAX_TTL=720h). */
const REFRESH_MAX_AGE_SECONDS = 60 * 60 * 24 * 30;

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

export function sessionCookies(pair: TokenPair): CookieToSet[] {
  return [
    {
      name: ACCESS_COOKIE,
      value: pair.accessToken,
      options: base(Math.max(pair.expiresIn - ACCESS_SKEW_SECONDS, 1))
    },
    { name: REFRESH_COOKIE, value: pair.refreshToken, options: base(REFRESH_MAX_AGE_SECONDS) }
  ];
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
