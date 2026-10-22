import { Button } from '@/components/ui/button';
import type { ApiMeta } from '@/lib/api/types';

/** Paginação simples (Anterior/Próxima) para as listagens paginadas pelo backend. */
export function Pager({ meta, onPage }: { meta?: ApiMeta; onPage: (page: number) => void }) {
  if (!meta || meta.pageCount <= 1) return null;
  return (
    <div className='flex items-center justify-end gap-2 text-sm'>
      <span className='text-muted-foreground'>
        Página {meta.page} de {meta.pageCount} · {meta.total} registros
      </span>
      <Button
        size='sm'
        variant='outline'
        disabled={meta.page <= 1}
        onClick={() => onPage(meta.page - 1)}
      >
        Anterior
      </Button>
      <Button
        size='sm'
        variant='outline'
        disabled={meta.page >= meta.pageCount}
        onClick={() => onPage(meta.page + 1)}
      >
        Próxima
      </Button>
    </div>
  );
}
