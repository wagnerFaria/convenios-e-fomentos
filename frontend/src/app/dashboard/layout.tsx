import KBar from '@/components/kbar';
import AppSidebar from '@/components/layout/app-sidebar';
import Header from '@/components/layout/header';
import { InfoSidebar } from '@/components/layout/info-sidebar';
import { InfobarProvider } from '@/components/ui/infobar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { AwaitingAccess } from '@/features/auth/components/awaiting-access';
import { fullName } from '@/features/auth/full-name';
import { SessionProvider } from '@/features/auth/session-provider';
import { getSession } from '@/lib/auth/session';
import type { Metadata } from 'next';
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

export const metadata: Metadata = {
  title: 'Painel',
  description: 'Painel administrativo de Convênios e Fomentos',
  robots: {
    index: false,
    follow: false
  }
};

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  // O proxy.ts renova tokens expirados; aqui a sessão é validada no backend (/api/admin/auth/me).
  // Layouts nem sempre re-renderizam junto com as páginas, então páginas sensíveis também
  // devem chamar getSession().
  const session = await getSession();
  if (!session) redirect('/auth/sign-in');
  // Primeiro acesso: a conta existe, mas ainda sem perfil (roles) — nada do painel abriria.
  if (!session.isSuperAdmin && session.permissions.length === 0) {
    return <AwaitingAccess name={fullName(session)} />;
  }
  // Persisting the sidebar state in the cookie.
  const cookieStore = await cookies();
  const defaultOpen = cookieStore.get('sidebar_state')?.value === 'true';
  return (
    <SessionProvider session={session}>
      <KBar>
        <SidebarProvider defaultOpen={defaultOpen}>
          <a
            href='#main-content'
            className='bg-background ring-ring sr-only rounded-md px-3 py-2 text-sm font-medium shadow focus:not-sr-only focus:absolute focus:top-2 focus:start-2 focus:z-50 focus:ring-2'
          >
            Pular para o conteúdo
          </a>
          <AppSidebar />
          <SidebarInset id='main-content' tabIndex={-1} className='scroll-mt-16'>
            <Header />
            <InfobarProvider defaultOpen={false}>
              {children}
              <InfoSidebar side='right' />
            </InfobarProvider>
          </SidebarInset>
        </SidebarProvider>
      </KBar>
    </SessionProvider>
  );
}
