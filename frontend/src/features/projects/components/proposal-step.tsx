'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useRef } from 'react';
import { toast } from 'sonner';
import { useStore } from '@tanstack/react-form';
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field';
import { useAppForm } from '@/lib/form';
import { projectKeys } from '../api/queries';
import { saveProposal } from '../api/service';
import { objectNatureLabels, type ObjectNature, type Project } from '../api/types';
import { MunicipalitiesMultiSelect } from './municipalities-multi-select';
import { ServerErrors } from './server-errors';
import { StepActions } from './step-actions';

const natureOptions = Object.entries(objectNatureLabels).map(([value, label]) => ({
  value,
  label
}));

export function ProposalStep({
  project,
  readOnly,
  onSaved,
  onDirty
}: {
  project: Project;
  readOnly: boolean;
  onSaved: (project: Project, advance: boolean) => void;
  onDirty: (dirty: boolean) => void;
}) {
  const qc = useQueryClient();
  const advance = useRef(false);
  const p = project.proposal;

  const save = useMutation({
    mutationFn: (v: ReturnType<typeof defaults>) =>
      saveProposal(project.id, {
        version: project.version,
        ...v,
        objectNature: v.objectNature as ObjectNature | ''
      }),
    onSuccess: (saved) => {
      qc.setQueryData(projectKeys.detail(project.id), saved);
      void qc.invalidateQueries({ queryKey: projectKeys.checklist(project.id) });
      void qc.invalidateQueries({ queryKey: projectKeys.lists() });
      toast.success('Proposta salva.');
      onSaved(saved, advance.current);
    }
  });

  const defaults = () => ({
    sigconNumber: p.sigconNumber,
    sigconRegisteredOn: p.sigconRegisteredOn,
    title: p.title,
    object: p.object,
    justification: p.justification,
    startDate: p.startDate,
    endDate: p.endDate,
    objectNature: p.objectNature as string,
    hasArtisticHiring: p.hasArtisticHiring,
    calamity: p.calamity,
    executionPlaceDetail: p.executionPlaceDetail,
    municipalityCodes: p.municipalities.map((m) => m.ibgeCode)
  });

  const form = useAppForm({
    defaultValues: defaults(),
    onSubmit: ({ value }) => save.mutateAsync(value).catch(() => undefined)
  });
  const dirty = useStore(form.store, (s) => s.isDirty);
  useEffect(() => onDirty(dirty), [dirty, onDirty]);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void form.handleSubmit();
      }}
    >
      <FieldGroup>
        <div className='grid gap-4 sm:grid-cols-2'>
          <form.AppField
            name='sigconNumber'
            children={(f) => (
              <f.TextField
                label='Nº da proposta no SIGCon'
                required
                placeholder='3216/2026'
                description='Número da proposta/termo, no formato NNNN/AAAA.'
              />
            )}
          />
          <form.AppField
            name='sigconRegisteredOn'
            children={(f) => (
              <f.TextField label='Data do cadastro no SIGCon' type='date' required />
            )}
          />
        </div>
        <form.AppField
          name='title'
          children={(f) => <f.TextField label='Título do projeto' required maxLength={200} />}
        />
        <form.AppField
          name='object'
          children={(f) => (
            <f.TextareaField
              label='Objeto'
              required
              rows={4}
              maxLength={5000}
              showCount
              description='O que será feito, em linguagem clara.'
            />
          )}
        />
        <form.AppField
          name='justification'
          children={(f) => (
            <f.TextareaField label='Justificativa' required rows={4} maxLength={10000} showCount />
          )}
        />
        <div className='grid gap-4 sm:grid-cols-2'>
          <form.AppField
            name='startDate'
            children={(f) => <f.TextField label='Início previsto' type='date' required />}
          />
          <form.AppField
            name='endDate'
            children={(f) => <f.TextField label='Término previsto' type='date' required />}
          />
        </div>
        <form.AppField
          name='objectNature'
          children={(f) => (
            <f.SelectField
              label='Natureza do objeto'
              required
              placeholder='Escolha'
              options={natureOptions}
              description='Define quais documentos de engenharia, licenças e regras específicas se aplicam.'
            />
          )}
        />
        <form.Field
          name='municipalityCodes'
          children={(f) => (
            <Field>
              <FieldLabel htmlFor='proposal-municipalities'>Municípios de execução *</FieldLabel>
              <MunicipalitiesMultiSelect
                id='proposal-municipalities'
                value={f.state.value}
                onChange={f.handleChange}
              />
              <FieldDescription>
                Onde o projeto será executado. Apenas municípios de Mato Grosso.
              </FieldDescription>
            </Field>
          )}
        />
        <form.AppField
          name='executionPlaceDetail'
          children={(f) => (
            <f.TextField
              label='Detalhe do local'
              maxLength={500}
              description='Endereço, trecho ou coordenadas, se ajudar.'
            />
          )}
        />
        <form.AppField
          name='hasArtisticHiring'
          children={(f) => (
            <f.SwitchField
              label='Haverá contratação de artistas (shows e eventos musicais)?'
              description='Aplica os limites da Lei Estadual 12.082/2023.'
            />
          )}
        />
        <form.AppField
          name='calamity'
          children={(f) => (
            <f.SwitchField
              label='Calamidade pública ou situação de emergência?'
              description='Dispensa a contrapartida, mediante o ato oficial.'
            />
          )}
        />
        <ServerErrors error={save.error} />
        <StepActions
          saving={save.isPending}
          disabled={readOnly}
          onSave={() => (advance.current = false)}
          onSaveAndContinue={() => (advance.current = true)}
          last
        />
      </FieldGroup>
    </form>
  );
}
