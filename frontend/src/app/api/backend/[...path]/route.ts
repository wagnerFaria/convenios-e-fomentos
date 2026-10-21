import { type NextRequest, NextResponse } from 'next/server';
import { backendUrl, forwardedHeaders } from '@/lib/auth/backend';
import {
  ACCESS_COOKIE,
  REFRESH_COOKIE,
  applyCookies,
  clearedSessionCookies,
  sessionCookies
} from '@/lib/auth/cookies';
import { refreshTokenPair } from '@/lib/auth/refresh';
import type { TokenPair } from '@/lib/api/types';

/**
 * Proxy autenticado para o backend (padrão BFF): o navegador só fala com o Next, que injeta o
 * Bearer (access token do MT Login) a partir do cookie httpOnly e renova o par de tokens quando
 * necessário.
 */

const FORWARDED_RESPONSE_HEADERS = ['content-type', 'retry-after', 'x-request-id'];

function errorResponse(status: number, code: string, message: string) {
  return NextResponse.json({ error: { status, code, message } }, { status });
}

function targetPath(segments: string[]): string | null {
  if (segments.some((s) => s === '' || s === '.' || s === '..' || s.includes('/'))) return null;
  const path = segments.join('/');
  if (!path.startsWith('admin/')) return null;
  return `/${path}`;
}

function sameOrigin(req: NextRequest): boolean {
  const origin = req.headers.get('origin');
  return !origin || origin === req.nextUrl.origin;
}

async function handle(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const { path } = await ctx.params;
  const target = targetPath(path);
  if (!target) return errorResponse(404, 'not_found', 'Recurso não encontrado.');
  if (req.method !== 'GET' && !sameOrigin(req)) {
    return errorResponse(403, 'forbidden', 'Origem não permitida.');
  }

  let access = req.cookies.get(ACCESS_COOKIE)?.value;
  const refresh = req.cookies.get(REFRESH_COOKIE)?.value;
  let renewed: TokenPair | null = null;
  const body = req.method === 'GET' || req.method === 'HEAD' ? undefined : await req.arrayBuffer();

  const call = (token: string | undefined) =>
    fetch(backendUrl(target + req.nextUrl.search), {
      method: req.method,
      cache: 'no-store',
      body,
      headers: {
        ...(req.headers.get('content-type')
          ? { 'Content-Type': req.headers.get('content-type')! }
          : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...forwardedHeaders(req.headers)
      }
    });

  const renew = async () => {
    if (!refresh) return false;
    try {
      renewed = await refreshTokenPair(refresh);
    } catch {
      return false;
    }
    if (!renewed) return false;
    access = renewed.accessToken;
    return true;
  };

  let loggedOut = false;
  if (!access && !(await renew())) loggedOut = true;

  let upstream: Response | null = null;
  if (!loggedOut) {
    try {
      upstream = await call(access);
      if (upstream.status === 401 && !renewed && (await renew())) upstream = await call(access);
    } catch {
      return errorResponse(502, 'bad_gateway', 'Não foi possível contatar o servidor.');
    }
    if (upstream.status === 401 && (!refresh || !renewed)) loggedOut = true;
  }

  const res =
    upstream && !loggedOut
      ? new NextResponse(upstream.status === 204 ? null : upstream.body, {
          status: upstream.status
        })
      : errorResponse(401, 'unauthorized', 'Sessão expirada.');

  if (upstream && !loggedOut) {
    for (const h of FORWARDED_RESPONSE_HEADERS) {
      const v = upstream.headers.get(h);
      if (v) res.headers.set(h, v);
    }
  }
  if (loggedOut) applyCookies(res.cookies, clearedSessionCookies());
  else if (renewed) applyCookies(res.cookies, sessionCookies(renewed));
  res.headers.set('Cache-Control', 'no-store');
  return res;
}

export { handle as GET, handle as POST, handle as PUT, handle as PATCH, handle as DELETE };
