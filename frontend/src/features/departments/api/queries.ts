import { queryOptions } from '@tanstack/react-query';
import { listDepartments } from './service';

export const departmentKeys = {
  all: ['departments'] as const,
  list: () => [...departmentKeys.all, 'list'] as const
};

export const departmentsQueryOptions = () =>
  queryOptions({ queryKey: departmentKeys.list(), queryFn: listDepartments });
