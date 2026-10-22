'use client';

import { useQuery } from '@tanstack/react-query';
import { QueryError } from '@/components/query-error';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { catalogActionsQueryOptions } from '../api/queries';

/** Rótulos dos recursos da API pública; os demais aparecem pelo identificador. */
const apiResourceLabels: Record<string, string> = { document: 'Documentos' };

/**
 * Checklist das ações da API pública (`api::`), agrupadas por recurso, a partir do catálogo.
 * Usado nos tokens de API do tipo personalizado e nas roles de usuário final.
 */
export function ApiActionPicker({
  value,
  onChange,
  disabled
}: {
  value: string[];
  onChange: (value: string[]) => void;
  disabled?: boolean;
}) {
  const catalog = useQuery(catalogActionsQueryOptions('api'));
  if (catalog.isLoading) return <Skeleton className='h-24 w-full' />;
  if (catalog.isError || !catalog.data) {
    return <QueryError error={catalog.error} what='as ações da API' />;
  }

  const actions = catalog.data.flatMap((g) => g.actions);
  const byResource = Map.groupBy(actions, (a) => a.resource || 'outros');
  const toggle = (id: string, on: boolean) =>
    onChange(on ? [...value, id] : value.filter((x) => x !== id));

  return (
    <div className='grid gap-4' role='group' aria-label='Ações da API'>
      {[...byResource].map(([resource, list]) => (
        <section key={resource} className='grid gap-2'>
          <h4 className='text-sm font-medium'>{apiResourceLabels[resource] ?? resource}</h4>
          <div className='grid gap-1.5 pl-1'>
            {list.map((a) => (
              <div key={a.id} className='flex flex-wrap items-center gap-2'>
                <Checkbox
                  id={a.id}
                  disabled={disabled}
                  checked={value.includes(a.id)}
                  onCheckedChange={(on) => toggle(a.id, !!on)}
                />
                <Label htmlFor={a.id} className='font-normal'>
                  {a.name}
                </Label>
                {a.method && (
                  <code className='text-muted-foreground text-xs'>
                    {a.method} {a.route}
                  </code>
                )}
              </div>
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
