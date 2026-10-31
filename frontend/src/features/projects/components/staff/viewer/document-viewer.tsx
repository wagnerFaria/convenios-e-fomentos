'use client';

import dynamic from 'next/dynamic';
import { Skeleton } from '@/components/ui/skeleton';
import type { FilePreview } from '../../../api/types';
import { ImageViewer } from './image-viewer';

// O leitor de PDF usa APIs do navegador: carrega só no cliente e só quando um PDF é aberto.
const PdfViewer = dynamic(() => import('./pdf-viewer').then((m) => m.PdfViewer), {
  ssr: false,
  loading: () => <Skeleton className='mx-auto my-4 h-[70vh] w-11/12' />
});

/** Escolhe o leitor pelo tipo devolvido pelo servidor (o Word chega já convertido em PDF). */
export function DocumentViewer({
  preview,
  filename,
  onShown,
  onReachedEnd,
  onRetry,
  onDownload
}: {
  preview: FilePreview;
  filename: string;
  onShown: () => void;
  onReachedEnd: () => void;
  onRetry: () => void;
  onDownload: () => void;
}) {
  if (preview.kind === 'image') {
    return (
      <ImageViewer
        key={preview.url}
        url={preview.url}
        filename={filename}
        onShown={onShown}
        onReachedEnd={onReachedEnd}
        onRetry={onRetry}
        onDownload={onDownload}
      />
    );
  }
  return (
    <PdfViewer
      key={preview.url}
      url={preview.url}
      onShown={onShown}
      onReachedEnd={onReachedEnd}
      onRetry={onRetry}
      onDownload={onDownload}
    />
  );
}
