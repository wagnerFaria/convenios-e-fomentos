export const onlyDigits = (s: string) => s.replace(/\D/g, '');

/** 12345678909 → 123.456.789-09 (entradas que não têm 11 dígitos voltam como estão). */
export function formatCpf(v: string) {
  const d = onlyDigits(v);
  return d.length === 11 ? `${d.slice(0, 3)}.${d.slice(3, 6)}.${d.slice(6, 9)}-${d.slice(9)}` : v;
}

/** 11222333000181 → 11.222.333/0001-81. */
export function formatCnpj(v: string) {
  const d = onlyDigits(v);
  return d.length === 14
    ? `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8, 12)}-${d.slice(12)}`
    : v;
}

/** 65999998888 → (65) 99999-8888; 6533334444 → (65) 3333-4444. */
export function formatPhone(v: string) {
  const d = onlyDigits(v);
  if (d.length === 11) return `(${d.slice(0, 2)}) ${d.slice(2, 7)}-${d.slice(7)}`;
  if (d.length === 10) return `(${d.slice(0, 2)}) ${d.slice(2, 6)}-${d.slice(6)}`;
  return v;
}

/** 78000000 → 78000-000. */
export function formatZip(v: string) {
  const d = onlyDigits(v);
  return d.length === 8 ? `${d.slice(0, 5)}-${d.slice(5)}` : v;
}

/** AAAA-MM-DD → DD/MM/AAAA, sem passar por Date (que deslocaria o dia pelo fuso). */
export function formatDate(iso: string) {
  const [y, m, d] = iso.split('-');
  return y && m && d ? `${d}/${m}/${y}` : iso;
}

const brl = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });

/** Centavos → "R$ 1.234,56". */
export const formatBrl = (cents: number) => brl.format(cents / 100);
