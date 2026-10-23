/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const projectActions = {
  read: 'admin::projects.read',
  create: 'admin::projects.create',
  update: 'admin::projects.update',
  delete: 'admin::projects.delete',
  submit: 'admin::projects.submit',
  withdraw: 'admin::projects.withdraw',
  share: 'admin::projects.share',
  profileRead: 'admin::proponent-profile.read',
  profileUpdate: 'admin::proponent-profile.update'
} as const;

export type Instrument = 'convenio' | 'fomento';
export type ProjectStatus = 'draft' | 'submitted' | 'archived';
export type ProjectStage = 'awaiting_checklist' | 'in_analysis' | 'awaiting_proponent';
export type PhoneType = 'mobile' | 'landline';
export type MemberLevel = 'read' | 'edit';
/** owner = titular; edit/read = convidado por CPF; vazio = equipe. */
export type ProjectAccess = 'owner' | MemberLevel | '';

export interface Phone {
  number: string;
  type: PhoneType;
}

export interface Email {
  address: string;
  primary: boolean;
}

/** Dados de contato do proponente, comuns ao perfil e ao projeto. */
export interface Contact {
  phones: Phone[];
  emails: Email[];
  uf: string;
  municipalityIbge: string;
  jobTitle: string;
}

export interface Profile extends Contact {
  name: string;
  cpf: string;
  complete: boolean;
}

export interface ProponentSection extends Contact {
  name: string;
  cpf: string;
  isLegalRepresentative: boolean | null;
  legalRepName: string;
  legalRepCpf: string;
  legalRepJobTitle: string;
}

export interface InstitutionSection {
  instrument: Instrument | '';
  cnpj: string;
  name: string;
  entityTypeId: string;
  /** Somente leitura: código do tipo escolhido; o tipo "Outro" pede o texto livre e a natureza. */
  entityTypeCode: string;
  entityTypeIsOther: boolean;
  entityTypeOther: string;
  nature: 'public' | 'private' | '';
  nonProfit: boolean;
  uf: string;
  municipalityIbge: string;
  zip: string;
  street: string;
  number: string;
  complement: string;
  district: string;
  email: string;
  phone: string;
  habilitationStatus: 'full' | 'partial' | '';
  habilitationNumber: string;
  habilitationValidUntil: string;
}

export interface ProjectMunicipality {
  ibgeCode: string;
  name: string;
  uf: string;
}

export type ObjectNature = 'works' | 'goods' | 'services' | 'event' | 'other';

export interface ProposalSection {
  sigconNumber: string;
  sigconRegisteredOn: string;
  title: string;
  object: string;
  justification: string;
  startDate: string;
  endDate: string;
  objectNature: ObjectNature | '';
  hasArtisticHiring: boolean;
  calamity: boolean;
  executionPlaceDetail: string;
  municipalities: ProjectMunicipality[];
}

/** Corpo de gravação da proposta: os municípios vão como lista de códigos IBGE. */
export type ProposalInput = Omit<ProposalSection, 'municipalities'> & {
  municipalityCodes: string[];
};

export type FundingType = 'amendment' | 'own' | 'mixed';
export type CounterpartKind = 'financial' | 'in_kind';
export type TenderStatus = 'held' | 'waived' | 'unfeasible' | 'not_applicable';

/** Emenda parlamentar. Informe `parliamentarianId` (cadastro) ou, se o parlamentar não estiver nele, o nome por extenso. */
export interface Amendment {
  /** Estável: reenviar o id mantém a emenda e os documentos anexados a ela. Vazio nas emendas novas. */
  id: string;
  parliamentarianId: string;
  parliamentarianName: string;
  number: string;
  year: number;
  amountCents: number;
}

/** Seção Recursos. Todos os valores são em centavos; os totais são calculados pelo servidor. */
export interface FundingSection {
  type: FundingType | '';
  ownAmountCents: number;
  amendments: Amendment[];
  amendmentTotalCents: number;
  totalCents: number;
  counterpartKind: CounterpartKind | '';
  counterpartCents: number;
  counterpartDescription: string;
  indirectCostsCents: number;
  tenderStatus: TenderStatus | '';
  tenderNoticeNumber: string;
  tenderApprovedOn: string;
  tenderLegalBasis: string;
  tenderJustification: string;
  artistFeesCents: number;
  regionalArtistFeesCents: number;
  governorWaiver: boolean;
  priceOldestCollectedOn: string;
}

/** Situações que acrescentam documentos ao checklist. */
export interface DocumentOptions {
  exclusiveSupplier: boolean;
  publicSpaceUse: boolean;
  reducedDeadline: boolean;
}

export interface DocumentSummary {
  total: number;
  done: number;
  justified: number;
  missing: number;
}

export interface DocFile {
  id: string;
  filename: string;
  contentType: string;
  sizeBytes: number;
  status: 'pending' | 'confirmed';
  /** SHA-256 do conteúdo, calculado pelo servidor na confirmação. */
  sha256?: string;
  createdAt: string;
}

export interface ChecklistItem {
  code: string;
  name: string;
  description: string;
  group: string;
  scope: 'project' | 'amendment';
  amendmentId: string;
  amendmentLabel: string;
  /** Item do checklist da SEDEC (ex.: C2, F2). */
  source: string;
  template: { name: string; url: string } | null;
  acceptsJustification: boolean;
  files: DocFile[];
  justification: string;
  status: 'missing' | 'done' | 'justified';
}

