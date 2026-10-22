import { ApiError } from '@/lib/api/errors';

/** Mensagem amigável de qualquer erro de mutação; prefere o erro do primeiro campo inválido. */
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    const field = err.fieldErrors[0];
    return field ? field.message : err.message;
  }
  return err instanceof Error ? err.message : 'Erro inesperado.';
}
