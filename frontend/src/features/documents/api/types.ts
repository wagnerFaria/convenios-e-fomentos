import type { ApiMeta } from '@/lib/api/types';

/** Ações do backend usadas para esconder elementos de UI (a autorização real é do backend). */
export const documentActions = {
  read: 'admin::content.document.read',
  create: 'admin::content.document.create',
  update: 'admin::content.document.update',
  delete: 'admin::content.document.delete',
  publish: 'admin::content.document.publish',
  fileRead: 'admin::content.document.file.read',
  fileUpload: 'admin::content.document.file.upload',
  fileDelete: 'admin::content.document.file.delete'
} as const;

export type DocumentStatus = 'rascunho' | 'publicado';

/** Cada campo só vem se a permissão de leitura o liberar; `id` e os de sistema sempre vêm. */
export interface Document {
  id: string;
  titulo?: string;
  conteudo?: string;
  status?: DocumentStatus;
  observacao_interna?: string;
  department_id?: string | null;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface DocumentList {
  data: Document[];
  meta?: ApiMeta;
}

export interface DocumentInput {
  titulo: string;
  conteudo?: string;
}

export type FileStatus = 'pendente' | 'confirmado';

export interface DocumentFile {
  id: string;
  document_id: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  status: FileStatus;
  created_at: string;
  confirmed_at: string | null;
}

export interface FileInput {
  filename: string;
  content_type: string;
  size_bytes: number;
}

/** Envio autorizado: POST multipart para `url` com `fields` e, por último, o campo `file`. */
export interface UploadGrant {
  url: string;
  fields: Record<string, string>;
  expires_at: string;
}

export interface FileDownload {
  url: string;
  expires_at: string;
}
