import type { ActionFailure } from './action-result';

/**
 * Converte a falha de uma server action no formato dos validadores de submit do TanStack Form
 * (`onSubmitAsync`): mensagem do formulário + erros por campo, limpos automaticamente ao editar.
 */
export function submitErrors(failure: ActionFailure) {
  return {
    form: failure.message,
    fields: Object.fromEntries(failure.fields.map((f) => [f.field, { message: f.message }]))
  };
}
