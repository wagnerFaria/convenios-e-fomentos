'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { AlertModal } from '@/components/modal/alert-modal';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { LoadingButton } from '@/components/ui/loading-button';
import { NativeSelect } from '@/components/ui/native-select';
import { Skeleton } from '@/components/ui/skeleton';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { membersQueryOptions, projectKeys } from '../api/queries';
import { inviteMember, revokeMember, updateMember } from '../api/service';
import {
  memberLevelLabels,
  projectActions,
  type Member,
  type MemberLevel,
  type Project
} from '../api/types';
import { formatCpf, onlyDigits } from '../format';

const dateTime = new Intl.DateTimeFormat('pt-BR', {
  dateStyle: 'short',
  timeZone: 'America/Cuiaba'
});

function LevelSelect({
  value,
  onChange,
  disabled,
  label
}: {
  value: MemberLevel;
  onChange: (v: MemberLevel) => void;
  disabled?: boolean;
  label: string;
}) {
  return (
    <NativeSelect size='sm' className='w-32'>
      <select
        aria-label={label}
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value as MemberLevel)}
      >
        <option value='read'>{memberLevelLabels.read}</option>
        <option value='edit'>{memberLevelLabels.edit}</option>
      </select>
    </NativeSelect>
  );
}

function MemberRow({ projectId, member }: { projectId: string; member: Member }) {
  const qc = useQueryClient();
  const [confirming, setConfirming] = useState(false);
  const refresh = () => qc.invalidateQueries({ queryKey: projectKeys.members(projectId) });
  const change = useMutation({
    mutationFn: (level: MemberLevel) =>
      updateMember(projectId, member.id, { level, label: member.label }),
    onSuccess: async () => {
      await refresh();
      toast.success('Nível de acesso alterado.');
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  const revoke = useMutation({
    mutationFn: () => revokeMember(projectId, member.id),
    onSuccess: async () => {
      await refresh();
      toast.success('Acesso revogado.');
    },
    onError: (err) => toast.error(errorMessage(err)),
    onSettled: () => setConfirming(false)
  });
  return (
    <li className='flex flex-wrap items-center justify-between gap-2 py-2'>
      <div className='min-w-0'>
        <p className='truncate text-sm font-medium'>{member.label || 'Sem identificação'}</p>
        <p className='text-muted-foreground text-xs'>
          <span className='font-mono'>{member.cpf}</span>
          {member.email && <> · {member.email}</>} ·{' '}
          {member.acceptedAt ? (
            `acessou em ${dateTime.format(new Date(member.acceptedAt))}`
          ) : (
            <Badge variant='outline'>Convite pendente</Badge>
          )}
        </p>
      </div>
      <div className='flex items-center gap-1'>
        <LevelSelect
          label={`Nível de ${member.label || member.cpf}`}
          value={member.level}
          disabled={change.isPending}
          onChange={(level) => level !== member.level && change.mutate(level)}
        />
        <Button
          size='sm'
          variant='ghost'
          aria-label={`Revogar o acesso de ${member.label || member.cpf}`}
          onClick={() => setConfirming(true)}
        >
          <Icons.trash aria-hidden />
        </Button>
      </div>
      <AlertModal
        isOpen={confirming}
        onClose={() => setConfirming(false)}
        onConfirm={() => revoke.mutate()}
        loading={revoke.isPending}
        title='Revogar este acesso?'
        description='A pessoa deixa de ver o projeto imediatamente, mesmo que esteja com ele aberto. Você pode convidá-la de novo depois.'
        confirmLabel='Revogar'
        cancelLabel='Voltar'
      />
    </li>
  );
}

function InviteForm({ projectId }: { projectId: string }) {
  const qc = useQueryClient();
  const [cpf, setCpf] = useState('');
  const [label, setLabel] = useState('');
  const [email, setEmail] = useState('');
  const [level, setLevel] = useState<MemberLevel>('read');
  const invite = useMutation({
    mutationFn: () =>
      inviteMember(projectId, {
        cpf: onlyDigits(cpf),
        label: label.trim(),
        email: email.trim(),
        level
      }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: projectKeys.members(projectId) });
      setCpf('');
      setLabel('');
      setEmail('');
      toast.success('Acesso concedido.');
    },
    onError: (err) => toast.error(errorMessage(err))
  });
  const ready = onlyDigits(cpf).length === 11;
  return (
    <form
      className='space-y-3 border-t pt-4'
      onSubmit={(e) => {
        e.preventDefault();
        if (ready) invite.mutate();
      }}
    >
      <h3 className='text-sm font-medium'>Dar acesso a outra pessoa</h3>
      <div className='grid gap-2 sm:grid-cols-2'>
        <div className='space-y-1'>
          <label htmlFor='member-cpf' className='text-muted-foreground text-xs'>
            CPF *
          </label>
          <Input
            id='member-cpf'
            inputMode='numeric'
            autoComplete='off'
            placeholder='000.000.000-00'
            maxLength={14}
            value={cpf}
            onChange={(e) => setCpf(formatCpf(onlyDigits(e.target.value).slice(0, 11)))}
          />
        </div>
        <div className='space-y-1'>
          <label htmlFor='member-label' className='text-muted-foreground text-xs'>
            Identificação (opcional)
          </label>
          <Input
            id='member-label'
            autoComplete='off'
            placeholder='Ex.: Ana, contadora'
            maxLength={100}
            value={label}
            onChange={(e) => setLabel(e.target.value)}
          />
        </div>
      </div>
      <div className='space-y-1'>
        <label htmlFor='member-email' className='text-muted-foreground text-xs'>
          E-mail da pessoa (opcional)
        </label>
        <Input
          id='member-email'
          type='email'
          autoComplete='off'
          placeholder='nome@exemplo.com'
          maxLength={254}
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <p className='text-muted-foreground text-xs'>
          Usado para avisar a pessoa do convite se ela ainda não tiver cadastro no sistema. Quem já
          tem cadastro é avisado no e-mail da própria conta.
        </p>
      </div>
      <div className='flex flex-wrap items-center gap-2'>
        <LevelSelect label='Nível do novo acesso' value={level} onChange={setLevel} />
        <LoadingButton type='submit' size='sm' loading={invite.isPending} disabled={!ready}>
          <Icons.add aria-hidden /> Convidar
        </LoadingButton>
      </div>
      <p className='text-muted-foreground text-xs'>
        <strong>Leitura</strong>: vê o projeto, os documentos e o comprovante.{' '}
        <strong>Edição</strong>: também altera, anexa documentos e envia. Só você, como titular,
        pode desistir, excluir, duplicar ou gerir os acessos. A pessoa não precisa ter conta: o
        acesso vale quando ela entrar com o CPF informado.
      </p>
    </form>
  );
}

/** Botão e janela de gestão do acesso ao projeto; aparece só para o titular. */
export function AccessButton({ project }: { project: Project }) {
  const can = useCan();
  const [open, setOpen] = useState(false);
  const members = useQuery({ ...membersQueryOptions(project.id), enabled: open });
  if (project.access !== 'owner' || !can(projectActions.share)) return null;
  return (
    <>
      <Button size='sm' variant='outline' onClick={() => setOpen(true)}>
        <Icons.teams aria-hidden /> Acesso
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className='sm:max-w-xl'>
          <DialogHeader>
            <DialogTitle>Quem tem acesso a este projeto</DialogTitle>
            <DialogDescription>
              Você é o titular. Convide outras pessoas pelo CPF, para ler ou editar com você.
            </DialogDescription>
          </DialogHeader>
          {members.isError && (
            <Alert variant='destructive' role='alert'>
              <AlertDescription>{errorMessage(members.error)}</AlertDescription>
            </Alert>
          )}
          {members.isLoading ? (
            <Skeleton className='h-16 w-full' />
          ) : (
            <ul className='divide-y' aria-label='Pessoas com acesso'>
              {(members.data ?? []).length === 0 && (
                <li className='text-muted-foreground py-2 text-sm'>
                  Ninguém além de você tem acesso a este projeto.
                </li>
              )}
              {(members.data ?? []).map((m) => (
                <MemberRow key={m.id} projectId={project.id} member={m} />
              ))}
            </ul>
          )}
          <InviteForm projectId={project.id} />
        </DialogContent>
      </Dialog>
    </>
  );
}
