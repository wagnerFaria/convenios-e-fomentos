/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const endUserActions = {
  users: {
    read: 'admin::plugin.enduser-users.read',
    create: 'admin::plugin.enduser-users.create',
    update: 'admin::plugin.enduser-users.update',
    delete: 'admin::plugin.enduser-users.delete'
  },
  roles: {
    read: 'admin::plugin.enduser-roles.read',
    create: 'admin::plugin.enduser-roles.create',
    update: 'admin::plugin.enduser-roles.update',
    delete: 'admin::plugin.enduser-roles.delete'
  },
  settings: {
    read: 'admin::plugin.enduser-settings.read',
    update: 'admin::plugin.enduser-settings.update'
  },
  templates: {
    read: 'admin::settings.email-templates.read',
    update: 'admin::settings.email-templates.update'
  }
} as const;

export interface EndUserRoleRef {
  id: string;
  /** `public`, `authenticated` ou `custom-<slug>`. */
  type: string;
  name: string;
}

export interface EndUser {
  id: string;
  username: string;
  email: string;
  confirmed: boolean;
  blocked: boolean;
  provider: string;
  role: EndUserRoleRef;
  profile: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface EndUserListParams {
  q?: string;
  roleId?: string;
  confirmed?: string;
  blocked?: string;
  page?: number;
}

export interface EndUserCreateInput {
  username: string;
  email: string;
  password: string;
  roleId?: string;
  confirmed: boolean;
  blocked: boolean;
}

export interface EndUserUpdateInput {
  username: string;
  email: string;
  /** Opcional: redefine a senha e revoga as sessões. */
  password?: string;
}

export interface EndUserRole {
  id: string;
  type: string;
  name: string;
  description: string;
  isSystem: boolean;
  userCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface EndUserPermission {
  action: string;
  name: string;
  method: string;
  route: string;
}

export interface EndUserRoleDetail extends EndUserRole {
  permissions: EndUserPermission[];
}

export interface EndUserRoleInput {
  name: string;
  description: string;
  /** Ações `api::`; ausente na edição = não altera. */
  permissions?: string[];
}

export interface EndUserSettings {
  defaultRoleId: string;
  oneAccountPerEmail: boolean;
  registrationEnabled: boolean;
  emailConfirmationRequired: boolean;
  resetPasswordUrl: string;
  emailConfirmationRedirectUrl: string;
  registerAllowedFields: string[];
}

export interface EmailTemplate {
  key: string;
  senderName: string;
  senderEmail: string;
  replyTo: string;
  subject: string;
  bodyHtml: string;
  updatedAt: string;
}

export type EmailTemplateInput = Omit<EmailTemplate, 'key' | 'updatedAt'>;

export const templateLabels: Record<string, string> = {
  'email-confirmation': 'Confirmação de e-mail',
  'reset-password': 'Redefinição de senha'
};
