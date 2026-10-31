'use client';

import { useState } from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { ViewerError } from './viewer-error';
import { ViewerToolbar } from './viewer-toolbar';

/** Imagem anexada, ajustada à largura da área de leitura, com zoom. `onShown` avisa uma vez que ela apareceu. */
export function ImageViewer({
  url,
  filename,
  onShown,
  onRetry,
  onDownload
}: {
  url: string;
  filename: string;
  onShown: () => void;
  onRetry: () => void;
  onDownload: () => void;
}) {
  const [zoom, setZoom] = useState(1);
  const [loaded, setLoaded] = useState(false);
  const [failed, setFailed] = useState(false);

  return (
    <div className='flex h-full min-h-0 flex-col'>
      <ViewerToolbar zoom={zoom} onZoom={setZoom} />
      {/* Região rolável precisa de foco para rolar pelo teclado (WCAG 2.1.1). */}
      {/* oxlint-disable-next-line jsx-a11y/no-noninteractive-tabindex */}
      <div className='bg-muted/40 min-h-0 flex-1 overflow-auto p-4' tabIndex={0}>
        {failed ? (
          <ViewerError
            message='Não foi possível exibir a imagem. O link pode ter expirado.'
            onRetry={onRetry}
            onDownload={onDownload}
          />
        ) : (
          <>
            {!loaded && <Skeleton className='mx-auto h-96 w-full max-w-3xl' />}
            {/* eslint-disable-next-line @next/next/no-img-element -- URL temporária do armazenamento, sem otimização do Next */}
            <img
              src={url}
              alt={`Imagem anexada: ${filename}`}
              onLoad={() => {
                setLoaded(true);
                onShown();
              }}
              onError={() => setFailed(true)}
              style={{ width: `${zoom * 100}%`, maxWidth: zoom === 1 ? '100%' : 'none' }}
              className={loaded ? 'mx-auto h-auto bg-white shadow-sm' : 'hidden'}
            />
          </>
        )}
      </div>
    </div>
  );
}
