import { Icons } from '@/components/icons';

/** Aviso padrão de página sem permissão (a autorização real é sempre do backend). */
export function AccessDenied({
  message = 'Você não tem permissão para acessar esta página.'
}: {
  message?: string;
}) {
  return (
    <div className='text-muted-foreground flex flex-col items-center gap-2 text-center'>
      <Icons.shieldLock className='size-10' aria-hidden />
      <p className='text-lg'>{message}</p>
    </div>
  );
}
