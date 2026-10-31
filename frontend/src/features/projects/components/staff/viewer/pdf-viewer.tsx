'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Document, Page, pdfjs } from 'react-pdf';
import 'react-pdf/dist/Page/TextLayer.css';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { ViewerError } from './viewer-error';
import { ViewerToolbar } from './viewer-toolbar';

// O worker precisa ser definido neste mesmo módulo, o que renderiza <Document> e <Page> (ver o README do react-pdf). Fica no pacote,
// sem CDN: o servidor do governo não deve depender de um terceiro para abrir documentos.
pdfjs.GlobalWorkerOptions.workerSrc = new URL(
  'pdfjs-dist/build/pdf.worker.min.mjs',
  import.meta.url
).toString();

/** Margem (em px) em volta da área visível em que as páginas ficam desenhadas; as de fora saem da memória. */
const RENDER_MARGIN = 1200;
const PAGE_GAP = 12;
const PADDING = 16;

function failureMessage(err: Error): string {
  switch (err.name) {
    case 'PasswordException':
      return 'Este PDF é protegido por senha e não pode ser exibido. Baixe o arquivo para conferir.';
    case 'InvalidPDFException':
      return 'O arquivo não é um PDF válido ou está corrompido. Baixe o arquivo para conferir.';
    case 'MissingPDFException':
    case 'UnexpectedResponseException':
      return 'Não foi possível abrir o arquivo. O link pode ter expirado.';
    default:
      return 'Não foi possível exibir o documento.';
  }
}

interface LazyPageProps {
  pageNumber: number;
  width: number;
  /** Altura estimada enquanto a página não foi desenhada (proporção da primeira página). */
  estimatedHeight: number;
  root: HTMLElement | null;
  onRendered: (pageNumber: number) => void;
}

/** Desenha a página só quando ela está perto da área visível; fora dela, guarda o lugar com a altura medida. */
function LazyPage({ pageNumber, width, estimatedHeight, root, onRendered }: LazyPageProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [near, setNear] = useState(pageNumber <= 2);
  const [measured, setMeasured] = useState<{ width: number; height: number }>();

  useEffect(() => {
    const el = ref.current;
    if (!el || !root) return;
    const io = new IntersectionObserver(([entry]) => setNear(entry.isIntersecting), {
      root,
      rootMargin: `${RENDER_MARGIN}px 0px`
    });
    io.observe(el);
    return () => io.disconnect();
  }, [root]);

  const placeholder = measured?.width === width ? measured.height : estimatedHeight;
  return (
    <div
      ref={ref}
      data-page={pageNumber}
      style={{ width, minHeight: near ? undefined : placeholder }}
      className='bg-white shadow-sm'
    >
      {near && (
        <Page
          pageNumber={pageNumber}
          width={width}
          renderAnnotationLayer={false}
          loading={<Skeleton style={{ width, height: placeholder }} className='rounded-none' />}
          onRenderSuccess={() => {
            const h = ref.current?.offsetHeight;
            if (h) setMeasured({ width, height: h });
            onRendered(pageNumber);
          }}
        />
      )}
    </div>
  );
}

/**
 * PDF em páginas contínuas, com zoom e navegação. `onShown` avisa uma vez que a primeira página foi desenhada (é o que conta como "ter
 * visto" o arquivo). A leitura é aos poucos (requisições com Range), então PDFs grandes abrem sem carregar tudo.
 */
export function PdfViewer({
  url,
  onShown,
  onRetry,
  onDownload
}: {
  url: string;
  onShown: () => void;
  onRetry: () => void;
  onDownload: () => void;
}) {
  const [root, setRoot] = useState<HTMLDivElement | null>(null);
  const [areaWidth, setAreaWidth] = useState(0);
  const [zoom, setZoom] = useState(1);
  const [pages, setPages] = useState(0);
  const [ratio, setRatio] = useState(Math.SQRT2);
  const [current, setCurrent] = useState(1);
  const [error, setError] = useState<string | null>(null);
  const shown = useRef(false);
  const file = useMemo(() => ({ url }), [url]);

  useEffect(() => {
    if (!root) return;
    const ro = new ResizeObserver(([entry]) => setAreaWidth(entry.contentRect.width));
    ro.observe(root);
    return () => ro.disconnect();
  }, [root]);

  // Página mais próxima do topo da área de leitura.
  const track = useCallback(() => {
    if (!root) return;
    const top = root.getBoundingClientRect().top + 8;
    let best = 1;
    for (const el of root.querySelectorAll<HTMLElement>('[data-page]')) {
      if (el.getBoundingClientRect().top - top > root.clientHeight / 3) break;
      best = Number(el.dataset.page);
    }
    setCurrent(best);
  }, [root]);

  const goTo = (page: number) => {
    const target = Math.min(Math.max(page, 1), pages);
    root?.querySelector<HTMLElement>(`[data-page="${target}"]`)?.scrollIntoView({ block: 'start' });
    setCurrent(target);
  };

  const handleRendered = useCallback(
    (pageNumber: number) => {
      if (pageNumber === 1 && !shown.current) {
        shown.current = true;
        onShown();
      }
    },
    [onShown]
  );

  const width = Math.max(200, (areaWidth - PADDING * 2) * zoom);

  return (
    <div className='flex h-full min-h-0 flex-col'>
      <ViewerToolbar zoom={zoom} onZoom={setZoom}>
        <Button
          size='icon-sm'
          variant='ghost'
          aria-label='Página anterior'
          disabled={current <= 1}
          onClick={() => goTo(current - 1)}
        >
          <Icons.chevronUp aria-hidden />
        </Button>
        <Button
          size='icon-sm'
          variant='ghost'
          aria-label='Próxima página'
          disabled={current >= pages}
          onClick={() => goTo(current + 1)}
        >
          <Icons.chevronDown aria-hidden />
        </Button>
        <span className='text-muted-foreground px-1 text-sm tabular-nums' aria-live='polite'>
          {pages > 0 ? `Página ${current} de ${pages}` : 'Carregando…'}
        </span>
      </ViewerToolbar>
      <div
        ref={setRoot}
        onScroll={track}
        // Região rolável precisa de foco para rolar pelo teclado (WCAG 2.1.1).
        // oxlint-disable-next-line jsx-a11y/no-noninteractive-tabindex
        tabIndex={0}
        aria-label='Conteúdo do documento'
        className='bg-muted/40 min-h-0 flex-1 overflow-auto'
      >
        {error ? (
          <ViewerError message={error} onRetry={onRetry} onDownload={onDownload} />
        ) : (
          <Document
            file={file}
            suspense={false}
            loading={<Skeleton className='mx-auto my-4 h-[70vh] w-11/12' />}
            error={null}
            noData={null}
            onLoadSuccess={async (pdf) => {
              setPages(pdf.numPages);
              const first = await pdf.getPage(1);
              const v = first.getViewport({ scale: 1 });
              setRatio(v.height / v.width);
            }}
            onLoadError={(err) => setError(failureMessage(err))}
            onPassword={() =>
              setError(failureMessage(Object.assign(new Error(), { name: 'PasswordException' })))
            }
          >
            <div
              className='mx-auto flex w-fit min-w-full flex-col items-center py-4'
              style={{ gap: PAGE_GAP }}
            >
              {areaWidth > 0 &&
                Array.from({ length: pages }, (_, i) => (
                  <LazyPage
                    key={i + 1}
                    pageNumber={i + 1}
                    width={width}
                    estimatedHeight={width * ratio}
                    root={root}
                    onRendered={handleRendered}
                  />
                ))}
            </div>
          </Document>
        )}
      </div>
    </div>
  );
}
