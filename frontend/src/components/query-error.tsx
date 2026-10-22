import { Alert, AlertDescription } from '@/components/ui/alert';
import { ApiError } from '@/lib/api/errors';

/** Mensagem de erro de uma consulta, distinguindo falta de permissão. */
export function QueryError({ error, what }: { error: unknown; what: string }) {
  const forbidden = error instanceof ApiError && error.status === 403;
  return (
    <Alert variant='destructive'>
      <AlertDescription>
        {forbidden
          ? `Você não tem permissão para ver ${what}.`
          : `Não foi possível carregar ${what}.`}
      </AlertDescription>
    </Alert>
  );
}
