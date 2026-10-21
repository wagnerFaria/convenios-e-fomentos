import { type NextRequest, NextResponse } from 'next/server';
import {
  ACCESS_COOKIE,
  REFRESH_COOKIE,
  applyCookies,
  clearedSessionCookies,
  sessionCookies
} from '@/lib/auth/cookies';
import { refreshTokenPair } from '@/lib/auth/refresh';

function toSignIn(req: NextRequest, clear: boolean) {
  const url = req.nextUrl.clone();
  url.pathname = '/auth/sign-in';
  url.search = '';
  const target = req.nextUrl.pathname + req.nextUrl.search;
  if (target !== '/dashboard') url.searchParams.set('redirect', target);
  const res = NextResponse.redirect(url);
  if (clear) applyCookies(res.cookies, clearedSessionCookies());
  return res;
}

/**
 * Protege /dashboard: sem sessão redireciona ao login e, com o access token expirado (o cookie
 * some junto), renova o par de tokens antes da renderização — único ponto em que uma navegação
 * consegue gravar cookies. A autorização de fato é do backend.
 */
export default async function proxy(req: NextRequest) {
  const access = req.cookies.get(ACCESS_COOKIE)?.value;
  if (access) return NextResponse.next();

  const refresh = req.cookies.get(REFRESH_COOKIE)?.value;
  if (!refresh) return toSignIn(req, false);

  let pair;
  try {
    pair = await refreshTokenPair(refresh);
  } catch {
    return new NextResponse('Serviço indisponível. Tente novamente.', { status: 503 });
  }
  if (!pair) return toSignIn(req, true);

  const cookies = sessionCookies(pair);
  // O novo access token precisa chegar às server components desta mesma requisição.
  for (const c of cookies) req.cookies.set(c.name, c.value);
  const res = NextResponse.next({ request: { headers: req.headers } });
  applyCookies(res.cookies, cookies);
  return res;
}

export const config = {
  matcher: ['/dashboard/:path*']
};
