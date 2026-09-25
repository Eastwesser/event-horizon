import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/cn';

interface FilterChipProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  active?: boolean;
  children: ReactNode;
}

/** Shared filter/tab chip — readable as a button (Shop, History, Inventory, Leaderboard). */
export function FilterChip({ active = false, className, children, ...rest }: FilterChipProps) {
  return (
    <button
      type="button"
      className={cn(
        'rounded-full border px-4 py-2 text-sm font-medium transition-colors',
        active
          ? 'border-indigo bg-indigo/20 text-indigo-soft'
          : 'border-white/15 bg-transparent text-text-secondary hover:border-indigo/40 hover:bg-indigo/10 hover:text-text-primary',
        className,
      )}
      aria-pressed={active}
      {...rest}
    >
      {children}
    </button>
  );
}
