import { apiClient } from '@/lib/api-client';
import { apiErrorFromResponse } from '@/lib/api/errors';
import type { ApiResponse } from '@/lib/api/types';
import type {
  Checklist,
  Contact,
  Declaration,
  Member,
  MemberLevel,
  DocFile,
  DocumentOptions,
  FileRequest,
  FundingSection,
  UploadGrant,
  InstitutionSection,
  Profile,
  Project,
  ProjectSummary,
  ProponentSection,
  ProposalInput
} from './types';

const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function getProfile() {
  return (await apiClient<ApiResponse<Profile>>('/admin/proponent-profile')).data;
}

export async function saveProfile(input: Contact) {
  return (
    await apiClient<ApiResponse<Profile>>('/admin/proponent-profile', {
      method: 'PUT',
      ...json(input)
    })
  ).data;
}

export interface ListProjectsParams {
  q?: string;
  page?: number;
  pageSize?: number;
}

export async function listProjects({ q, page = 1, pageSize = 25 }: ListProjectsParams = {}) {
  const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
  if (q) params.set('q', q);
  return await apiClient<ApiResponse<ProjectSummary[]>>(`/admin/projects?${params}`);
}

export async function getProject(id: string) {
  return (await apiClient<ApiResponse<Project>>(`/admin/projects/${id}`)).data;
}

export async function createProject() {
  return (await apiClient<ApiResponse<Project>>('/admin/projects', { method: 'POST' })).data;
}

export async function deleteProject(id: string) {
  await apiClient<void>(`/admin/projects/${id}`, { method: 'DELETE' });
}

type WithVersion<T> = T & { version: number };

async function saveSection<T>(id: string, section: string, body: WithVersion<T>) {
  return (
    await apiClient<ApiResponse<Project>>(`/admin/projects/${id}/${section}`, {
      method: 'PUT',
      ...json(body)
    })
  ).data;
}

export const saveProponent = (
  id: string,
  body: WithVersion<Omit<ProponentSection, 'name' | 'cpf'>>
) => saveSection(id, 'proponent', body);

export const saveInstitution = (
  id: string,
  body: WithVersion<Omit<InstitutionSection, 'entityTypeIsOther' | 'entityTypeCode'>>
) => saveSection(id, 'institution', body);

export const saveFunding = (
  id: string,
  body: WithVersion<Omit<FundingSection, 'amendmentTotalCents' | 'totalCents'>>
) => saveSection(id, 'funding', body);

export const saveProposal = (id: string, body: WithVersion<ProposalInput>) =>
  saveSection(id, 'proposal', body);

export const saveDocumentOptions = (id: string, body: WithVersion<DocumentOptions>) =>
  saveSection(id, 'document-options', body);

const docs = (id: string) => `/admin/projects/${id}/documents`;

export async function getChecklist(id: string) {
  return (await apiClient<ApiResponse<Checklist>>(docs(id))).data;
}

/** Passo 1 do envio: registra o arquivo e obtém a autorização para enviá-lo ao armazenamento. */
export async function requestFileUpload(id: string, code: string, input: FileRequest) {
  return (
    await apiClient<ApiResponse<{ file: DocFile; upload: UploadGrant }>>(
      `${docs(id)}/${code}/files`,
      {
        method: 'POST',
        ...json(input)
      }
    )
  ).data;
}

/** Passo 3: a API confere o objeto no armazenamento e confirma o arquivo. */
export async function confirmFileUpload(id: string, fileId: string) {
  return (
    await apiClient<ApiResponse<DocFile>>(`${docs(id)}/files/${fileId}/confirm`, { method: 'POST' })
  ).data;
}

export async function getFileDownload(id: string, fileId: string) {
  return (
    await apiClient<ApiResponse<{ url: string; expiresAt: string }>>(
      `${docs(id)}/files/${fileId}/download`
    )
  ).data;
}

export async function deleteProjectFile(id: string, fileId: string) {
  await apiClient<void>(`${docs(id)}/files/${fileId}`, { method: 'DELETE' });
}

/** Cria o documento adicional (sem arquivo ainda); devolve o checklist atualizado. */
export async function createExtraDocument(
  id: string,
  input: { title: string; description: string }
) {
  return (
    await apiClient<ApiResponse<Checklist>>(`${docs(id)}/extra`, { method: 'POST', ...json(input) })
  ).data;
}

