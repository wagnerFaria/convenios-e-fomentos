'use client';

import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';

/** Falha ao exibir o arquivo: diz o motivo e deixa tentar de novo ou baixar o original. */
export function ViewerError({
  message,
  onRetry,
  onDownload
}: {
  message: string;
  onRetry?: () => void;
  onDownload?: () => void;
}) {
  return (
    <div
      role='alert'
      className='flex h-full min-h-64 flex-col items-center justify-center gap-3 p-6 text-center'
    >
      <Icons.warning className='text-destructive size-8' aria-hidden />
      <p className='max-w-md text-sm'>{message}</p>
      <div className='flex flex-wrap justify-center gap-2'>
        {onRetry && (
          <Button variant='outline' size='sm' onClick={onRetry}>
            <Icons.refresh aria-hidden /> Tentar de novo
          </Button>
        )}
        {onDownload && (
          <Button variant='outline' size='sm' onClick={onDownload}>
            <Icons.download aria-hidden /> Baixar o arquivo
          </Button>
        )}
      </div>
    </div>
  );
}
