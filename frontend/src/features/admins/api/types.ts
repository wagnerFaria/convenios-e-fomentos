import type { AdminPermission, AdminRoleRef } from '@/lib/api/types';

/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const adminActions = {
  read: 'admin::settings.users.read',
  create: 'admin::settings.users.create',
  update: 'admin::settings.users.update',
  delete: 'admin::settings.users.delete'
} as const;

/**
 * - `pending`: pré-cadastrado, ainda não fez o primeiro acesso
 * - `awaiting`: já acessou, mas sem nenhuma role (aguardando liberação)
 */
export type AdminStatus = 'active' | 'awaiting' | 'pending' | 'inactive' | 'blocked';

export const adminStatusLabels: Record<AdminStatus, string> = {
  active: 'Ativo',
  awaiting: 'Aguardando liberação',
  pending: 'Pré-cadastrado',
  inactive: 'Inativo',
  blocked: 'Bloqueado'
};

export interface AdminUser {
  id: string;
  firstName: string;
  lastName: string;
  /** Sempre mascarado (***.456.789-**). */
  cpf: string;
  email: string;
  status: AdminStatus;
  isActive: boolean;
  blocked: boolean;
  departmentId: string | null;
  roles: AdminRoleRef[];
  lastLoginAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface AdminListParams {
  q?: string;
  status?: string;
  roleId?: string;
  page?: number;
  pageSize?: number;
}

export interface RegisterAdminInput {
  cpf: string;
  roleIds: string[];
  departmentId?: string | null;
}

export interface AdminPermissions {
  isSuperAdmin: boolean;
  permissions: AdminPermission[];
}
