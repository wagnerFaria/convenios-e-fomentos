import 'server-only';
import { isProjectProponent, isProjectStaff } from '@/features/projects/access';
import { getSession } from './session';

/** Para gatear páginas no servidor; a autorização real continua sendo do backend. */
export async function hasPermission(action: string): Promise<boolean> {
  const session = await getSession();
  return (
    !!session && (session.isSuperAdmin || session.permissions.some((p) => p.action === action))
  );
}

/** Equipe de projetos (Analista, Revisor, Secretário, Supervisor) ou Super Admin; mesmo critério do menu. */
export async function isProjectStaffSession(): Promise<boolean> {
  const session = await getSession();
  return !!session && isProjectStaff(session);
}

/** Quem cadastra projetos (Proponente) ou Super Admin; mesmo critério do menu. */
export async function isProjectProponentSession(): Promise<boolean> {
  const session = await getSession();
  return !!session && isProjectProponent(session);
}
