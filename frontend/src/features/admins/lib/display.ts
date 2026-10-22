import type { AdminUser } from '../api/types';

/** Quem ainda não fez o primeiro acesso não tem nome: o MT Login o preenche no primeiro login. */
export function adminDisplayName(a: Pick<AdminUser, 'firstName' | 'lastName' | 'status'>): string {
  const name = `${a.firstName} ${a.lastName}`.trim();
  if (name) return name;
  return a.status === 'pending' ? 'Aguardando o primeiro acesso' : 'Sem nome';
}
