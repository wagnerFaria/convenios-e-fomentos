'use client';

import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Icons } from '@/components/icons';
import { Badge } from '@/components/ui/badge';
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

const MAX = 50;

/** Municípios de execução do projeto: vários, todos de Mato Grosso. */
export function MunicipalitiesMultiSelect({
  id,
  value,
  onChange
}: {
  id?: string;
  value: string[];
  onChange: (codes: string[]) => void;
}) {
  const [open, setOpen] = useState(false);
  const list = useQuery(municipalitiesByUfQueryOptions('MT'));
  const names = new Map(list.data?.map((m) => [m.ibgeCode, m.name]));

  const toggle = (code: string) =>
    onChange(value.includes(code) ? value.filter((c) => c !== code) : [...value, code]);

  return (
    <div className='space-y-2'>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger
          disabled={list.isLoading}
          render={
            <Button
              id={id}
              type='button'
              variant='outline'
              role='combobox'
              aria-expanded={open}
              aria-controls={`${id ?? 'municipalities'}-list`}
              className='text-muted-foreground w-full justify-between font-normal'
            />
          }
        >
          {list.isLoading ? 'Carregando municípios…' : 'Adicionar município de execução'}
          <Icons.chevronsUpDown className='ml-2 h-4 w-4 shrink-0 opacity-50' />
        </PopoverTrigger>
        <PopoverContent className='w-(--anchor-width) p-0'>
          <Command>
            <CommandInput placeholder='Buscar município de Mato Grosso…' />
            <CommandList id={`${id ?? 'municipalities'}-list`}>
              <CommandEmpty>Nenhum município encontrado.</CommandEmpty>
              <CommandGroup>
                {list.data?.map((m) => (
                  <CommandItem
                    key={m.ibgeCode}
                    value={m.ibgeCode}
                    keywords={[m.name]}
                    disabled={!value.includes(m.ibgeCode) && value.length >= MAX}
                    onSelect={() => toggle(m.ibgeCode)}
                  >
                    <Icons.check
                      className={cn(
                        'mr-2 h-4 w-4',
                        value.includes(m.ibgeCode) ? 'opacity-100' : 'opacity-0'
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
      {value.length > 0 && (
        <ul className='flex flex-wrap gap-1.5' aria-label='Municípios de execução escolhidos'>
          {value.map((code) => (
            <li key={code}>
              <Badge variant='secondary' className='gap-1 pr-1'>
                {names.get(code) ?? code}
                <button
                  type='button'
                  onClick={() => toggle(code)}
                  aria-label={`Remover ${names.get(code) ?? code}`}
                  className='hover:bg-muted-foreground/20 rounded-sm p-0.5'
                >
                  <Icons.close className='h-3 w-3' aria-hidden />
                </button>
              </Badge>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
