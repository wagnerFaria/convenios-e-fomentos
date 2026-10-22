import type { TokenPair } from '@/lib/api/types';
import { OidcRejectedError, refreshWithToken } from './oidc';

/**
 * O Keycloak pode rotacionar o refresh token a cada uso e revogar a sessão se um token já usado for
 * reapresentado. Duas requisições simultâneas renovando com o mesmo token derrubariam o usuário,
 * então a renovação é serializada por token e o resultado fica em cache por alguns segundos para
 * atender quem chegou com o cookie antigo. O estado vive em globalThis para ser compartilhado entre
 * o proxy, os route handlers e as server actions do mesmo processo (instâncias múltiplas exigiriam
 * um lock distribuído).
 */
const RESULT_TTL_MS = 10_000;

interface Entry {
  promise: Promise<TokenPair | null>;
  expiresAt: number;
}

const store = globalThis as typeof globalThis & { refreshLocks?: Map<string, Entry> };
const locks = (store.refreshLocks ??= new Map<string, Entry>());

function prune(now: number) {
  for (const [key, entry] of locks) if (entry.expiresAt <= now) locks.delete(key);
}

/**
 * Troca o refresh token por um novo par. Devolve `null` se o MT Login o recusou (sessão
 * encerrada ou expirada); erros de rede ou 5xx são propagados.
 */
export function refreshTokenPair(refreshToken: string): Promise<TokenPair | null> {
  const now = Date.now();
  prune(now);
  const hit = locks.get(refreshToken);
  if (hit) return hit.promise;

  const promise = refreshWithToken(refreshToken).catch((err: unknown) => {
    if (err instanceof OidcRejectedError) return null;
    locks.delete(refreshToken);
    throw err;
  });
  locks.set(refreshToken, { promise, expiresAt: now + RESULT_TTL_MS });
  return promise;
}
