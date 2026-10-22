'use client';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import * as Sentry from '@sentry/nextjs';
import { useEffect } from 'react';

export default function OverviewError({ error, reset }: { error: Error; reset: () => void }) {
  useEffect(() => {
    Sentry.captureException(error);
  }, [error]);

  return (
    <Alert variant='destructive'>
      <AlertTitle>Não foi possível carregar a página</AlertTitle>
      <AlertDescription>
        <Button variant='outline' size='sm' onClick={reset}>
          Tentar novamente
        </Button>
      </AlertDescription>
    </Alert>
  );
}
