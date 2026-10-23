import { queryOptions } from '@tanstack/react-query';
import { listParliamentarians } from './service';

export const parliamentarianKeys = {
  all: ['parliamentarians'] as const,
  list: () => [...parliamentarianKeys.all, 'list'] as const
};

export const parliamentariansQueryOptions = () =>
  queryOptions({ queryKey: parliamentarianKeys.list(), queryFn: listParliamentarians });
