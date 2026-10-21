import type { AdminMe } from '@/lib/api/types';

export function fullName(admin: Pick<AdminMe, 'firstName' | 'lastName'>): string {
  return `${admin.firstName} ${admin.lastName}`.trim();
}
