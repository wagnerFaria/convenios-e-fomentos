'use client';

import type { ReactNode } from 'react';
import { FormReadOnlyProvider } from '@/lib/form-context';

/**
 * Deixa o formulário só para leitura: os campos do `useAppForm` leem o contexto, e o `fieldset` desabilitado cobre os controles nativos
 * dos editores próprios (inputs, selects e botões). `contents` mantém o layout como está.
 */
export function ReadOnlyScope({ readOnly, children }: { readOnly: boolean; children: ReactNode }) {
  return (
    <FormReadOnlyProvider value={readOnly}>
      <fieldset disabled={readOnly} className='contents'>
        {children}
      </fieldset>
    </FormReadOnlyProvider>
  );
}
