'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { useState } from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { useDebounce } from '@/hooks/use-debounce';
import { documentsQueryOptions } from '../api/queries';

export function DocumentsList() {
  const [q, setQ] = useState('');
  const [page, setPage] = useState(1);
  const search = useDebounce(q, 300);
  const docs = useQuery({
    ...documentsQueryOptions(search, page),
    placeholderData: (prev) => prev
  });

  return (
    <div className='grid gap-4'>
      <Input
        value={q}
        onChange={(e) => {
          setQ(e.target.value);
          setPage(1);
        }}
        placeholder='Buscar por título ou conteúdo…'
        aria-label='Buscar documentos'
        className='max-w-sm'
      />
      {docs.isError ? (
        <Alert variant='destructive'>
          <AlertDescription>Não foi possível carregar os documentos.</AlertDescription>
        </Alert>
      ) : docs.isLoading ? (
        <Skeleton className='h-40 w-full' />
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Título</TableHead>
                <TableHead>Situação</TableHead>
                <TableHead>Criado em</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {docs.data?.data.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className='text-muted-foreground text-center'>
                    Nenhum documento encontrado.
                  </TableCell>
                </TableRow>
              )}
              {docs.data?.data.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className='font-medium'>
                    <Link href={`/dashboard/documentos/${d.id}`} className='hover:underline'>
                      {d.titulo ?? d.id}
                    </Link>
                  </TableCell>
                  <TableCell>
                    {d.status && (
                      <Badge variant={d.status === 'publicado' ? 'default' : 'secondary'}>
                        {d.status === 'publicado' ? 'Publicado' : 'Rascunho'}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>{new Date(d.created_at).toLocaleDateString('pt-BR')}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {docs.data?.meta && docs.data.meta.pageCount > 1 && (
            <div className='flex items-center justify-end gap-2 text-sm'>
              <span className='text-muted-foreground'>
                Página {docs.data.meta.page} de {docs.data.meta.pageCount}
              </span>
              <Button
                size='sm'
                variant='outline'
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
              >
                Anterior
              </Button>
              <Button
                size='sm'
                variant='outline'
                disabled={page >= docs.data.meta.pageCount}
                onClick={() => setPage(page + 1)}
              >
                Próxima
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
