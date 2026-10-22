import { apiErrorFromResponse } from '@/lib/api/errors';

export function backendUrl(path: string): string {
  const base = process.env.API_URL ?? 'http://localhost:8080';
  return `${base.replace(/\/$/, '')}${path}`;
}

/** Cabeçalhos repassados ao backend para que o rate limit enxergue o IP real. */
export function forwardedHeaders(incoming: Headers): Record<string, string> {
  const out: Record<string, string> = {};
  const xff = incoming.get('x-forwarded-for');
  if (xff) out['X-Forwarded-For'] = xff;
  // O backend grava o user agent da sessão; sem isto apareceria o do servidor do Next.
  const ua = incoming.get('user-agent');
  if (ua) out['User-Agent'] = ua;
  const reqId = incoming.get('x-request-id');
  if (reqId) out['X-Request-ID'] = reqId;
  return out;
}

/** Chamada JSON ao backend; lança ApiError em respostas não-2xx. */
export async function backendJson<T>(
  path: string,
  init: { method?: string; body?: unknown; token?: string; headers?: Headers } = {}
): Promise<T> {
  const res = await fetch(backendUrl(path), {
    method: init.method ?? 'GET',
    cache: 'no-store',
    headers: {
      ...(init.body !== undefined ? { 'Content-Type': 'application/json' } : {}),
      ...(init.token ? { Authorization: `Bearer ${init.token}` } : {}),
      ...(init.headers ? forwardedHeaders(init.headers) : {})
    },
    body: init.body !== undefined ? JSON.stringify(init.body) : undefined
  });
  if (!res.ok) throw await apiErrorFromResponse(res);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}
