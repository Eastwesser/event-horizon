import type { ReactNode } from 'react';
import { cn } from '../../lib/cn';

type Tone = 'indigo' | 'gold' | 'success' | 'danger' | 'cyan' | 'warning';
type Size = 'md' | 'sm';

interface StatCardProps {
  value: ReactNode;
  label: string;
  sub?: string;
  /** Value color — defaults to indigo-soft (brand). Use gold for reward/score highlights. */
  tone?: Tone;
  /** 'md' (default) is the dashboard size. 'sm' is a compact HUD pill. */
  size?: Size;
  className?: string;
}

const toneClasses: Record<Tone, string> = {
  indigo: 'text-indigo-soft',
  gold: 'text-horizon-gold',
  success: 'text-success',
  danger: 'text-error',
  cyan: 'text-photon-cyan',
  warning: 'text-warning',
};

export function StatCard({
  value,
  label,
  sub,
  tone = 'indigo',
  size = 'md',
  className,
}: StatCardProps) {
  const isSm = size === 'sm';

  return (
    <div
      className={cn(
        'rounded-md border border-white/10 bg-nebula text-center transition-transform duration-200 hover:-translate-y-1 hover:border-indigo/40',
        isSm ? 'min-w-[76px] px-3 py-1.5' : 'p-4',
        className,
      )}
    >
      <div
        className={cn(
          'font-hud font-semibold tabular-nums',
          isSm ? 'text-lg' : 'text-2xl',
          toneClasses[tone],
        )}
      >
        {value}
      </div>
      <div className={cn('text-text-secondary', isSm ? 'text-[0.65rem]' : 'mt-1 text-xs')}>{label}</div>
      {sub && <div className="mt-0.5 text-[0.65rem] text-text-muted">{sub}</div>}
    </div>
  );
}
