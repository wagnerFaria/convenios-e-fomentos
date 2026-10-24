/** Data no formato brasileiro (ex.: 24/10/2026). */
export function formatDate(date: Date | string | number | undefined) {
  if (!date) return '';
  const d = new Date(date);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('pt-BR');
}

/** Data e hora no formato brasileiro (ex.: 22/10/2026 14:30). */
export function formatDateTime(date: Date | string | number | null | undefined) {
  if (!date) return '';
  const d = new Date(date);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' });
}
