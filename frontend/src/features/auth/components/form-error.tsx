'use client';

import { Alert, AlertDescription } from '@/components/ui/alert';
import { useFormContext } from '@/lib/form-context';

/** Erro geral do formulário vindo do validador de submit (`errorMap.onSubmit`). */
export function FormError() {
  const form = useFormContext();
  return (
    <form.Subscribe selector={(state) => state.errorMap.onSubmit}>
      {(error) =>
        typeof error === 'string' && error ? (
          <Alert variant='destructive'>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null
      }
    </form.Subscribe>
  );
}
