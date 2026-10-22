/** Destino pós-login: só caminhos internos do dashboard (evita open redirect). */
export function safeRedirect(target: string | null | undefined): string {
  if (target && /^\/dashboard(\/|\?|$)/.test(target) && !target.includes('//')) return target;
  return '/dashboard/overview';
}
