import { queryOptions } from '@tanstack/react-query';
import { listRuleParameterKeys, listRuleParameters } from './service';

export const ruleParameterKeys = {
  all: ['rule-parameters'] as const,
  list: () => [...ruleParameterKeys.all, 'list'] as const,
  keys: () => [...ruleParameterKeys.all, 'keys'] as const
};

export const ruleParametersQueryOptions = () =>
  queryOptions({ queryKey: ruleParameterKeys.list(), queryFn: listRuleParameters });

export const ruleParameterKeysQueryOptions = () =>
  queryOptions({
    queryKey: ruleParameterKeys.keys(),
    queryFn: listRuleParameterKeys,
    staleTime: 5 * 60_000
  });
