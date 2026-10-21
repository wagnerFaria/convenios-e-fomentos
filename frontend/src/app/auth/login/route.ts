import { NextResponse, type NextRequest } from 'next/server';
import { applyCookies, loginFlowCookie } from '@/lib/auth/cookies';
import { OidcConfigError, appOriginOr, startLogin } from '@/lib/auth/oidc';
import { safeRedirect } from '@/features/auth/action-result';

/**
 * Inicia o login: gera state, nonce e PKCE, guarda-os num cookie curto e envia o usuário ao MT Login.
 * É um route handler (GET) em vez de server action para ser um simples link na tela de entrada.
 */
export async function GET(req: NextRequest) {
  const target = req.nextUrl.searchParams.get('redirect');
  try {
    const login = await startLogin(target ? safeRedirect(target) : null);
    const res = NextResponse.redirect(login.url);
    applyCookies(res.cookies, [loginFlowCookie(JSON.stringify(login.flow))]);
    res.headers.set('Cache-Control', 'no-store');
    return res;
  } catch (err) {
    const code = err instanceof OidcConfigError ? 'config' : 'unavailable';
    return NextResponse.redirect(
      new URL(`/auth/sign-in?error=${code}`, appOriginOr(req.nextUrl.origin))
    );
  }
}
