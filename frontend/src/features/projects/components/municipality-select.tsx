'use client';

import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList
} from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import { municipalitiesByUfQueryOptions } from '@/features/municipalities/api/queries';

/** Escolha de um município da UF informada (lista do IBGE, com busca). */
export function MunicipalitySelect({
  id,
  uf,
  value,
  onChange,
  invalid
}: {
  id?: string;
  uf: string;
  value: string;
  onChange: (ibgeCode: string) => void;
  invalid?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const list = useQuery(municipalitiesByUfQueryOptions(uf));
  const selected = list.data?.find((m) => m.ibgeCode === value);
  const placeholder = !uf
    ? 'Escolha a UF primeiro'
    : list.isLoading
      ? 'Carregando municípios…'
      : 'Escolha o município';

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        disabled={!uf || list.isLoading}
        render={
          <Button
            id={id}
            type='button'
            variant='outline'
            role='combobox'
            aria-expanded={open}
            aria-controls={`${id ?? 'municipality'}-list`}
            aria-invalid={invalid}
            className={cn(
              'w-full justify-between font-normal',
              !selected && 'text-muted-foreground'
            )}
          />
        }
      >
        {selected?.name ?? (value && uf ? 'Município selecionado' : placeholder)}
        <Icons.chevronsUpDown className='ml-2 h-4 w-4 shrink-0 opacity-50' />
      </PopoverTrigger>
      <PopoverContent className='w-(--anchor-width) p-0'>
        <Command>
          <CommandInput placeholder='Buscar município…' />
          <CommandList id={`${id ?? 'municipality'}-list`}>
            <CommandEmpty>Nenhum município encontrado.</CommandEmpty>
            <CommandGroup>
              {list.data?.map((m) => (
                <CommandItem
                  key={m.ibgeCode}
                  value={m.ibgeCode}
                  keywords={[m.name]}
                  onSelect={() => {
                    onChange(m.ibgeCode);
                    setOpen(false);
                  }}
                >
                  <Icons.check
                    className={cn(
                      'mr-2 h-4 w-4',
                      value === m.ibgeCode ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  {m.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
