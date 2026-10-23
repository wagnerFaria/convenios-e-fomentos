/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const ruleParameterActions = {
  read: 'admin::settings.rule-parameters.read',
  create: 'admin::settings.rule-parameters.create',
  update: 'admin::settings.rule-parameters.update',
  delete: 'admin::settings.rule-parameters.delete'
} as const;

export interface RuleParameter {
  id: string;
  key: string;
  /** JSON no formato definido pelo backend para a chave. */
  value: unknown;
  /** Datas AAAA-MM-DD. */
  validFrom: string;
  validTo: string | null;
  note: string;
  createdAt: string;
  updatedAt: string;
}

export interface RuleParameterKey {
  key: string;
  description: string;
  /** Norma de origem. */
  source: string;
}

export interface RuleParameterInput {
  key: string;
  value: unknown;
  validFrom: string;
  validTo: string | null;
  note: string;
}

/** Exemplo do formato do valor de cada chave, mostrado no formulário. */
export const valueExamples: Record<string, unknown> = {
  complement_days: 3,
  price_reference_max_age_days: 60,
  electoral_deadline: '2026-05-15',
  lei_12082: {
    ceiling_cents: 60000000,
    min_municipal_counterpart_pct: 10,
    max_fees_pct: 50,
    min_regional_fees_pct: 30
  },
  indirect_costs_tiers: [
    { up_to_cents: 10000000, max_pct: 20, cap_cents: 1500000 },
    { up_to_cents: 50000000, max_pct: 15, cap_cents: 5000000 },
    { up_to_cents: null, max_pct: 10, cap_cents: 20000000 }
  ]
};
