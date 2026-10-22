/** CPF: máscara e validação, espelhando backend/internal/cpf. */

/** Remove tudo que não é dígito. */
export function onlyDigits(value: string): string {
  return value.replace(/\D/g, '');
}

/** Formata enquanto digita: 000.000.000-00 (aceita CPF parcial). */
export function formatCpf(value: string): string {
  const d = onlyDigits(value).slice(0, 11);
  return d
    .replace(/^(\d{3})(\d)/, '$1.$2')
    .replace(/^(\d{3})\.(\d{3})(\d)/, '$1.$2.$3')
    .replace(/^(\d{3})\.(\d{3})\.(\d{3})(\d)/, '$1.$2.$3-$4');
}

function checkDigit(base: string): number {
  let sum = 0;
  for (let i = 0; i < base.length; i++) sum += Number(base[i]) * (base.length + 1 - i);
  const mod = (sum * 10) % 11;
  return mod === 10 ? 0 : mod;
}

/** Valida os dígitos verificadores (rejeita sequências repetidas como 111.111.111-11). */
export function isValidCpf(value: string): boolean {
  const d = onlyDigits(value);
  if (d.length !== 11 || /^(\d)\1{10}$/.test(d)) return false;
  const first = checkDigit(d.slice(0, 9));
  return first === Number(d[9]) && checkDigit(d.slice(0, 10)) === Number(d[10]);
}
