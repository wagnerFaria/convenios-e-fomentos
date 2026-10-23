/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const entityTypeActions = {
  read: 'admin::catalog.entity-types.read',
  create: 'admin::catalog.entity-types.create',
  update: 'admin::catalog.entity-types.update',
  delete: 'admin::catalog.entity-types.delete'
} as const;

export type EntityNature = 'public' | 'private';

export const natureLabels: Record<EntityNature, string> = {
  public: 'Pública',
  private: 'Privada'
};

export interface EntityType {
  id: string;
  code: string;
  name: string;
  nature: EntityNature;
  nonProfit: boolean;
  active: boolean;
  sortOrder: number;
  /** O tipo "Outro": o cadastro do projeto pede o texto livre e a natureza. */
  isOther: boolean;
  isSystem: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface EntityTypeInput {
  /** Só na criação; o código não muda depois. */
  code?: string;
  name: string;
  nature: EntityNature;
  nonProfit: boolean;
  active: boolean;
  sortOrder: number;
}
