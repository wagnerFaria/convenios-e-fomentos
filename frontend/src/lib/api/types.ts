/** Contratos do backend (backend/api/openapi.yaml) usados pelo frontend. */

export interface ApiMeta {
  page: number;
  pageSize: number;
  total: number;
  pageCount: number;
}

export interface ApiResponse<T> {
  data: T;
  meta?: ApiMeta;
}

export interface TokenPair {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  /** Segundos até o access token expirar. */
  expiresIn: number;
}

export interface AdminRoleRef {
  id: string;
  code: string;
  name: string;
}

export interface AdminPermission {
  action: string;
  subject: string;
  /** `null` significa todos os campos. */
  fields: string[] | null;
  /** Vazio significa sem condição; várias condições combinam com OR. */
  conditions: string[];
}

export interface AdminMe {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  departmentId: string | null;
  isSuperAdmin: boolean;
  roles: AdminRoleRef[];
  permissions: AdminPermission[];
}

export interface AdminSession {
  id: string;
  userAgent: string;
  ip: string;
  createdAt: string;
  lastUsedAt: string;
  expiresAt: string;
  current: boolean;
}
