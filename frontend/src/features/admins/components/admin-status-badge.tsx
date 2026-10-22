import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { adminStatusLabels, type AdminStatus } from '../api/types';

const variants: Record<AdminStatus, 'default' | 'secondary' | 'outline' | 'destructive'> = {
  active: 'default',
  awaiting: 'outline',
  pending: 'secondary',
  inactive: 'secondary',
  blocked: 'destructive'
};

export function AdminStatusBadge({ status }: { status: AdminStatus }) {
  return (
    <Badge
      variant={variants[status]}
      className={cn(
        status === 'awaiting' && 'border-amber-500/60 text-amber-700 dark:text-amber-400'
      )}
    >
      {adminStatusLabels[status]}
    </Badge>
  );
}
