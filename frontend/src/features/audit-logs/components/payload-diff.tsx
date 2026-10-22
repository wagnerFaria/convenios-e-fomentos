import { cn } from '@/lib/utils';

const show = (v: unknown) => (v === undefined ? '—' : JSON.stringify(v, null, 2));
const same = (a: unknown, b: unknown) => JSON.stringify(a) === JSON.stringify(b);

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v);
}

/** Antes e depois lado a lado, por campo, destacando o que mudou. */
export function PayloadDiff({ before, after }: { before?: unknown; after?: unknown }) {
  if (before === undefined && after === undefined) {
    return <p className='text-muted-foreground text-sm'>Sem detalhes registrados.</p>;
  }
  const b = isRecord(before) ? before : undefined;
  const a = isRecord(after) ? after : undefined;
  // Estados que não são objetos simples: compara como um bloco só.
  if ((before !== undefined && !b) || (after !== undefined && !a)) {
    return (
      <div className='grid gap-3 sm:grid-cols-2'>
        <Block title='Antes' value={before} />
        <Block title='Depois' value={after} />
      </div>
    );
  }
  const keys = [...new Set([...Object.keys(b ?? {}), ...Object.keys(a ?? {})])];
  return (
    <div className='grid gap-2'>
      <div className='text-muted-foreground grid grid-cols-[8rem_1fr_1fr] gap-2 text-xs font-medium'>
        <span>Campo</span>
        <span>Antes</span>
        <span>Depois</span>
      </div>
      {keys.map((k) => {
        const changed = !same(b?.[k], a?.[k]);
        return (
          <div
            key={k}
            className={cn(
              'grid grid-cols-[8rem_1fr_1fr] gap-2 rounded-md p-2 text-xs',
              changed && 'bg-amber-500/10 ring-1 ring-amber-500/30'
            )}
          >
            <span className='font-medium break-words'>{k}</span>
            <pre className='font-mono break-words whitespace-pre-wrap'>{show(b?.[k])}</pre>
            <pre className='font-mono break-words whitespace-pre-wrap'>{show(a?.[k])}</pre>
          </div>
        );
      })}
    </div>
  );
}

function Block({ title, value }: { title: string; value: unknown }) {
  return (
    <div className='grid gap-1'>
      <span className='text-muted-foreground text-xs font-medium'>{title}</span>
      <pre className='bg-muted/40 rounded-md p-2 font-mono text-xs break-words whitespace-pre-wrap'>
        {show(value)}
      </pre>
    </div>
  );
}
