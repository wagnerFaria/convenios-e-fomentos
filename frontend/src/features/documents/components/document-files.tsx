'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useRef, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan } from '@/features/auth/session-provider';
import { ApiError } from '@/lib/api/errors';
import { cn, formatBytes } from '@/lib/utils';
import { documentFilesQueryOptions, documentKeys } from '../api/queries';
import { confirmUpload, deleteFile, getDownload, requestUpload } from '../api/service';
import { documentActions, type DocumentFile } from '../api/types';
import {
  ACCEPT_ATTR,
  MAX_FILES_PER_DOCUMENT,
  contentTypeOf,
  validateFile
} from '../lib/file-rules';
import { uploadToStorage } from '../lib/upload';

interface Transfer {
  id: string;
  name: string;
  percent: number;
  stage: 'enviando' | 'confirmando' | 'erro';
  error?: string;
}

function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    const field = err.fieldErrors[0];
    return field ? field.message : err.message;
  }
  return err instanceof Error ? err.message : 'Erro inesperado.';
}

function FileIcon({ type }: { type: string }) {
  const Icon =
    type === 'application/pdf'
      ? Icons.fileTypePdf
      : type.startsWith('image/')
        ? Icons.media
        : Icons.fileTypeDoc;
  return <Icon className='text-muted-foreground size-5 shrink-0' aria-hidden />;
}

