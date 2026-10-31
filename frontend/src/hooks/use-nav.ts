'use client';

/**
 * Filtra a navegação pelas permissões efetivas do administrador (`/api/admin/auth/me`).
 * Afeta apenas a visibilidade: a autorização real é feita pelo backend em cada requisição.
 */

import { useMemo } from 'react';
import { useCan, useSession } from '@/features/auth/session-provider';
import { isProjectProponent, isProjectStaff } from '@/features/projects/access';
import type { AdminMe } from '@/lib/api/types';
import type { NavItem, NavGroup } from '@/types';

/** Aplica `access.permission` e `access.when` de um item (os dois, se houver). */
function isAllowed(item: NavItem, can: (action: string) => boolean, session: AdminMe) {
  const { permission, when } = item.access ?? {};
  if (permission && !can(permission)) return false;
  if (when === 'project-staff') return isProjectStaff(session);
  if (when === 'project-proponent') return isProjectProponent(session);
  return true;
}

export function useFilteredNavItems(items: NavItem[]) {
  const can = useCan();
  const session = useSession();

  return useMemo(() => {
    const allowed = (item: NavItem) => isAllowed(item, can, session);
    return items
      .filter(allowed)
      .map((item) => (item.items?.length ? { ...item, items: item.items.filter(allowed) } : item));
  }, [items, can, session]);
}

export function useFilteredNavGroups(groups: NavGroup[]) {
  const can = useCan();
  const session = useSession();

  return useMemo(() => {
    const allowed = (item: NavItem) => isAllowed(item, can, session);
    return groups
      .map((group) => ({
        ...group,
        items: group.items
          .filter(allowed)
          .map((item) =>
            item.items?.length ? { ...item, items: item.items.filter(allowed) } : item
          )
      }))
      .filter((group) => group.items.length > 0);
  }, [groups, can, session]);
}
