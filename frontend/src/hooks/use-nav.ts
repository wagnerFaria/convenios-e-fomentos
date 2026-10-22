'use client';

/**
 * Filtra a navegação pelas permissões efetivas do administrador (`/api/admin/auth/me`).
 * Afeta apenas a visibilidade: a autorização real é feita pelo backend em cada requisição.
 */

import { useMemo } from 'react';
import { useCan } from '@/features/auth/session-provider';
import type { NavItem, NavGroup } from '@/types';

export function useFilteredNavItems(items: NavItem[]) {
  const can = useCan();

  return useMemo(() => {
    const allowed = (item: NavItem) => !item.access?.permission || can(item.access.permission);
    return items
      .filter(allowed)
      .map((item) => (item.items?.length ? { ...item, items: item.items.filter(allowed) } : item));
  }, [items, can]);
}

export function useFilteredNavGroups(groups: NavGroup[]) {
  const can = useCan();

  return useMemo(() => {
    const allowed = (item: NavItem) => !item.access?.permission || can(item.access.permission);
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
  }, [groups, can]);
}
