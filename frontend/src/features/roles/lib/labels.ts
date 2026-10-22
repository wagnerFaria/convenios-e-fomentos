/** Rótulos em português das categorias do catálogo de ações do backend. */
export const categoryLabels: Record<string, string> = {
  'collection-types': 'Conteúdo',
  'single-types': 'Páginas únicas',
  plugins: 'Plugins',
  settings: 'Configurações',
  application: 'Aplicação'
};

export const categoryLabel = (category: string) => categoryLabels[category] ?? category;

/** Ordem de exibição das categorias. */
export const categoryOrder = [
  'collection-types',
  'single-types',
  'application',
  'plugins',
  'settings'
];

export const sortCategories = <T extends { category: string }>(groups: T[]): T[] =>
  groups.toSorted(
    (a, b) =>
      (categoryOrder.indexOf(a.category) + 1 || 99) - (categoryOrder.indexOf(b.category) + 1 || 99)
  );

/** Rótulos dos recursos do catálogo; os demais aparecem pelo identificador. */
export const resourceLabels: Record<string, string> = {
  document: 'Documentos',
  'settings.users': 'Administradores',
  'settings.roles': 'Perfis de administrador',
  'settings.departments': 'Departamentos',
  'settings.api-tokens': 'Tokens de API',
  'settings.email-templates': 'Templates de e-mail',
  'settings.audit-logs': 'Auditoria',
  'plugin.enduser-users': 'Usuários finais',
  'plugin.enduser-roles': 'Perfis de usuário final',
  'plugin.enduser-settings': 'Configurações de usuários finais'
};

export const resourceLabel = (resource: string) => resourceLabels[resource] ?? resource;
