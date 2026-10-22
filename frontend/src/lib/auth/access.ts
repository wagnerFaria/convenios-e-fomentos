import 'server-only';
import { getSession } from './session';

/** Para gatear páginas no servidor; a autorização real continua sendo do backend. */
export async function hasPermission(action: string): Promise<boolean> {
  const session = await getSession();
  return (
    !!session && (session.isSuperAdmin || session.permissions.some((p) => p.action === action))
  );
}
