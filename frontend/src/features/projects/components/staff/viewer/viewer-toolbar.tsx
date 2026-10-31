'use client';

import type { ReactNode } from 'react';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';

/** Níveis de zoom (1 = ajustado à largura da área de leitura). */
export const ZOOM_STEPS = [0.5, 0.75, 1, 1.25, 1.5, 2, 3];

export function nextZoom(current: number, direction: 1 | -1) {
  const i = ZOOM_STEPS.findIndex((z) => z >= current);
  const base = i === -1 ? ZOOM_STEPS.length - 1 : i;
  return ZOOM_STEPS[Math.min(ZOOM_STEPS.length - 1, Math.max(0, base + direction))];
}

/** Barra de zoom do visualizador; `children` entra antes dos controles (navegação de páginas). */
export function ViewerToolbar({
  zoom,
  onZoom,
  children
}: {
  zoom: number;
  onZoom: (zoom: number) => void;
  children?: ReactNode;
}) {
  return (
    <div className='bg-background flex flex-wrap items-center justify-between gap-2 border-b px-3 py-1.5'>
      <div className='flex items-center gap-1'>{children}</div>
      <div className='flex items-center gap-1'>
        <Button
          size='icon-sm'
          variant='ghost'
          aria-label='Diminuir o zoom'
          disabled={zoom <= ZOOM_STEPS[0]}
          onClick={() => onZoom(nextZoom(zoom, -1))}
        >
          <Icons.zoomOut aria-hidden />
        </Button>
        <Button
          size='sm'
          variant='ghost'
          className='min-w-14 tabular-nums'
          aria-label='Ajustar à largura'
          onClick={() => onZoom(1)}
        >
          {Math.round(zoom * 100)}%
        </Button>
        <Button
          size='icon-sm'
          variant='ghost'
          aria-label='Aumentar o zoom'
          disabled={zoom >= ZOOM_STEPS[ZOOM_STEPS.length - 1]}
          onClick={() => onZoom(nextZoom(zoom, 1))}
        >
          <Icons.zoomIn aria-hidden />
        </Button>
      </div>
    </div>
  );
}
