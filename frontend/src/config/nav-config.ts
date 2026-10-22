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
