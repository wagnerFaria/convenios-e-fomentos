import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { ruleParameterActions } from '@/features/rule-parameters/api/types';
import { RuleParametersList } from '@/features/rule-parameters/components/rule-parameters-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Parâmetros de regras' };

export default async function ParametrosPage() {
  return (
    <PageContainer
      access={await hasPermission(ruleParameterActions.read)}
      pageTitle='Parâmetros de regras'
      pageDescription='Limites, prazos e datas usados nas validações do cadastro de projetos. Cada parâmetro tem versões com vigência, para acompanhar as mudanças das normas.'
    >
      <RuleParametersList />
    </PageContainer>
  );
}
