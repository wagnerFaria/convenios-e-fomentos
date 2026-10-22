import { Avatar, AvatarFallback } from '@/components/ui/avatar';

interface UserAvatarProfileProps {
  className?: string;
  showInfo?: boolean;
  user: {
    fullName: string;
    email: string;
  } | null;
}

export function UserAvatarProfile({ className, showInfo = false, user }: UserAvatarProfileProps) {
  return (
    <div className='flex items-center gap-2'>
      <Avatar className={className}>
        <AvatarFallback className='rounded-lg'>
          {user?.fullName.slice(0, 2).toUpperCase() || 'CF'}
        </AvatarFallback>
      </Avatar>

      {showInfo && (
        <div className='grid flex-1 text-left text-sm leading-tight'>
          <span className='truncate font-semibold'>{user?.fullName ?? ''}</span>
          <span className='truncate text-xs'>{user?.email ?? ''}</span>
        </div>
      )}
    </div>
  );
}
