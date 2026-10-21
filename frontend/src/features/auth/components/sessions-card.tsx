'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ApiError } from '@/lib/api/errors';
import {
  revokeOtherSessions,
  revokeSession,
  sessionKeys,
  sessionsQueryOptions
} from '../api/sessions';

const dateTime = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' });

const onError = (err: unknown) =>
  toast.error(err instanceof ApiError ? err.message : 'Não foi possível concluir a operação.');

export function SessionsCard() {
  const qc = useQueryClient();
  const { data: sessions, isPending, error } = useQuery(sessionsQueryOptions());

  const refresh = () => qc.invalidateQueries({ queryKey: sessionKeys.all });

  const revoke = useMutation({ mutationFn: revokeSession, onSuccess: refresh, onError });
  const revokeOthers = useMutation({
    mutationFn: revokeOtherSessions,
    onSuccess: () => {
      toast.success('Os outros dispositivos foram desconectados.');
      return refresh();
    },
    onError
  });

  const others = sessions?.filter((s) => !s.current) ?? [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sessões ativas</CardTitle>
        <CardDescription>Dispositivos conectados à sua conta.</CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-3'>
        {isPending && <p className='text-muted-foreground text-sm'>Carregando…</p>}
        {error && <p className='text-destructive text-sm'>Não foi possível carregar as sessões.</p>}
        {sessions?.map((s) => (
          <div
            key={s.id}
            className='flex items-center justify-between gap-4 rounded-lg border p-3 text-sm'
          >
            <div className='min-w-0'>
              <p className='truncate font-medium'>
                {s.userAgent || 'Dispositivo desconhecido'}
                {s.current && (
                  <Badge variant='secondary' className='ml-2'>
                    Esta sessão
                  </Badge>
                )}
              </p>
              <p className='text-muted-foreground text-xs'>
                {s.ip} · último uso {dateTime.format(new Date(s.lastUsedAt))}
              </p>
            </div>
            {!s.current && (
              <Button
                variant='outline'
                size='sm'
                disabled={revoke.isPending}
                onClick={() => revoke.mutate(s.id)}
              >
                Desconectar
              </Button>
            )}
          </div>
        ))}
        {others.length > 0 && (
          <Button
            variant='outline'
            className='self-start'
            disabled={revokeOthers.isPending}
            onClick={() => revokeOthers.mutate()}
          >
            Desconectar todos os outros dispositivos
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
