import type { FundingType, InstitutionSection } from './api/types';

/**
 * Espelho da regra do servidor (a autoridade é o backend): o chamamento público é exigido no Termo de Fomento com
 * recurso próprio da SEDEC (puro ou misto) e no Convênio com entidade privada.
 */
export function tenderRequired(
  inst: Pick<InstitutionSection, 'instrument' | 'nature'>,
  type: FundingType | ''
) {
  if (inst.instrument === 'fomento') return type === 'own' || type === 'mixed';
  if (inst.instrument === 'convenio') return inst.nature === 'private';
  return false;
}

export const tenderStatusOptions = {
  fomento: [
    { value: 'held', label: 'Chamamento público realizado' },
    { value: 'waived', label: 'Chamamento dispensado' },
    { value: 'unfeasible', label: 'Chamamento inexigível' }
  ],
  convenio: [
    { value: 'held', label: 'Chamamento público realizado' },
    { value: 'waived', label: 'Chamamento dispensado (exceção do art. 14)' }
  ]
} as const;

/** Fundamentos legais aceitos; os códigos são os do servidor. */
export const tenderBasisOptions: Record<string, { value: string; label: string }[]> = {
  'fomento:waived': [
    {
      value: 'art19_i',
      label:
        'Art. 19, I: urgência por paralisação de atividade de relevante interesse público (até 180 dias)'
    },
    {
      value: 'art19_ii',
      label:
        'Art. 19, II: guerra, calamidade pública, grave perturbação da ordem ou ameaça à paz social'
    },
    { value: 'art19_iii', label: 'Art. 19, III: programa de proteção a pessoas ameaçadas' },
    {
      value: 'art19_iv',
      label: 'Art. 19, IV: educação, saúde ou assistência social, com OSC credenciada no SIGCon'
    }
  ],
  'fomento:unfeasible': [
    { value: 'art20_i', label: 'Art. 20, I: compromisso internacional que indica a instituição' },
    { value: 'art20_ii', label: 'Art. 20, II: lei que identifica a entidade beneficiária' },
    {
      value: 'art20_singular',
      label: 'Art. 20: objeto singular ou meta que só uma entidade pode atingir'
    }
  ],
  'convenio:waived': [
    {
      value: 'conv14_i',
      label: 'Art. 14, I (IN 001/2015): emergência ou calamidade pública, até 180 dias'
    },
    { value: 'conv14_ii', label: 'Art. 14, II (IN 001/2015): proteção a pessoas ameaçadas' },
    {
      value: 'conv14_iii',
      label:
        'Art. 14, III (IN 001/2015): mesma entidade executa o objeto há 5 anos, com contas aprovadas'
    }
  ]
};
