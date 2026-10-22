import { queryOptions } from '@tanstack/react-query';
import { listTokens } from './service';

export const tokenKeys = {
  all: ['api-tokens'] as const,
  list: (includeRevoked: boolean) => [...tokenKeys.all, 'list', includeRevoked] as const
};

export const tokensQueryOptions = (includeRevoked = false) =>
  queryOptions({
    queryKey: tokenKeys.list(includeRevoked),
    queryFn: () => listTokens(includeRevoked)
  });
