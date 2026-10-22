import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { departmentActions } from '@/features/departments/api/types';
import { DepartmentDialog } from '@/features/departments/components/department-dialog';
import { DepartmentsList } from '@/features/departments/components/departments-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Departamentos' };

export default async function DepartamentosPage() {
  return (
    <PageContainer
      access={await hasPermission(departmentActions.read)}
      pageTitle='Departamentos'
      pageDescription='Agrupam administradores e documentos e delimitam o alcance de permissões por departamento.'
      pageHeaderAction={<DepartmentDialog />}
    >
      <DepartmentsList />
    </PageContainer>
  );
}
