import type { ReactNode } from 'react';
import { cn } from '../../lib/cn';

interface PageHeaderProps {
  title: string;
  /** Optional line of muted copy shown under the title, inside the same header block. */
  subtitle?: ReactNode;
  /** Renders a round back button that navigates here (or calls the given handler). */
  onBack?: () => void;
  /** Accessible label for the back button (visible title stays the page name). */
  backLabel?: string;
  actions?: ReactNode;
  className?: string;
}

export function PageHeader({
  title,
  subtitle,
  onBack,
  backLabel = 'Назад',
  actions,
  className,
}: PageHeaderProps) {
  return (
    <div className={cn('mb-8 border-b border-white/10 pb-4', className)}>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {onBack && (
            <button
              type="button"
              aria-label={backLabel}
              onClick={onBack}
              className="flex h-9 w-9 items-center justify-center rounded-full border border-indigo/40 text-indigo-soft transition-colors hover:bg-indigo/10"
            >
              ←
            </button>
          )}
          <h1 className="font-display text-2xl font-semibold leading-tight text-text-primary sm:text-3xl">{title}</h1>
        </div>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
      {subtitle && <p className="mt-2 text-sm leading-normal text-text-secondary">{subtitle}</p>}
    </div>
  );
}
