const brl = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });

/** Centavos → "R$ 600.000,00". */
export const money = (cents: unknown) =>
  typeof cents === 'number' ? brl.format(cents / 100) : '—';

/** AAAA-MM-DD → DD/MM/AAAA (sem passar por Date, que deslocaria o dia pelo fuso). */
export function formatDate(iso: string) {
  const [y, m, d] = iso.split('-');
  return y && m && d ? `${d}/${m}/${y}` : iso;
}

/** Hoje (AAAA-MM-DD) no fuso de Cuiabá, o mesmo usado pelo backend para a vigência. */
export function todayInCuiaba() {
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Cuiaba' }).format(new Date());
}

/** Resumo legível do valor de um parâmetro; chaves desconhecidas caem no JSON. */
export function summarize(key: string, value: unknown): string[] {
  const v = value as Record<string, unknown> | unknown[] | string | number;
  switch (key) {
    case 'complement_days':
    case 'price_reference_max_age_days':
      return [`${String(v)} dias`];
    case 'electoral_deadline':
      return [typeof v === 'string' ? formatDate(v) : JSON.stringify(v)];
    case 'lei_12082': {
      const o = v as Record<string, unknown>;
      return [
        `Teto do repasse: ${money(o.ceiling_cents)}`,
        `Contrapartida mínima de municípios: ${String(o.min_municipal_counterpart_pct)}%`,
        `Cachês: até ${String(o.max_fees_pct)}% do repasse; ${String(o.min_regional_fees_pct)}% no mínimo para artistas regionais`
      ];
    }
    case 'indirect_costs_tiers':
      return (v as { up_to_cents: number | null; max_pct: number; cap_cents: number }[]).map(
        (t) =>
          `${t.up_to_cents === null ? 'Acima da faixa anterior' : `Até ${money(t.up_to_cents)}`}: ${t.max_pct}%, limitado a ${money(t.cap_cents)}`
      );
    default:
      return [JSON.stringify(value)];
  }
}
