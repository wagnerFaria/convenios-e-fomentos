import type { FieldError } from '@/lib/api/errors';

export type ActionFailure = {
  ok: false;
  code: string;
  message: string;
  fields: FieldError[];
};

export type ActionResult = { ok: true } | ActionFailure;

/** Destino pós-login: só caminhos internos do dashboard (evita open redirect). */
export function safeRedirect(target: string | null | undefined): string {
  if (target && /^\/dashboard(\/|\?|$)/.test(target) && !target.includes('//')) return target;
  return '/dashboard/overview';
}
