/** Rótulos em português das ações auditadas pelo backend; as desconhecidas aparecem pelo identificador. */
export const auditActionLabels: Record<string, string> = {
  'admin.auto-provisioned': 'Primeiro acesso de administrador',
  'admin.ensure-super-admin': 'Super Admin garantido (CLI)',
  'admin.create-super-admin': 'Super Admin criado (CLI)',
  'admin.register': 'Administrador pré-cadastrado',
  'admin.update': 'Administrador editado',
  'admin.activate': 'Administrador ativado',
  'admin.deactivate': 'Administrador desativado',
  'admin.block': 'Administrador bloqueado',
  'admin.unblock': 'Administrador desbloqueado',
  'admin.delete': 'Administrador excluído',
  'admin.roles-add': 'Perfis adicionados ao administrador',
  'admin.roles-remove': 'Perfil removido do administrador',
  'admin.roles-set': 'Perfis do administrador definidos',
  'admin-role.create': 'Perfil criado',
  'admin-role.update': 'Perfil editado',
  'admin-role.duplicate': 'Perfil duplicado',
  'admin-role.delete': 'Perfil excluído',
  'admin-role.permissions-replace': 'Permissões do perfil alteradas',
  'department.create': 'Departamento criado',
  'department.update': 'Departamento renomeado',
  'department.delete': 'Departamento excluído',
  'api-token.create': 'Token de API criado',
  'api-token.update': 'Token de API editado',
  'api-token.regenerate': 'Token de API regenerado',
  'api-token.revoke': 'Token de API revogado',
  'email-template.update': 'Template de e-mail editado',
  'end-user.register': 'Usuário final cadastrado',
  'end-user.create': 'Usuário final criado',
  'end-user.update': 'Usuário final editado',
  'end-user.delete': 'Usuário final excluído',
  'end-user.block': 'Usuário final bloqueado',
  'end-user.unblock': 'Usuário final desbloqueado',
  'end-user.confirm': 'E-mail do usuário final confirmado',
  'end-user.email-confirmed': 'E-mail confirmado pelo usuário',
  'end-user.role-change': 'Perfil do usuário final alterado',
  'end-user.password-change': 'Senha alterada',
  'end-user.password-reset': 'Senha redefinida',
  'end-user-role.create': 'Perfil público criado',
  'end-user-role.update': 'Perfil público editado',
  'end-user-role.delete': 'Perfil público excluído',
  'end-user-settings.update': 'Configurações de usuários finais alteradas',
  'auth.refresh-reuse': 'Reuso de token de renovação detectado'
};

export const auditActionLabel = (action: string) => auditActionLabels[action] ?? action;

export const auditEntityLabels: Record<string, string> = {
  admin: 'Administrador',
  admin_role: 'Perfil',
  department: 'Departamento',
  api_token: 'Token de API',
  document: 'Documento',
  email_template: 'Template de e-mail',
  end_user: 'Usuário final',
  end_user_role: 'Perfil público',
  end_user_session: 'Sessão de usuário final',
  settings: 'Configurações'
};

export const auditEntityLabel = (entity: string) => auditEntityLabels[entity] ?? entity;
