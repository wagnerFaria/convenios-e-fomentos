'use client';

import { LoadingButton } from '@/components/ui/loading-button';

/**
 * Botões de rodapé de cada etapa. "Salvar" fica na etapa; "Salvar e continuar" avança depois de gravar.
 * Os dois enviam o formulário; os callbacks avisam o envio qual deles foi clicado.
 */
export function StepActions({
  saving,
  onSave,
  onSaveAndContinue,
  last,
  disabled
}: {
  saving: boolean;
  onSave: () => void;
  onSaveAndContinue: () => void;
  last?: boolean;
  disabled?: boolean;
}) {
  // Sem permissão de gravar, não há o que salvar: os botões nem aparecem.
  if (disabled) return null;
  return (
    <div className='flex flex-wrap gap-2 pt-2'>
      <LoadingButton type='submit' loading={saving} disabled={saving || disabled} onClick={onSave}>
        Salvar
      </LoadingButton>
      {!last && (
        <LoadingButton
          type='submit'
          variant='outline'
          disabled={saving || disabled}
          onClick={onSaveAndContinue}
        >
          Salvar e continuar
        </LoadingButton>
      )}
    </div>
  );
}
