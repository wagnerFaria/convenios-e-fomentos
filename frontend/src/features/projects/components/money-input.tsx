'use client';

import { Field, FieldDescription, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { formatBrl } from '../format';

const MAX_CENTS = 1_000_000_000_000; // R$ 10 bilhões, o mesmo limite do servidor

/** Valor em reais digitado como numa máscara bancária: só dígitos, sempre com centavos. O estado é em centavos. */
export function MoneyInput({
  id,
  value,
  onChange,
  disabled,
  'aria-label': ariaLabel
}: {
  id?: string;
  value: number;
  onChange: (cents: number) => void;
  disabled?: boolean;
  'aria-label'?: string;
}) {
  return (
    <Input
      id={id}
      aria-label={ariaLabel}
      inputMode='numeric'
      autoComplete='off'
      placeholder='R$ 0,00'
      disabled={disabled}
      value={value ? formatBrl(value) : ''}
      onChange={(e) => {
        const digits = e.target.value.replace(/\D/g, '').slice(0, 13);
        onChange(digits ? Math.min(Number(digits), MAX_CENTS) : 0);
      }}
    />
  );
}

export function MoneyField({
  id,
  label,
  description,
  required,
  value,
  onChange
}: {
  id: string;
  label: string;
  description?: string;
  required?: boolean;
  value: number;
  onChange: (cents: number) => void;
}) {
  return (
    <Field>
      <FieldLabel htmlFor={id}>
        {label}
        {required && ' *'}
      </FieldLabel>
      <MoneyInput id={id} value={value} onChange={onChange} />
      {description && <FieldDescription>{description}</FieldDescription>}
    </Field>
  );
}
