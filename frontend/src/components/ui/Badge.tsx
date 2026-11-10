import { cn } from '@/utils/helpers';

interface BadgeProps {
  children: React.ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info';
  className?: string;
}

export function Badge({ children, variant = 'default', className }: BadgeProps) {
  const variantClasses = {
    default: 'bg-dark-100 text-dark-700 dark:bg-dark-800 dark:text-dark-200',
    success: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    error: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
    info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  };

  return (
    <span className={cn('inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', variantClasses[variant], className)}>
      {children}
    </span>
  );
}

export function StatusBadge({ status }: { status: string }) {
  const statusConfig: Record<string, { label: string; variant: BadgeProps['variant'] }> = {
    connected: { label: 'Connected', variant: 'success' },
    disconnected: { label: 'Disconnected', variant: 'default' },
    reconnecting: { label: 'Reconnecting', variant: 'warning' },
    error: { label: 'Error', variant: 'error' },
    running: { label: 'Running', variant: 'success' },
    stopped: { label: 'Stopped', variant: 'default' },
    starting: { label: 'Starting', variant: 'info' },
    stopping: { label: 'Stopping', variant: 'warning' },
    active: { label: 'Active', variant: 'success' },
    inactive: { label: 'Inactive', variant: 'default' },
    VERIFIED: { label: 'Verified', variant: 'success' },
    PARTIALLY_VERIFIED: { label: 'Partially Verified', variant: 'warning' },
    UNVERIFIED: { label: 'Unverified', variant: 'default' },
    UNSUPPORTED: { label: 'Unsupported', variant: 'error' },
    DISABLED: { label: 'Disabled', variant: 'default' },
  };

  const config = statusConfig[status] || { label: status, variant: 'default' };

  return <Badge variant={config.variant}>{config.label}</Badge>;
}