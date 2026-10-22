'use client';

import { useState } from 'react';
import { toast } from 'sonner';
import { Icons } from '@/components/icons';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import type { CreatedToken } from '../api/types';

/** Mostra o valor do token uma única vez; depois que fechar, não há como recuperá-lo. */
export function TokenCreatedDialog({
  created,
  onClose
}: {
  created: CreatedToken | null;
  onClose: () => void;
}) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    if (!created) return;
    try {
      await navigator.clipboard.writeText(created.token);
      setCopied(true);
    } catch {
      toast.error('Não foi possível copiar. Selecione o texto e copie manualmente.');
    }
  };

  return (
    <Dialog
      open={created !== null}
      onOpenChange={(open) => {
        if (!open) {
          setCopied(false);
          onClose();
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Token &quot;{created?.name}&quot; gerado</DialogTitle>
          <DialogDescription>
            Use-o no cabeçalho <code>Authorization: Bearer …</code> da API pública.
          </DialogDescription>
        </DialogHeader>
        <Alert>
          <Icons.warning aria-hidden />
          <AlertDescription>
            Copie agora: por segurança, este valor não será exibido novamente. Se perdê-lo, será
            preciso regenerar o token.
          </AlertDescription>
        </Alert>
        <div className='flex items-center gap-2'>
          <code
            aria-label='Valor do token'
            className='bg-muted min-w-0 flex-1 rounded-md p-2 text-xs break-all select-all'
          >
            {created?.token}
          </code>
          <Button variant='outline' onClick={() => void copy()}>
            {copied ? <Icons.check aria-hidden /> : <Icons.copy aria-hidden />}
            {copied ? 'Copiado' : 'Copiar'}
          </Button>
        </div>
        <Button onClick={onClose}>Concluir</Button>
      </DialogContent>
    </Dialog>
  );
}
