import { NavGroup } from '@/types';

/**
 * Configuração da navegação (sidebar e Cmd+K).
 *
 * Cada item pode exigir uma ação do backend em `access.permission`; o Super Admin sempre vê tudo.
 * Ex.: `access: { permission: 'admin::settings.users.read' }`.
 */
export const navGroups: NavGroup[] = [
  {
    label: 'Geral',
    items: [
      {
        title: 'Visão geral',
        url: '/dashboard/overview',
        icon: 'dashboard',
        isActive: false,
        shortcut: ['d', 'd'],
        items: []
      }
    ]
  },
  {
    label: 'Conteúdo',
    items: [
      {
        title: 'Documentos',
        url: '/dashboard/documentos',
        icon: 'post',
        isActive: false,
        shortcut: ['d', 'o'],
        access: { permission: 'admin::content.document.read' },
        items: []
      }
    ]
  },
  {
    label: 'Administração',
    items: [
      {
        title: 'Usuários',
        url: '/dashboard/administracao/usuarios',
        icon: 'teams',
        isActive: false,
        shortcut: ['a', 'u'],
        access: { permission: 'admin::settings.users.read' },
        items: []
      },
      {
        title: 'Perfis e permissões',
        url: '/dashboard/administracao/perfis',
        icon: 'shield',
        isActive: false,
        shortcut: ['a', 'p'],
        access: { permission: 'admin::settings.roles.read' },
        items: []
      },
      {
        title: 'Departamentos',
        url: '/dashboard/administracao/departamentos',
        icon: 'building',
        isActive: false,
        shortcut: ['a', 'd'],
        access: { permission: 'admin::settings.departments.read' },
        items: []
      },
      {
        title: 'Auditoria',
        url: '/dashboard/administracao/auditoria',
        icon: 'history',
        isActive: false,
        shortcut: ['a', 'a'],
        access: { permission: 'admin::settings.audit-logs.read' },
        items: []
      },
      {
        title: 'Tokens de API',
        url: '/dashboard/tokens-api',
        icon: 'key',
        isActive: false,
        shortcut: ['a', 't'],
        access: { permission: 'admin::settings.api-tokens.read' },
        items: []
      }
    ]
  },
  {
    label: 'Portal público',
    items: [
      {
        title: 'Usuários finais',
        url: '/dashboard/portal/usuarios',
        icon: 'user',
        isActive: false,
        shortcut: ['p', 'u'],
        access: { permission: 'admin::plugin.enduser-users.read' },
        items: []
      },
      {
        title: 'Perfis públicos',
        url: '/dashboard/portal/perfis',
        icon: 'lock',
        isActive: false,
        shortcut: ['p', 'p'],
        access: { permission: 'admin::plugin.enduser-roles.read' },
        items: []
      },
      {
        title: 'Configurações',
        url: '/dashboard/portal/configuracoes',
        icon: 'settings',
        isActive: false,
        shortcut: ['p', 'c'],
        access: { permission: 'admin::plugin.enduser-settings.read' },
        items: []
      },
      {
        title: 'Templates de e-mail',
        url: '/dashboard/portal/templates-email',
        icon: 'send',
        isActive: false,
        shortcut: ['p', 't'],
        access: { permission: 'admin::settings.email-templates.read' },
        items: []
      }
    ]
  },
  {
    label: 'Conta',
    items: [
      {
        title: 'Meu perfil',
        url: '/dashboard/profile',
        icon: 'account',
        isActive: false,
        shortcut: ['m', 'm'],
        items: []
      }
    ]
  }
];