export function DocumentFiles({ documentId }: { documentId: string }) {
  const can = useCan();
  const qc = useQueryClient();
  const files = useQuery({
    ...documentFilesQueryOptions(documentId),
    enabled: can(documentActions.fileRead)
  });
  const count = files.data?.length ?? 0;
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [toDelete, setToDelete] = useState<DocumentFile | null>(null);
  const seq = useRef(0);

  const refresh = useCallback(
    () => qc.invalidateQueries({ queryKey: documentKeys.files(documentId) }),
    [qc, documentId]
  );
  const patch = useCallback(
    (id: string, change: Partial<Transfer>) =>
      setTransfers((all) => all.map((t) => (t.id === id ? { ...t, ...change } : t))),
    []
  );

  /** Fluxo completo de um arquivo: autoriza → envia ao armazenamento → confirma. */
  const upload = useCallback(
    async (file: File) => {
      const id = `t${++seq.current}`;
      setTransfers((all) => [...all, { id, name: file.name, percent: 0, stage: 'enviando' }]);
      let fileId: string | undefined;
      try {
        const problem = validateFile(file);
        if (problem) throw new Error(problem);
        const grant = await requestUpload(documentId, {
          filename: file.name,
          content_type: contentTypeOf(file)!,
          size_bytes: file.size
        });
        fileId = grant.file.id;
        await uploadToStorage(grant.upload, file, (percent) => patch(id, { percent }));
        patch(id, { percent: 100, stage: 'confirmando' });
        await confirmUpload(documentId, fileId);
        setTransfers((all) => all.filter((t) => t.id !== id));
      } catch (err) {
        patch(id, { stage: 'erro', error: errorMessage(err) });
        // Não deixa um registro pendente para trás (o cleanup do servidor também o removeria).
        // Uma recusa na confirmação (422) já apagou objeto e registro no servidor.
        const rejectedByServer = err instanceof ApiError && err.status === 422;
        if (fileId && !rejectedByServer)
          await deleteFile(documentId, fileId).catch(() => undefined);
      } finally {
        await refresh();
      }
    },
    [documentId, refresh, patch]
  );

  const onDrop = useCallback(
    (accepted: File[]) => {
      const room = MAX_FILES_PER_DOCUMENT - count;
      if (accepted.length > room) {
        toast.error(`O documento comporta mais ${Math.max(room, 0)} arquivo(s).`);
        accepted = accepted.slice(0, Math.max(room, 0));
      }
      accepted.forEach((f) => void upload(f));
    },
    [count, upload]
  );

  const dropzone = useDropzone({
    onDrop,
    // A validação por extensão/tamanho acontece em validateFile, com mensagens por arquivo.
    noClick: false,
    multiple: true,
    disabled: !can(documentActions.fileUpload)
  });

  const download = useMutation({
    mutationFn: async (f: DocumentFile) => (await getDownload(documentId, f.id)).url,
    onSuccess: (url) => {
      const a = document.createElement('a');
      a.href = url;
      a.rel = 'noopener';
      a.click();
    },
    onError: (err) => toast.error(errorMessage(err))
  });

  const remove = useMutation({
    mutationFn: (f: DocumentFile) => deleteFile(documentId, f.id),
    onSuccess: () => {
      toast.success('Arquivo excluído.');
      setToDelete(null);
      void refresh();
    },
    onError: (err) => toast.error(errorMessage(err))
  });

  if (!can(documentActions.fileRead)) return null;

  const list = files.data ?? [];
  return (
    <Card>
      <CardHeader>
        <CardTitle>Arquivos</CardTitle>
        <CardDescription>
          PDF e DOCX até 200 MB; PNG, JPEG e WebP até 25 MB. Até {MAX_FILES_PER_DOCUMENT} arquivos.
        </CardDescription>
      </CardHeader>
      <CardContent className='grid gap-4'>
        {can(documentActions.fileUpload) && (
          <div
            {...dropzone.getRootProps()}
            className={cn(
              'border-muted-foreground/25 hover:bg-muted/40 grid cursor-pointer place-items-center gap-1 rounded-lg border-2 border-dashed px-4 py-8 text-center transition',
              dropzone.isDragActive && 'border-primary bg-muted/60'
            )}
          >
            <input {...dropzone.getInputProps()} accept={ACCEPT_ATTR} data-testid='file-input' />
            <Icons.upload className='text-muted-foreground size-8' aria-hidden />
            <p className='text-sm font-medium'>
              {dropzone.isDragActive
                ? 'Solte os arquivos aqui'
                : 'Arraste arquivos ou clique para selecionar'}
            </p>
          </div>
        )}

        {transfers.length > 0 && (
          <ul className='grid gap-2' aria-label='Envios em andamento'>
            {transfers.map((t) => (
              <li key={t.id} className='grid gap-1 rounded-md border p-3 text-sm'>
                <div className='flex items-center justify-between gap-2'>
                  <span className='truncate font-medium'>{t.name}</span>
                  {t.stage === 'erro' ? (
                    <Button
                      size='xs'
                      variant='ghost'
                      onClick={() => setTransfers((all) => all.filter((x) => x.id !== t.id))}
                    >
                      Dispensar
                    </Button>
                  ) : (
                    <span className='text-muted-foreground text-xs'>
                      {t.stage === 'confirmando' ? 'Confirmando…' : `${t.percent}%`}
                    </span>
                  )}
                </div>
                {t.stage === 'erro' ? (
                  <Alert variant='destructive'>
                    <AlertDescription>{t.error}</AlertDescription>
                  </Alert>
                ) : (
                  <Progress value={t.percent} aria-label={`Envio de ${t.name}`} />
                )}
              </li>
            ))}
          </ul>
        )}

        {files.isLoading ? (
          <Skeleton className='h-16 w-full' />
        ) : files.isError ? (
          <Alert variant='destructive'>
            <AlertDescription>{errorMessage(files.error)}</AlertDescription>
          </Alert>
        ) : list.length === 0 ? (
          <p className='text-muted-foreground text-sm'>Nenhum arquivo anexado.</p>
        ) : (
          <ul className='divide-y rounded-md border' aria-label='Arquivos anexados'>
            {list.map((f) => (
              <li key={f.id} className='flex items-center gap-3 p-3 text-sm'>
                <FileIcon type={f.content_type} />
                <div className='min-w-0 flex-1'>
                  <p className='truncate font-medium'>{f.filename}</p>
                  <p className='text-muted-foreground text-xs'>{formatBytes(f.size_bytes)}</p>
                </div>
                {f.status === 'pendente' ? (
                  <Badge variant='outline'>Pendente</Badge>
                ) : (
                  <Button
                    size='sm'
                    variant='outline'
                    disabled={download.isPending}
                    onClick={() => download.mutate(f)}
                    aria-label={`Baixar ${f.filename}`}
                  >
                    <Icons.arrowRight className='rotate-90' aria-hidden /> Baixar
                  </Button>
                )}
                {can(documentActions.fileDelete) && (
                  <Button
                    size='icon-sm'
                    variant='ghost'
                    onClick={() => setToDelete(f)}
                    aria-label={`Excluir ${f.filename}`}
                  >
                    <Icons.trash aria-hidden />
                  </Button>
                )}
              </li>
            ))}
          </ul>
        )}
      </CardContent>
      <AlertModal
        isOpen={toDelete !== null}
        onClose={() => setToDelete(null)}
        onConfirm={() => toDelete && remove.mutate(toDelete)}
        loading={remove.isPending}
        title='Excluir arquivo?'
        description={`“${toDelete?.filename ?? ''}” será removido definitivamente.`}
        confirmLabel='Excluir'
        cancelLabel='Cancelar'
      />
    </Card>
  );
}
