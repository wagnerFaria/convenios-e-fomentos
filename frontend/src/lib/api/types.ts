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

/** Tokens do MT Login (Keycloak). */
export interface TokenPair {
  accessToken: string;
  /** Ausente se o MT Login não devolver refresh token: ao expirar o access token, novo login (o SSO costuma ser transparente). */
  refreshToken?: string;
  tokenType: string;
  /** Segundos até o access token expirar. */
  expiresIn: number;
  /** Segundos até o refresh token expirar (a sessão SSO), quando informado. */
  refreshExpiresIn?: number;
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
  /** CPF completo e formatado, só do próprio usuário. */
  cpf: string;
  /** Nome, CPF e e-mail vêm do MT Login. */
  email: string;
  firstName: string;
  lastName: string;
  departmentId: string | null;
  isSuperAdmin: boolean;
  roles: AdminRoleRef[];
  permissions: AdminPermission[];
}
