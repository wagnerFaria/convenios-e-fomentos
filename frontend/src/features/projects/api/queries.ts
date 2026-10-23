import { queryOptions } from '@tanstack/react-query';
import {
  getChecklist,
  getProfile,
  getProject,
  listMembers,
  listProjects,
  type ListProjectsParams
} from './service';

export const projectKeys = {
  all: ['projects'] as const,
  lists: () => [...projectKeys.all, 'list'] as const,
  list: (params: ListProjectsParams) => [...projectKeys.lists(), params] as const,
  detail: (id: string) => [...projectKeys.all, 'detail', id] as const,
  checklist: (id: string) => [...projectKeys.all, 'checklist', id] as const,
  members: (id: string) => [...projectKeys.all, 'members', id] as const,
  profile: () => ['proponent-profile'] as const
};

export const projectsQueryOptions = (params: ListProjectsParams = {}) =>
  queryOptions({ queryKey: projectKeys.list(params), queryFn: () => listProjects(params) });

export const projectQueryOptions = (id: string) =>
  queryOptions({ queryKey: projectKeys.detail(id), queryFn: () => getProject(id) });

export const profileQueryOptions = () =>
  queryOptions({ queryKey: projectKeys.profile(), queryFn: getProfile });

export const checklistQueryOptions = (id: string) =>
  queryOptions({ queryKey: projectKeys.checklist(id), queryFn: () => getChecklist(id) });

export const membersQueryOptions = (id: string) =>
  queryOptions({ queryKey: projectKeys.members(id), queryFn: () => listMembers(id) });
