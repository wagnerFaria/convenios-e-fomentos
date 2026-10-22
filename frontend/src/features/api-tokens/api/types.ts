export const tokenActions = {
  read: 'admin::settings.api-tokens.read',
  create: 'admin::settings.api-tokens.create',
  update: 'admin::settings.api-tokens.update',
  delete: 'admin::settings.api-tokens.delete',
  regenerate: 'admin::settings.api-tokens.regenerate'
} as const;

export type TokenType = 'read-only' | 'full-access' | 'custom';
export type TokenLifespan = '7d' | '30d' | '90d' | 'unlimited';
export type TokenStatus = 'active' | 'expired' | 'revoked';

export const tokenTypeLabels: Record<TokenType, string> = {
  'read-only': 'Somente leitura',
  'full-access': 'Acesso total',
  custom: 'Personalizado'
};

export const tokenTypeDescriptions: Record<TokenType, string> = {
  'read-only': 'Só as ações de leitura (GET) da API pública.',
  'full-access': 'Todas as ações da API pública.',
  custom: 'Somente as ações escolhidas.'
};

export const lifespanLabels: Record<TokenLifespan, string> = {
  '7d': '7 dias',
  '30d': '30 dias',
  '90d': '90 dias',
  unlimited: 'Sem expiração'
};

export const tokenStatusLabels: Record<TokenStatus, string> = {
  active: 'Ativo',
  expired: 'Expirado',
  revoked: 'Revogado'
};

export interface ApiToken {
  id: string;
  name: string;
  description: string;
  type: TokenType;
  /** Ações `api::`; só relevante no tipo personalizado. */
  permissions: string[];
  /** Últimos caracteres, ex.: `gat_…a1b2`. O valor completo nunca é devolvido. */
  hint: string;
  lifespan: TokenLifespan;
  status: TokenStatus;
  expiresAt: string | null;
  lastUsedAt: string | null;
  createdBy: string | null;
  createdAt: string;
  updatedAt: string;
  revokedAt: string | null;
}

/** Devolvido só ao criar ou regenerar: é a única vez em que o valor do token aparece. */
export interface CreatedToken extends ApiToken {
  token: string;
}

export interface TokenInput {
  name: string;
  description: string;
  type: TokenType;
  permissions: string[];
}
