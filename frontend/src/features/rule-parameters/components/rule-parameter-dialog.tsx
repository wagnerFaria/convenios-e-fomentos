'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { FieldGroup } from '@/components/ui/field';
import { useCan } from '@/features/auth/session-provider';
import { errorMessage } from '@/lib/errors';
import { useAppForm } from '@/lib/form';
import { ruleParameterKeys } from '../api/queries';
import { createRuleParameter, updateRuleParameter } from '../api/service';
import {
  ruleParameterActions,
  valueExamples,
  type RuleParameter,
  type RuleParameterInput,
  type RuleParameterKey
} from '../api/types';

const isoDate = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Informe a data.');

const schema = z.object({
  value: z
    .string()
    .trim()
    .min(1, 'Informe o valor.')
    .refine(
      (s) => {
        try {
          JSON.parse(s);
          return true;
        } catch {
          return false;
        }
      },
      { message: 'O valor precisa ser um JSON válido.' }
    ),
  validFrom: isoDate,
  validTo: z.string().refine((s) => s === '' || /^\d{4}-\d{2}-\d{2}$/.test(s), 'Data inválida.'),
  note: z.string().max(500, 'Máximo de 500 caracteres.')
});

const pretty = (v: unknown) => JSON.stringify(v, null, 2);

function RuleParameterForm({
  paramKey,
  parameter,
  onDone
}: {
  paramKey: RuleParameterKey;
  parameter?: RuleParameter;
  onDone: () => void;
}) {
  const qc = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const save = useMutation({
    mutationFn: (input: RuleParameterInput) =>
      parameter ? updateRuleParameter(parameter.id, input) : createRuleParameter(input),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ruleParameterKeys.all });
      toast.success(parameter ? 'Versão atualizada.' : 'Versão criada.');
      onDone();
    },
    onError: (err) => setError(errorMessage(err))
  });
  const form = useAppForm({
    defaultValues: {
      value: pretty(parameter ? parameter.value : (valueExamples[paramKey.key] ?? '')),
      validFrom: parameter?.validFrom ?? '',
      validTo: parameter?.validTo ?? '',
      note: parameter?.note ?? ''
    },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      setError(null);
      if (value.validTo && value.validTo < value.validFrom) {
        setError('O fim da vigência não pode ser anterior ao início.');
        return;
      }
      return save.mutateAsync({
        key: paramKey.key,
        value: JSON.parse(value.value) as unknown,
        validFrom: value.validFrom,
        validTo: value.validTo || null,
        note: value.note.trim()
      });
    }
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.AppField
          name='value'
          children={(f) => (
            <f.TextareaField
              label='Valor'
              required
              rows={8}
              className='font-mono text-xs'
              description={`Formato JSON. ${paramKey.description}`}
            />
          )}
        />
        <div className='grid gap-4 sm:grid-cols-2'>
          <form.AppField
            name='validFrom'
            children={(f) => <f.TextField type='date' label='Vigente desde' required />}
          />
          <form.AppField
            name='validTo'
            children={(f) => (
              <f.TextField type='date' label='Vigente até' description='Vazio = sem fim.' />
            )}
          />
        </div>
        <form.AppField
          name='note'
          children={(f) => (
            <f.TextField label='Observação' description='Norma ou motivo da mudança (opcional).' />
          )}
        />
        {error && (
          <p role='alert' className='text-destructive text-sm'>
            {error}
          </p>
        )}
        <form.AppForm>
          <form.SubmitButton>{parameter ? 'Salvar' : 'Criar versão'}</form.SubmitButton>
        </form.AppForm>
      </FieldGroup>
    </form>
  );
}

/** Sem `parameter`, é o botão "Nova versão" da chave; com ele, o diálogo (controlado) de edição. */
export function RuleParameterDialog(
  props:
    | {
        paramKey: RuleParameterKey;
        parameter: RuleParameter;
        open: boolean;
        onOpenChange: (open: boolean) => void;
      }
    | {
        paramKey: RuleParameterKey;
        parameter?: undefined;
        open?: undefined;
        onOpenChange?: undefined;
      }
) {
  const can = useCan();
  const [open, setOpen] = useState(false);

  if (props.parameter) {
    return (
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent className='sm:max-w-xl'>
          <DialogHeader>
            <DialogTitle>Editar versão</DialogTitle>
            <DialogDescription className='font-mono'>{props.paramKey.key}</DialogDescription>
          </DialogHeader>
          <RuleParameterForm
            paramKey={props.paramKey}
            parameter={props.parameter}
            onDone={() => props.onOpenChange(false)}
          />
        </DialogContent>
      </Dialog>
    );
  }
  if (!can(ruleParameterActions.create)) return null;
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button size='sm' variant='outline' />}>
        <Icons.add aria-hidden /> Nova versão
      </DialogTrigger>
      <DialogContent className='sm:max-w-xl'>
        <DialogHeader>
          <DialogTitle>Nova versão</DialogTitle>
          <DialogDescription>
            Para trocar um valor a partir de uma data, encerre a versão atual (vigente até) e crie a
            nova. Versões da mesma chave não podem se sobrepor.
          </DialogDescription>
        </DialogHeader>
        <RuleParameterForm paramKey={props.paramKey} onDone={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}
