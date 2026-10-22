import Link from 'next/link';
import { InteractiveGridPattern } from './interactive-grid';
import { cn } from '@/lib/utils';

export const APP_NAME = 'Convênios e Fomentos';

export function AuthShell({
  title,
  description,
  children,
  footer
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
}) {
  return (
    <div className='relative flex min-h-screen flex-col items-center justify-center overflow-hidden md:grid lg:max-w-none lg:grid-cols-2 lg:px-0'>
      <div className='relative hidden h-full flex-col p-10 lg:flex dark:border-r'>
        <div className='bg-sidebar absolute inset-0' />
        <Link
          href='/'
          className='text-sidebar-foreground relative z-20 flex items-center text-lg font-medium'
        >
          {APP_NAME}
        </Link>
        <InteractiveGridPattern
          className={cn(
            'mask-[radial-gradient(400px_circle_at_center,white,transparent)]',
            'inset-x-0 inset-y-[0%] h-full skew-y-12'
          )}
        />
        <div className='text-sidebar-foreground relative z-20 mt-auto'>
          <p className='text-lg'>Gestão de convênios, termos de fomento e prestação de contas.</p>
        </div>
      </div>
      <div className='flex h-full w-full items-center justify-center p-4 lg:p-8'>
        <div className='flex w-full max-w-sm flex-col space-y-6'>
          <div className='flex flex-col space-y-1.5'>
            <h1 className='text-2xl font-semibold tracking-tight'>{title}</h1>
            {description && <p className='text-muted-foreground text-sm'>{description}</p>}
          </div>
          {children}
          {footer && <div className='text-muted-foreground text-center text-sm'>{footer}</div>}
        </div>
      </div>
    </div>
  );
}
