import PageContainer from '@/components/layout/page-container';

export default function OverviewLayout({ children }: { children: React.ReactNode }) {
  return <PageContainer pageTitle='Visão geral'>{children}</PageContainer>;
}
