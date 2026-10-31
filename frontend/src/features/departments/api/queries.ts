import { queryOptions } from '@tanstack/react-query';
import { listDepartments } from './service';

export const departmentKeys = {
  all: ['departments'] as const,
  list: (finalistic = false) => [...departmentKeys.all, 'list', { finalistic }] as const
};

export const departmentsQueryOptions = (finalistic = false) =>
  queryOptions({
    queryKey: departmentKeys.list(finalistic),
    queryFn: () => listDepartments({ finalistic })
  });