export async function updateExtraDocument(
  id: string,
  itemId: string,
  input: { title: string; description: string }
) {
  return (
    await apiClient<ApiResponse<Checklist>>(`${docs(id)}/extra/${itemId}`, {
      method: 'PUT',
      ...json(input)
    })
  ).data;
}

/** Remove o documento adicional e o arquivo dele. */
export async function deleteExtraDocument(id: string, itemId: string) {
  return (
    await apiClient<ApiResponse<Checklist>>(`${docs(id)}/extra/${itemId}`, { method: 'DELETE' })
  ).data;
}

/** Passo 1 do envio do arquivo de um documento adicional. */
export async function requestExtraFileUpload(id: string, itemId: string, input: FileRequest) {
  return (
    await apiClient<ApiResponse<{ file: DocFile; upload: UploadGrant }>>(
      `${docs(id)}/extra/${itemId}/files`,
      { method: 'POST', ...json(input) }
    )
  ).data;
}

export async function saveJustification(
  id: string,
  code: string,
  input: { amendmentId?: string; text: string }
) {
  return (
    await apiClient<ApiResponse<Checklist>>(`${docs(id)}/${code}/justification`, {
      method: 'PUT',
      ...json(input)
    })
  ).data;
}

export async function submitProject(id: string, input: { version: number; declaration: boolean }) {
  return (
    await apiClient<ApiResponse<Project>>(`/admin/projects/${id}/submit`, {
      method: 'POST',
      ...json(input)
    })
  ).data;
}

/** Comprovante de recebimento (PDF). A resposta não é JSON, por isso não passa pelo apiClient. */
export async function downloadReceipt(id: string): Promise<Blob> {
  const res = await fetch(`/bff/admin/projects/${id}/receipt`);
  if (!res.ok) throw await apiErrorFromResponse(res);
  return res.blob();
}

export async function resubmitProject(
  id: string,
  input: { version: number; declaration: boolean }
) {
  return (
    await apiClient<ApiResponse<Project>>(`/admin/projects/${id}/resubmit`, {
      method: 'POST',
      ...json(input)
    })
  ).data;
}

export async function withdrawProject(id: string, note: string) {
  return (
    await apiClient<ApiResponse<Project>>(`/admin/projects/${id}/withdraw`, {
      method: 'POST',
      ...json({ note })
    })
  ).data;
}

export async function duplicateProject(id: string, copyDocuments: boolean) {
  return (
    await apiClient<ApiResponse<Project>>(`/admin/projects/${id}/duplicate`, {
      method: 'POST',
      ...json({ copyDocuments })
    })
  ).data;
}

export interface MemberInput {
  level: MemberLevel;
  label: string;
  /** Opcional: avisa quem ainda não tem conta (quem já tem é avisado no e-mail da conta). */
  email?: string;
}

export async function listMembers(id: string) {
  return (await apiClient<ApiResponse<Member[]>>(`/admin/projects/${id}/members`)).data;
}

export async function inviteMember(id: string, input: MemberInput & { cpf: string }) {
  return (
    await apiClient<ApiResponse<Member>>(`/admin/projects/${id}/members`, {
      method: 'POST',
      ...json(input)
    })
  ).data;
}

export async function updateMember(id: string, memberId: string, input: MemberInput) {
  return (
    await apiClient<ApiResponse<Member>>(`/admin/projects/${id}/members/${memberId}`, {
      method: 'PUT',
      ...json(input)
    })
  ).data;
}

export async function revokeMember(id: string, memberId: string) {
  await apiClient<void>(`/admin/projects/${id}/members/${memberId}`, { method: 'DELETE' });
}

const declaration = (id: string, code: string) => `/admin/projects/${id}/declarations/${code}`;

/** Prévia da declaração e situação da última emissão (a consulta também conclui a assinatura já feita no MT-ID). */
export async function getDeclaration(id: string, code: string) {
  return (await apiClient<ApiResponse<Declaration>>(declaration(id, code))).data;
}

/** Aceita o texto lido, gera o PDF e pede a assinatura ao MT-ID. */
export async function acceptDeclaration(id: string, code: string, paramsSha256: string) {
  return (
    await apiClient<ApiResponse<Declaration>>(`${declaration(id, code)}/accept`, {
      method: 'POST',
      ...json({ paramsSha256, confirm: true })
    })
  ).data;
}

export async function cancelDeclaration(id: string, code: string) {
  await apiClient<void>(declaration(id, code), { method: 'DELETE' });
}
