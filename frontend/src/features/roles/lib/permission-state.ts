import type { PermissionView } from '../api/types';

/** Estado de edição das permissões de uma role, por ação. */
export interface PermEntry {
  /** `null` = todos os campos. */
  fields: string[] | null;
  /** Vazio = sem restrição; várias condições combinam com OR. */
  conditions: string[];
}

export type PermState = Record<string, PermEntry>;

export type PermAction =
  | { type: 'reset'; state: PermState }
  | { type: 'toggle'; action: string; on: boolean }
  | { type: 'toggleMany'; actions: string[]; on: boolean }
  | { type: 'fields'; action: string; fields: string[] | null }
  | { type: 'conditions'; action: string; conditions: string[] };

const emptyEntry = (): PermEntry => ({ fields: null, conditions: [] });

export function permReducer(state: PermState, a: PermAction): PermState {
  switch (a.type) {
    case 'reset':
      return a.state;
    case 'toggle':
    case 'toggleMany': {
      const actions = a.type === 'toggle' ? [a.action] : a.actions;
      const next = { ...state };
      for (const id of actions) {
        if (!a.on) delete next[id];
        else next[id] ??= emptyEntry();
      }
      return next;
    }
    case 'fields':
      return state[a.action]
        ? { ...state, [a.action]: { ...state[a.action], fields: a.fields } }
        : state;
    case 'conditions':
      return state[a.action]
        ? { ...state, [a.action]: { ...state[a.action], conditions: a.conditions } }
        : state;
  }
}

export function toState(permissions: PermissionView[]): PermState {
  return Object.fromEntries(
    permissions.map((p) => [p.action, { fields: p.fields, conditions: p.conditions }])
  );
}

/** Monta as permissões a enviar; `resources` mapeia a ação ao recurso (subject) do catálogo. */
export function toPermissions(state: PermState, resources: Map<string, string>): PermissionView[] {
  return Object.entries(state)
    .filter(([action]) => resources.has(action))
    .map(([action, e]) => ({
      action,
      subject: resources.get(action),
      fields: e.fields,
      conditions: e.conditions
    }));
}

const normalize = (state: PermState) =>
  JSON.stringify(
    Object.keys(state)
      .toSorted()
      .map((k) => [k, state[k].fields?.toSorted() ?? null, state[k].conditions.toSorted()])
  );

export const isDirty = (a: PermState, b: PermState) => normalize(a) !== normalize(b);