export interface Checklist {
  options: DocumentOptions;
  instrument: Instrument | '';
  summary: DocumentSummary;
  items: ChecklistItem[];
}

export interface FileRequest {
  filename: string;
  contentType: string;
  sizeBytes: number;
  amendmentId?: string;
}

export interface UploadGrant {
  url: string;
  fields: Record<string, string>;
  expires_at: string;
}

export type SectionKey =
  | 'proponent'
  | 'institution'
  | 'proposal'
  | 'funding'
  | 'documents'
  | 'submission';

export interface Issue {
  section: SectionKey;
  field: string;
  code: string;
  severity: 'block' | 'warn';
  message: string;
}

export interface Project {
  id: string;
  status: ProjectStatus;
  version: number;
  isOwner: boolean;
  /** Relação de quem consulta com o projeto; vazio para a equipe. */
  access: ProjectAccess;
  /** Código, etapa e data só existem depois do envio. */
  code: string;
  stage: ProjectStage | '';
  submittedAt: string | null;
  lateSubmission: boolean;
  complement: Complement | null;
  archive: ArchiveInfo | null;
  originProjectId: string;
  originCode: string;
  createdAt: string;
  updatedAt: string;
  proponent: ProponentSection;
  institution: InstitutionSection;
  proposal: ProposalSection;
  funding: FundingSection;
  documentOptions: DocumentOptions;
  documents: DocumentSummary;
  issues: Issue[];
}

export interface ProjectSummary {
  id: string;
  status: ProjectStatus;
  code: string;
  submittedAt: string | null;
  stage: ProjectStage | '';
  archiveReason: ArchiveReason | '';
  complementDueAt: string | null;
  version: number;
  instrument: Instrument | '';
  title: string;
  institutionName: string;
  institutionCnpj: string;
  sigconNumber: string;
  totalCents: number;
  isOwner: boolean;
  access: ProjectAccess;
  createdAt: string;
  updatedAt: string;
}

export const instrumentLabels: Record<Instrument, string> = {
  convenio: 'Convênio',
  fomento: 'Termo de Fomento'
};

export const statusLabels: Record<ProjectStatus, string> = {
  draft: 'Rascunho',
  submitted: 'Enviado',
  archived: 'Arquivado'
};

export const stageLabels: Record<ProjectStage, string> = {
  awaiting_checklist: 'Aguardando análise de checklist',
  in_analysis: 'Em análise',
  awaiting_proponent: 'Aguardando correções do proponente'
};

export type ArchiveReason =
  | 'complement_deadline'
  | 'withdrawn'
  | 'denied'
  | 'missing_documents'
  | 'other';

export const archiveReasonLabels: Record<ArchiveReason, string> = {
  complement_deadline: 'Prazo de correção vencido',
  withdrawn: 'Desistência do proponente',
  denied: 'Negado pela Secretaria',
  missing_documents: 'Falta de documentos obrigatórios',
  other: 'Decisão da equipe'
};

/** Item devolvido pela equipe: uma seção ou um documento do checklist. */
export interface ComplementItem {
  kind: 'section' | 'document';
  /** Seção: proponent, institution, proposal, funding ou document-options. Documento: código, ou código@id-da-emenda. */
  target: string;
  label: string;
  note: string;
  answered: boolean;
}

export interface Complement {
  number: number;
  status: 'open' | 'answered' | 'expired' | 'cancelled';
  requestedAt: string;
  dueAt: string;
  overdue: boolean;
  message: string;
  items: ComplementItem[];
}

export interface ArchiveInfo {
  at: string;
  reason: ArchiveReason;
  note: string;
}

export const objectNatureLabels: Record<ObjectNature, string> = {
  works: 'Obra ou serviço de engenharia',
  goods: 'Aquisição de bens',
  services: 'Contratação de serviços',
  event: 'Evento',
  other: 'Outro'
};

export const sectionLabels: Record<SectionKey, string> = {
  proponent: 'Proponente',
  institution: 'Instituição',
  proposal: 'Proposta',
  funding: 'Recursos',
  documents: 'Documentos',
  submission: 'Revisão e envio'
};

export const fundingTypeLabels: Record<FundingType, string> = {
  amendment: 'Emenda parlamentar',
  own: 'Recurso próprio da SEDEC',
  mixed: 'Emenda parlamentar e recurso próprio da SEDEC'
};

/** Pessoa com acesso ao projeto. O CPF vem mascarado e a API não informa se a pessoa tem conta. */
export interface Member {
  id: string;
  cpf: string;
  label: string;
  level: MemberLevel;
  /** E-mail que o titular informou (mascarado); vazio se não informou. */
  email: string;
  createdAt: string;
  /** Primeira vez em que a pessoa abriu o projeto; vazio = convite pendente. */
  acceptedAt: string | null;
}

export const memberLevelLabels: Record<MemberLevel, string> = {
  read: 'Leitura',
  edit: 'Edição'
};

export const accessLabels: Record<Exclude<ProjectAccess, ''>, string> = {
  owner: 'Titular',
  edit: 'Edição',
  read: 'Leitura'
};
