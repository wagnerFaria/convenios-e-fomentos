'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { LoadingButton } from '@/components/ui/loading-button';
import { Textarea } from '@/components/ui/textarea';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { projectKeys } from '../api/queries';
import { duplicateProject, withdrawProject } from '../api/service';
import { projectActions, type Project } from '../api/types';

/** Duplica o projeto num novo rascunho; os documentos só são copiados se o proponente marcar (certidões vencem). */
export function DuplicateButton({
  project,
  variant = 'outline'
}: {
  project: Pick<Project, 'id'> & { isOwner: boolean };
  variant?: 'outline' | 'ghost';
}) {
  const can = useCan();
  const router = useRouter();
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [copyDocs, setCopyDocs] = useState(false);
  const duplicate = useMutation({
    mutationFn: () => duplicateProject(project.id, copyDocs),
    onSuccess: async (created) => {
      await qc.invalidateQueries({ queryKey: projectKeys.lists() });
      toast.success('Novo rascunho criado a partir do projeto.');
      setOpen(false);
      router.push(`/dashboard/projetos/${created.id}`);
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  if (!project.isOwner || !can(projectActions.create)) return null;
  return (
    <>
      <Button size='sm' variant={variant} onClick={() => setOpen(true)}>
        <Icons.copy aria-hidden /> Duplicar
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Duplicar projeto</DialogTitle>
            <DialogDescription>
              Cria um novo rascunho com os mesmos dados (instituição, proposta, recursos e emendas).
              O projeto original não muda. Você poderá ajustar tudo antes de enviar.
            </DialogDescription>
          </DialogHeader>
          <label className='flex items-start gap-3 text-sm'>
            <input
              type='checkbox'
              className='mt-1 size-4'
              checked={copyDocs}
              onChange={(e) => setCopyDocs(e.target.checked)}
            />
            <span>
              Copiar também os documentos anexados e as justificativas
              <span className='text-muted-foreground block'>
                Certidões e comprovantes costumam vencer; copie só se ainda valem.
              </span>
            </span>
          </label>
          <DialogFooter>
            <Button variant='outline' onClick={() => setOpen(false)}>
              Cancelar
            </Button>
            <LoadingButton loading={duplicate.isPending} onClick={() => duplicate.mutate()}>
              Criar rascunho
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

/** Desistência do projeto já enviado: arquiva, com o motivo. Não há volta. */
export function WithdrawButton({ project }: { project: Project }) {
  const can = useCan();
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [note, setNote] = useState('');
  const withdraw = useMutation({
    mutationFn: () => withdrawProject(project.id, note),
    onSuccess: async (saved) => {
      qc.setQueryData(projectKeys.detail(project.id), saved);
      await qc.invalidateQueries({ queryKey: projectKeys.lists() });
      toast.success('Projeto arquivado.');
      setOpen(false);
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  if (!project.isOwner || project.status !== 'submitted' || !can(projectActions.withdraw))
    return null;
  const valid = note.trim().length >= 3;
  return (
    <>
      <Button size='sm' variant='outline' onClick={() => setOpen(true)}>
        <Icons.close aria-hidden /> Desistir do projeto
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Desistir do projeto {project.code}?</DialogTitle>
            <DialogDescription>
              O projeto será arquivado e <strong>não poderá ser reaberto</strong>. Se quiser tentar
              de novo depois, use Duplicar projeto para começar um novo com os mesmos dados.
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-2'>
            <label className='text-sm font-medium' htmlFor='withdraw-note'>
              Motivo da desistência *
            </label>
            <Textarea
              id='withdraw-note'
              rows={3}
              maxLength={1000}
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setOpen(false)}>
              Voltar
            </Button>
            <LoadingButton
              variant='destructive'
              loading={withdraw.isPending}
              disabled={!valid}
              onClick={() => withdraw.mutate()}
            >
              Arquivar projeto
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
