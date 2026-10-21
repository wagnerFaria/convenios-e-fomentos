import { NextResponse, type NextRequest } from 'next/server';
import {
  LOGIN_FLOW_COOKIE,
  TokenTooLargeError,
  applyCookies,
  clearedLoginFlowCookie,
  sessionCookies
} from '@/lib/auth/cookies';
import { OidcRejectedError, appOriginOr, exchangeCode, type LoginFlow } from '@/lib/auth/oidc';
import { safeRedirect } from '@/features/auth/action-result';

function parseFlow(raw: string | undefined): LoginFlow | null {
  if (!raw) return null;
  try {
    const flow = JSON.parse(raw) as Partial<LoginFlow>;
    if (flow.state && flow.nonce && flow.verifier) return flow as LoginFlow;
  } catch {
    // cookie corrompido: tratado como login expirado
  }
  return null;
}

/**
 * Volta do MT Login: confere o state, troca o `code` pelos tokens (PKCE) e abre a sessão gravando
 * os tokens em cookies httpOnly. O administrador é criado/atualizado no backend na primeira chamada
 * autenticada (/admin/auth/me), feita pelo layout do dashboard.
 */
export async function GET(req: NextRequest) {
  const origin = appOriginOr(req.nextUrl.origin);
  const toSignIn = (error: string) => {
    const res = NextResponse.redirect(new URL(`/auth/sign-in?error=${error}`, origin));
    applyCookies(res.cookies, [clearedLoginFlowCookie()]);
    return res;
  };

  const params = req.nextUrl.searchParams;
  const flow = parseFlow(req.cookies.get(LOGIN_FLOW_COOKIE)?.value);
  if (params.get('error')) {
    return toSignIn(params.get('error') === 'access_denied' ? 'access_denied' : 'provider_error');
  }
  const code = params.get('code');
  if (!flow || !code || params.get('state') !== flow.state) return toSignIn('expired');

  try {
    const pair = await exchangeCode(code, flow);
    const res = NextResponse.redirect(new URL(safeRedirect(flow.redirect), origin));
    applyCookies(res.cookies, [...sessionCookies(pair), clearedLoginFlowCookie()]);
    res.headers.set('Cache-Control', 'no-store');
    return res;
  } catch (err) {
    if (err instanceof TokenTooLargeError) return toSignIn('token_too_large');
    if (err instanceof OidcRejectedError) return toSignIn('expired');
    return toSignIn('unavailable');
  }
}
