'use client';

import { useQuery } from '@tanstack/react-query';
import { departmentsQueryOptions } from '@/features/departments/api/queries';
import { useSession } from '@/features/auth/session-provider';
import { rolesQueryOptions } from '@/features/roles/api/queries';

export const NO_DEPARTMENT = 'none';

/** Opções de role. Só um Super Admin concede a role Super Admin, então ela fica desabilitada para os demais. */
export function useRoleOptions() {
  const session = useSession();
  const roles = useQuery(rolesQueryOptions());
  const options = (roles.data ?? []).map((r) => ({
    value: r.id,
    label: r.isSuperAdmin ? `${r.name} (acesso total)` : r.name,
    disabled: r.isSuperAdmin && !session.isSuperAdmin
  }));
  return { options, roles: roles.data ?? [], isLoading: roles.isLoading, isError: roles.isError };
}

export function useDepartmentOptions() {
  const departments = useQuery(departmentsQueryOptions());
  const options = [
    { value: NO_DEPARTMENT, label: 'Sem departamento' },
    ...(departments.data ?? []).map((d) => ({ value: d.id, label: d.name }))
  ];
  return { options, departments: departments.data ?? [], isLoading: departments.isLoading };
}

export const toDepartmentId = (value: string) => (value === NO_DEPARTMENT ? null : value);
export const fromDepartmentId = (id: string | null) => id ?? NO_DEPARTMENT;
