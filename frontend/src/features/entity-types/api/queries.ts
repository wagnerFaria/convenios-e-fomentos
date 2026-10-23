import { queryOptions } from '@tanstack/react-query';
import { listEntityTypes } from './service';

export const entityTypeKeys = {
  all: ['entity-types'] as const,
  list: () => [...entityTypeKeys.all, 'list'] as const
};

export const entityTypesQueryOptions = () =>
  queryOptions({ queryKey: entityTypeKeys.list(), queryFn: listEntityTypes });
