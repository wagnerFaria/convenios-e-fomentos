export interface FieldError {
  field: string;
  message: string;
}

/** Erro devolvido pelo backend: `{ error: { status, code, message, details } }`. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly details?: Record<string, unknown>;
  readonly requestId?: string;

  constructor(init: {
    status: number;
    code: string;
    message: string;
    details?: Record<string, unknown>;
    requestId?: string;
  }) {
    super(init.message);
    this.name = 'ApiError';
    this.status = init.status;
    this.code = init.code;
    this.details = init.details;
    this.requestId = init.requestId;
  }

  /** Erros de validação (422) por campo. */
  get fieldErrors(): FieldError[] {
    const fields = this.details?.fields;
    return Array.isArray(fields) ? (fields as FieldError[]) : [];
  }
}

/** Constrói o ApiError a partir de uma resposta de erro, sem nunca lançar. */
export async function apiErrorFromResponse(res: Response): Promise<ApiError> {
  const requestId = res.headers.get('X-Request-ID') ?? undefined;
  try {
    const body = await res.json();
    const e = body?.error;
    if (e && typeof e.code === 'string') {
      return new ApiError({
        status: e.status ?? res.status,
        code: e.code,
        message: e.message ?? 'Erro inesperado.',
        details: e.details,
        requestId
      });
    }
  } catch {
    // corpo ausente ou que não é JSON: cai no erro genérico abaixo
  }
  return new ApiError({
    status: res.status,
    code: res.status >= 500 ? 'internal_error' : 'unknown_error',
    message: res.status >= 500 ? 'Erro interno. Tente novamente.' : 'Erro inesperado.',
    requestId
  });
}
