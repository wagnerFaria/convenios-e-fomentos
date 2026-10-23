import { queryOptions } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import type { ApiResponse } from '@/lib/api/types';

/** Município do IBGE (catálogo somente leitura do backend). */
export interface Municipality {
  ibgeCode: string;
  name: string;
  uf: string;
}

export const UFS = [
  'AC',
  'AL',
  'AP',
  'AM',
  'BA',
  'CE',
  'DF',
  'ES',
  'GO',
  'MA',
  'MT',
  'MS',
  'MG',
  'PA',
  'PB',
  'PR',
  'PE',
  'PI',
  'RJ',
  'RN',
  'RS',
  'RO',
  'RR',
  'SC',
  'SP',
  'SE',
  'TO'
] as const;

export const municipalityActions = { read: 'admin::catalog.municipalities.read' } as const;

const PAGE_SIZE = 500;

/** Todos os municípios de uma UF (até 853 em MG): busca por páginas de 500. */
async function listByUf(uf: string): Promise<Municipality[]> {
  const out: Municipality[] = [];
  for (let page = 1; ; page++) {
    const res = await apiClient<ApiResponse<Municipality[]>>(
      `/admin/catalog/municipalities?uf=${uf}&pageSize=${PAGE_SIZE}&page=${page}`
    );
    out.push(...res.data);
    if (!res.meta || page >= res.meta.pageCount) return out;
  }
}

export const municipalityKeys = {
  all: ['municipalities'] as const,
  byUf: (uf: string) => [...municipalityKeys.all, 'uf', uf] as const
};

/** A lista de uma UF não muda: fica em cache por bastante tempo. */
export const municipalitiesByUfQueryOptions = (uf: string) =>
  queryOptions({
    queryKey: municipalityKeys.byUf(uf),
    queryFn: () => listByUf(uf),
    enabled: uf !== '',
    staleTime: 60 * 60_000
  });
