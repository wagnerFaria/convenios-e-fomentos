import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';
import type {
  Document,
  DocumentFile,
  DocumentInput,
  DocumentList,
  FileDownload,
  FileInput,
  UploadGrant
} from './types';

const base = '/admin/content/documents';
const json = (body: unknown): RequestInit => ({ body: JSON.stringify(body) });

export async function listDocuments(params: { q?: string; page?: number } = {}) {
  const qs = new URLSearchParams({ pageSize: '25', sort: 'created_at:desc' });
  if (params.q) qs.set('q', params.q);
  if (params.page) qs.set('page', String(params.page));
  return apiClient<DocumentList>(`${base}?${qs}`);
}

export async function getDocument(id: string) {
  return (await apiClient<ApiResponse<Document>>(`${base}/${id}`)).data;
}

export async function createDocument(input: DocumentInput) {
  return (await apiClient<ApiResponse<Document>>(base, { method: 'POST', ...json(input) })).data;
}

export async function listFiles(documentId: string) {
  return (await apiClient<ApiResponse<DocumentFile[]>>(`${base}/${documentId}/files`)).data;
}

/** Passo 1 do envio: registra o arquivo e obtém a autorização para enviá-lo ao armazenamento. */
export async function requestUpload(documentId: string, input: FileInput) {
  return (
    await apiClient<ApiResponse<{ file: DocumentFile; upload: UploadGrant }>>(
      `${base}/${documentId}/files`,
      { method: 'POST', ...json(input) }
    )
  ).data;
}

/** Passo 3: a API confere o objeto no armazenamento e confirma o arquivo. */
export async function confirmUpload(documentId: string, fileId: string) {
  return (
    await apiClient<ApiResponse<DocumentFile>>(`${base}/${documentId}/files/${fileId}/confirm`, {
      method: 'POST'
    })
  ).data;
}

export async function getDownload(documentId: string, fileId: string) {
  return (
    await apiClient<ApiResponse<FileDownload>>(`${base}/${documentId}/files/${fileId}/download`)
  ).data;
}

export async function deleteFile(documentId: string, fileId: string) {
  await apiClient<void>(`${base}/${documentId}/files/${fileId}`, { method: 'DELETE' });
}
