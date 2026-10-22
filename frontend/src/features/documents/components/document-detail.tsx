'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { Icons } from '@/components/icons';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { buttonVariants } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { ApiError } from '@/lib/api/errors';
import { documentQueryOptions } from '../api/queries';
import { DocumentFiles } from './document-files';

export function DocumentDetail({ id }: { id: string }) {
  const doc = useQuery(documentQueryOptions(id));

  if (doc.isLoading) return <Skeleton className='h-40 w-full max-w-3xl' />;
  if (doc.isError || !doc.data) {
    const err = doc.error;
    return (
      <Alert variant='destructive' className='max-w-3xl'>
        <AlertDescription>
          {err instanceof ApiError && err.status === 404
            ? 'Documento não encontrado.'
            : err instanceof ApiError && err.status === 403
              ? 'Você não tem acesso a este documento.'
              : 'Não foi possível carregar o documento.'}
        </AlertDescription>
      </Alert>
    );
  }

  const d = doc.data;
  return (
    <div className='grid max-w-3xl gap-4'>
      <div>
        <Link
          href='/dashboard/documentos'
          className={buttonVariants({ variant: 'ghost', size: 'sm' })}
        >
          <Icons.chevronLeft aria-hidden /> Documentos
        </Link>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            {d.titulo ?? 'Documento'}
            {d.status && (
              <Badge variant={d.status === 'publicado' ? 'default' : 'secondary'}>
                {d.status === 'publicado' ? 'Publicado' : 'Rascunho'}
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        {d.conteudo && (
          <CardContent>
            <p className='text-sm whitespace-pre-wrap'>{d.conteudo}</p>
          </CardContent>
        )}
      </Card>
      <DocumentFiles documentId={id} />
    </div>
  );
}
