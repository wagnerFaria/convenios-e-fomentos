/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const departmentActions = {
  read: 'admin::settings.departments.read',
  create: 'admin::settings.departments.create',
  update: 'admin::settings.departments.update',
  delete: 'admin::settings.departments.delete'
} as const;

export interface Department {
  id: string;
  name: string;
  /** Área finalística: só departamento marcado pode ser indicado para enquadrar um projeto. */
  isFinalistic: boolean;
  userCount: number;
  documentCount: number;
}
