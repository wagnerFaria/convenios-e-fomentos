import { projectActions } from './api/types';
import type { ChecklistItem, Complement, Project, SectionKey } from './api/types';

/**
 * O que o proponente pode alterar agora (espelho da regra do servidor, que é a autoridade): no rascunho, tudo; depois do envio, só o que a
 * equipe devolveu, enquanto o pedido de complementação está aberto e dentro do prazo; no mais, nada.
 */
export interface Access {
  /** Rascunho: tudo editável. */
  draft: boolean;
  /** Pedido de complementação aberto e dentro do prazo. */
  complement: Complement | null;
  /** A etapa pode ser alterada (no todo ou em parte). */
  step: (s: SectionKey) => boolean;
  /** O documento do checklist pode ser alterado. */
  document: (item: ChecklistItem) => boolean;
  /** Documentos adicionais: criar (locked = false) ou alterar um existente. Não dependem de a equipe ter devolvido algo. */
  extra: (locked: boolean) => boolean;
}

const sectionTarget: Partial<Record<SectionKey, string>> = {
  proponent: 'proponent',
  institution: 'institution',
  proposal: 'proposal',
  funding: 'funding',
  documents: 'document-options'
};

export function documentKey(item: Pick<ChecklistItem, 'code' | 'amendmentId'>) {
  return item.amendmentId ? `${item.code}@${item.amendmentId}` : item.code;
}

export function accessOf(project: Project, base: boolean, now: number): Access {
  const c = project.complement;
  const open =
    base &&
    project.status === 'submitted' &&
    project.stage === 'awaiting_proponent' &&
    c?.status === 'open' &&
    !c.overdue &&
    new Date(c.dueAt).getTime() > now
      ? c
      : null;
  const returned = new Set(open?.items.map((i) => `${i.kind}:${i.target}`));
  const draft = base && project.status === 'draft';
  return {
    draft,
    complement: open,
    step: (s) => {
      if (draft) return true;
      if (!open) return false;
      if (s === 'documents') {
        return (
          returned.has('section:document-options') ||
          [...returned].some((k) => k.startsWith('document:')) ||
          project.documents.missing > 0
        );
      }
      if (s === 'submission') return true;
      const target = sectionTarget[s];
      return target ? returned.has(`section:${target}`) : false;
    },
    document: (item) =>
      draft ||
      (!!open && (returned.has(`document:${documentKey(item)}`) || item.status === 'missing')),
    extra: (locked) => draft || (!!open && !locked)
  };
}

// ---- Quem é o quê, para o menu e as páginas ----

/** O que os critérios abaixo precisam da sessão (`AdminMe` serve). */
interface SessionLike {
  isSuperAdmin: boolean;
  permissions: { action: string; conditions: string[] }[];
}

/**
 * Condições da leitura de projetos que só alcançam "os meus": o criador e os membros convidados. Quem lê projetos com qualquer outra
 * condição (ou sem nenhuma) é da equipe. Os nomes são os do backend (`backend/internal/project/members.go` e `visibility.go`): se mudarem
 * lá, mudam aqui. Serve só para esconder menu; a autorização real é do servidor.
 */
const ownProjectConditions = new Set([
  'admin::is-creator',
  'admin::project-member',
  'admin::project-editor'
]);

/** Equipe de projetos (Analista, Revisor, Secretário, Supervisor): lê projetos além dos que cadastrou. O Super Admin também. */
export function isProjectStaff(session: SessionLike): boolean {
  if (session.isSuperAdmin) return true;
  return session.permissions.some(
    (p) =>
      p.action === projectActions.read &&
      (p.conditions.length === 0 || p.conditions.some((c) => !ownProjectConditions.has(c)))
  );
}

/** Quem pode cadastrar projetos (Proponente). O Super Admin também. */
export function isProjectProponent(session: SessionLike): boolean {
  return (
    session.isSuperAdmin || session.permissions.some((p) => p.action === projectActions.create)
  );
}
