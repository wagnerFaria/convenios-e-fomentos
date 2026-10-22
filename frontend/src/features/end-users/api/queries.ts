import { queryOptions } from '@tanstack/react-query';
import {
  getEndUserRole,
  getEndUserSettings,
  listEmailTemplates,
  listEndUserRoles,
  listEndUsers
} from './service';
import type { EndUserListParams } from './types';

export const endUserKeys = {
  all: ['end-users'] as const,
  list: (params: EndUserListParams) => [...endUserKeys.all, 'list', params] as const,
  roles: () => [...endUserKeys.all, 'roles'] as const,
  role: (id: string) => [...endUserKeys.all, 'roles', id] as const,
  settings: () => [...endUserKeys.all, 'settings'] as const,
  templates: () => [...endUserKeys.all, 'templates'] as const
};

export const endUsersQueryOptions = (params: EndUserListParams = {}) =>
  queryOptions({ queryKey: endUserKeys.list(params), queryFn: () => listEndUsers(params) });

export const endUserRolesQueryOptions = () =>
  queryOptions({ queryKey: endUserKeys.roles(), queryFn: listEndUserRoles });

export const endUserRoleQueryOptions = (id: string) =>
  queryOptions({ queryKey: endUserKeys.role(id), queryFn: () => getEndUserRole(id) });

export const endUserSettingsQueryOptions = () =>
  queryOptions({ queryKey: endUserKeys.settings(), queryFn: getEndUserSettings });

export const emailTemplatesQueryOptions = () =>
  queryOptions({ queryKey: endUserKeys.templates(), queryFn: listEmailTemplates });
