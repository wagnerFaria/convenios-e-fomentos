'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo, useReducer, useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger
} from '@/components/ui/accordion';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { adminKeys } from '@/features/admins/api/queries';
import { useSession } from '@/features/auth/session-provider';
import { ApiError } from '@/lib/api/errors';
import { errorMessage } from '@/lib/errors';
import { roleKeys } from '../api/queries';
import { replacePermissions } from '../api/service';
import type { CatalogAction, CatalogCondition, CatalogGroup, RoleDetail } from '../api/types';
import { categoryLabel, resourceLabel, sortCategories } from '../lib/labels';
import {
  isDirty,
  permReducer,
  toPermissions,
  toState,
  type PermEntry
} from '../lib/permission-state';

/** Se o backend recusar por escalonamento, lista o que o ator não pode conceder. */
function saveError(err: unknown, names: Map<string, string>): string {
  if (err instanceof ApiError && err.code === 'privilege_escalation') {
    const list = (err.details?.permissions as { action: string }[] | undefined) ?? [];
    const labels = list.map((p) => names.get(p.action) ?? p.action);
    return `${err.message}${labels.length ? ` Sem permissão para conceder: ${labels.join('; ')}.` : ''}`;
  }
  return errorMessage(err);
}

function ActionOptions({
  action,
  entry,
  conditions,
  disabled,
  onFields,
  onConditions
}: {
  action: CatalogAction;
  entry: PermEntry;
  conditions: CatalogCondition[];
  disabled: boolean;
  onFields: (fields: string[] | null) => void;
  onConditions: (conditions: string[]) => void;
}) {
  const available = action.availableFields ?? [];
  const allFields = entry.fields === null;
  return (
    <div className='bg-muted/40 grid gap-3 rounded-md border p-3 text-sm'>
      {action.supportsFields && (
        <div className='grid gap-2'>
          <div className='flex items-center gap-2'>
            <Switch
              id={`all-${action.id}`}
              size='sm'
              checked={allFields}
              disabled={disabled}
              onCheckedChange={(on) => onFields(on ? null : available)}
            />
            <Label htmlFor={`all-${action.id}`}>Todos os campos</Label>
          </div>
          {!allFields && (
            <div
              className='grid grid-cols-2 gap-2 sm:grid-cols-3'
              role='group'
              aria-label='Campos permitidos'
            >
              {available.map((f) => (
                <div key={f} className='flex items-center gap-2'>
                  <Checkbox
                    id={`${action.id}-f-${f}`}
                    disabled={disabled}
                    checked={entry.fields?.includes(f) ?? false}
                    onCheckedChange={(on) =>
                      onFields(
                        on
                          ? [...(entry.fields ?? []), f]
                          : (entry.fields ?? []).filter((x) => x !== f)
                      )
                    }
                  />
                  <Label htmlFor={`${action.id}-f-${f}`} className='font-normal'>
                    {f}
                  </Label>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
      {action.supportsConditions && (
        <div className='grid gap-2' role='group' aria-label='Condições'>
          <p className='text-muted-foreground'>
            Restringir a registros em que… (sem seleção: sem restrição; várias: vale qualquer uma)
          </p>
          {conditions.map((c) => (
            <div key={c.id} className='flex items-center gap-2'>
              <Checkbox
                id={`${action.id}-c-${c.id}`}
                disabled={disabled}
                checked={entry.conditions.includes(c.id)}
                onCheckedChange={(on) =>
                  onConditions(
                    on ? [...entry.conditions, c.id] : entry.conditions.filter((x) => x !== c.id)
                  )
                }
              />
              <Label htmlFor={`${action.id}-c-${c.id}`} className='font-normal'>
                {c.name}
              </Label>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function PermissionEditor({
  role,
  catalog,
  conditions
}: {
  role: RoleDetail;
  catalog: CatalogGroup[];
  conditions: CatalogCondition[];
}) {
  const session = useSession();
  const qc = useQueryClient();
  const initial = useMemo(() => toState(role.permissions), [role.permissions]);
  const [state, dispatch] = useReducer(permReducer, initial);
  const [error, setError] = useState<string | null>(null);

  const actions = useMemo(
    () => new Map(catalog.flatMap((g) => g.actions.map((a) => [a.id, a]))),
    [catalog]
  );
  const resources = useMemo(
    () => new Map([...actions].map(([id, a]) => [id, a.resource])),
    [actions]
  );
  const names = useMemo(() => new Map([...actions].map(([id, a]) => [id, a.name])), [actions]);
  const owned = useMemo(
    () => new Set(session.permissions.map((p) => p.action)),
    [session.permissions]
  );
  const dirty = isDirty(state, initial);

  const ownRole = !session.isSuperAdmin && session.roles.some((r) => r.id === role.id);
  const readOnly = role.isSuperAdmin || ownRole;
  const canGrant = (id: string) => session.isSuperAdmin || owned.has(id);

  // Avisa ao sair da página com alterações pendentes.
  useEffect(() => {
    if (!dirty) return;
    const warn = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener('beforeunload', warn);
    return () => window.removeEventListener('beforeunload', warn);
  }, [dirty]);

  const save = useMutation({
    mutationFn: () => replacePermissions(role.id, toPermissions(state, resources)),
    onSuccess: async (saved) => {
      dispatch({ type: 'reset', state: toState(saved.permissions) });
      await Promise.all([
        qc.invalidateQueries({ queryKey: roleKeys.all }),
        qc.invalidateQueries({ queryKey: adminKeys.all })
      ]);
      toast.success('Permissões salvas.');
    },
    onError: (err) => setError(saveError(err, names))
  });

  if (role.isSuperAdmin) {
    return (
      <Alert>
        <Icons.shieldLock aria-hidden />
        <AlertDescription>
          O Super Admin tem acesso total e não possui permissões editáveis.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className='grid gap-4'>
      {ownRole && (
        <Alert>
          <Icons.info aria-hidden />
          <AlertDescription>
            Você possui este perfil, então não pode alterar as permissões dele. Peça a outro
            administrador.
          </AlertDescription>
        </Alert>
      )}
      <Accordion multiple defaultValue={[]}>
        {sortCategories(catalog).map((group) => {
          const groupIds = group.actions.map((a) => a.id);
          const count = groupIds.filter((id) => state[id]).length;
          const byResource = Map.groupBy(group.actions, (a) => a.resource);
          return (
            <AccordionItem key={group.category} value={group.category}>
              <AccordionTrigger>
                <span className='flex items-center gap-2'>
                  {categoryLabel(group.category)}
                  <Badge variant={count ? 'default' : 'secondary'}>
                    {count}/{groupIds.length}
                  </Badge>
                </span>
              </AccordionTrigger>
              <AccordionContent>
                <div className='grid gap-5'>
                  {[...byResource].map(([resource, list]) => {
                    const ids = list.map((a) => a.id);
                    const on = ids.filter((id) => state[id]).length;
                    return (
                      <section key={resource} className='grid gap-2'>
                        <div className='flex items-center gap-2'>
                          <h4 className='text-sm font-medium'>{resourceLabel(resource)}</h4>
                          {!readOnly && (
                            <Button
                              size='xs'
                              variant='ghost'
                              disabled={
                                on < ids.length && ids.some((id) => !canGrant(id) && !state[id])
                              }
                              onClick={() =>
                                dispatch({ type: 'toggleMany', actions: ids, on: on < ids.length })
                              }
                            >
                              {on < ids.length ? 'Marcar todas' : 'Limpar'}
                            </Button>
                          )}
                        </div>
                        <div className='grid gap-2 pl-6'>
                          {list.map((a) => {
                            const entry = state[a.id];
                            const refinable = a.supportsFields || a.supportsConditions;
                            return (
                              <div key={a.id} className='grid gap-2'>
                                <div className='flex items-center gap-2'>
                                  <Checkbox
                                    id={a.id}
                                    disabled={readOnly || (!canGrant(a.id) && !entry)}
                                    checked={!!entry}
                                    onCheckedChange={(v) =>
                                      dispatch({ type: 'toggle', action: a.id, on: !!v })
                                    }
                                  />
                                  <Label htmlFor={a.id} className='font-normal'>
                                    {a.name}
                                  </Label>
                                  {!canGrant(a.id) && !readOnly && (
                                    <span className='text-muted-foreground text-xs'>
                                      (você não possui esta permissão)
                                    </span>
                                  )}
                                  {entry && refinable && (
                                    <span className='text-muted-foreground text-xs'>
                                      {entry.fields ? `${entry.fields.length} campo(s)` : ''}
                                      {entry.conditions.length
                                        ? ` · ${entry.conditions.length} condição(ões)`
                                        : ''}
                                    </span>
                                  )}
                                </div>
                                {entry && refinable && (
                                  <ActionOptions
                                    action={a}
                                    entry={entry}
                                    conditions={conditions}
                                    disabled={readOnly}
                                    onFields={(fields) =>
                                      dispatch({ type: 'fields', action: a.id, fields })
                                    }
                                    onConditions={(c) =>
                                      dispatch({ type: 'conditions', action: a.id, conditions: c })
                                    }
                                  />
                                )}
                              </div>
                            );
                          })}
                        </div>
                      </section>
                    );
                  })}
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>

      {error && (
        <Alert variant='destructive'>
          <AlertDescription role='alert'>{error}</AlertDescription>
        </Alert>
      )}
      {!readOnly && (
        <div className='bg-background sticky bottom-0 flex flex-wrap items-center justify-between gap-2 border-t py-3'>
          <span className='text-muted-foreground text-sm'>
            {dirty ? 'Há alterações não salvas.' : 'Sem alterações.'} {Object.keys(state).length}{' '}
            permissão(ões) concedida(s).
          </span>
          <div className='flex gap-2'>
            <Button
              variant='outline'
              disabled={!dirty || save.isPending}
              onClick={() => {
                setError(null);
                dispatch({ type: 'reset', state: initial });
              }}
            >
              Descartar
            </Button>
            <Button
              disabled={!dirty || save.isPending}
              onClick={() => {
                setError(null);
                save.mutate();
              }}
            >
              Salvar permissões
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
