import type { Metadata } from 'next';
import PageContainer from '@/components/layout/page-container';
import { entityTypeActions } from '@/features/entity-types/api/types';
import { EntityTypeDialog } from '@/features/entity-types/components/entity-type-dialog';
import { EntityTypesList } from '@/features/entity-types/components/entity-types-list';
import { hasPermission } from '@/lib/auth/access';

export const metadata: Metadata = { title: 'Tipos de entidade' };

export default async function TiposEntidadePage() {
  return (
    <PageContainer
      access={await hasPermission(entityTypeActions.read)}
      pageTitle='Tipos de entidade'
      pageDescription='Lista de tipos que o proponente escolhe ao cadastrar um projeto. A natureza (pública ou privada) define o checklist e as regras do instrumento.'
      pageHeaderAction={<EntityTypeDialog />}
    >
      <EntityTypesList />
    </PageContainer>
  );
}
