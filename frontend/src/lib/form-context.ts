'use client';

import { createFormHookContexts } from '@tanstack/react-form';
import { createContext, useContext } from 'react';

export const { fieldContext, formContext, useFieldContext, useFormContext } =
  createFormHookContexts();

/**
 * The doc's "touched + invalid" check, shared by every field component:
 * errors surface after the user interacts with a field or a submit paints it.
 */
export function useFieldInvalid() {
  const field = useFieldContext();
  return field.state.meta.isTouched && !field.state.meta.isValid;
}

/** Props every field component shares. */
export type BaseFieldProps = {
  label: string;
  description?: string;
  /** Appends ' *' to the label. Purely visual — rules live in the schema. */
  required?: boolean;
};

const FormReadOnlyContext = createContext(false);

/** Provider do modo somente leitura; use `ReadOnlyScope` (components/forms/read-only-scope) para ligar um formulário inteiro. */
export const FormReadOnlyProvider = FormReadOnlyContext.Provider;

/** O formulário está só para leitura: os campos mostram o valor, mas não deixam alterá-lo. */
export function useFormReadOnly() {
  return useContext(FormReadOnlyContext);
}
