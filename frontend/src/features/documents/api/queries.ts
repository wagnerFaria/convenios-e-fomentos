import { queryOptions } from '@tanstack/react-query';
import { getDocument, listDocuments, listFiles } from './service';

export const documentKeys = {
  all: ['documents'] as const,
  list: (q: string, page: number) => [...documentKeys.all, 'list', q, page] as const,
  detail: (id: string) => [...documentKeys.all, 'detail', id] as const,
  files: (id: string) => [...documentKeys.all, 'files', id] as const
};

export const documentsQueryOptions = (q = '', page = 1) =>
  queryOptions({ queryKey: documentKeys.list(q, page), queryFn: () => listDocuments({ q, page }) });

export const documentQueryOptions = (id: string) =>
  queryOptions({ queryKey: documentKeys.detail(id), queryFn: () => getDocument(id) });

export const documentFilesQueryOptions = (id: string) =>
  queryOptions({ queryKey: documentKeys.files(id), queryFn: () => listFiles(id) });
