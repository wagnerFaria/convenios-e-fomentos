import { Alert, AlertDescription } from '@/components/ui/alert';
import { ApiError } from '@/lib/api/errors';
import { errorMessage } from '@/lib/errors';

/** Erros devolvidos pelo backend ao gravar uma seção: lista todos os campos recusados, não só o primeiro. */
export function ServerErrors({ error }: { error: unknown }) {
  if (!error) return null;
  const fields = error instanceof ApiError ? error.fieldErrors : [];
  return (
    <Alert variant='destructive' role='alert'>
      <AlertDescription>
        {fields.length > 1 ? (
          <>
            <p>Corrija os campos abaixo e salve de novo:</p>
            <ul className='list-disc pl-5'>
              {fields.map((f, i) => (
                <li key={i}>{f.message}</li>
              ))}
            </ul>
          </>
        ) : (
          errorMessage(error)
        )}
      </AlertDescription>
    </Alert>
  );
}
