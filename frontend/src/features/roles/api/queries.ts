import { queryOptions } from '@tanstack/react-query';
import { getCatalogActions, getCatalogConditions, getRole, listRoles } from './service';
import type { CatalogDomain } from './types';

export const roleKeys = {
  all: ['roles'] as const,
  list: () => [...roleKeys.all, 'list'] as const,
  detail: (id: string) => [...roleKeys.all, 'detail', id] as const,
  catalog: (domain: CatalogDomain) => [...roleKeys.all, 'catalog', domain] as const,
  conditions: () => [...roleKeys.all, 'conditions'] as const
};

export const rolesQueryOptions = () =>
  queryOptions({ queryKey: roleKeys.list(), queryFn: listRoles });

export const roleQueryOptions = (id: string) =>
  queryOptions({ queryKey: roleKeys.detail(id), queryFn: () => getRole(id) });

/** O catálogo só muda com um novo deploy, então fica em cache por mais tempo. */
export const catalogActionsQueryOptions = (domain: CatalogDomain) =>
  queryOptions({
    queryKey: roleKeys.catalog(domain),
    queryFn: () => getCatalogActions(domain),
    staleTime: 10 * 60_000
  });

export const catalogConditionsQueryOptions = () =>
  queryOptions({
    queryKey: roleKeys.conditions(),
    queryFn: getCatalogConditions,
    staleTime: 10 * 60_000
  });
