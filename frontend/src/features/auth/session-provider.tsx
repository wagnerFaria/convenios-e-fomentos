'use client';

import { createContext, useCallback, useContext, useMemo } from 'react';
import type { AdminMe } from '@/lib/api/types';

const SessionContext = createContext<AdminMe | null>(null);

export function SessionProvider({
  session,
  children
}: {
  session: AdminMe;
  children: React.ReactNode;
}) {
  return <SessionContext.Provider value={session}>{children}</SessionContext.Provider>;
}

export function useSession(): AdminMe {
  const session = useContext(SessionContext);
  if (!session) throw new Error('useSession deve ser usado dentro de <SessionProvider>.');
  return session;
}

/**
 * Verifica se o administrador possui a ação (ex.: `admin::content.document.update`).
 * Serve apenas para esconder elementos de UI: a autorização real é sempre do backend.
 */
export function useCan() {
  const session = useSession();
  const actions = useMemo(
    () => new Set(session.permissions.map((p) => p.action)),
    [session.permissions]
  );
  return useCallback(
    (action: string) => session.isSuperAdmin || actions.has(action),
    [session.isSuperAdmin, actions]
  );
}

/** Renderiza os filhos somente se o administrador tiver a permissão. */
export function Can({
  action,
  children,
  fallback = null
}: {
  action: string;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  const can = useCan();
  return can(action) ? children : fallback;
}
