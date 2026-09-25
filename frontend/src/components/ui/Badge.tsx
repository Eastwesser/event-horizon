import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/cn';

type Tone = 'gold' | 'indigo' | 'cyan' | 'success' | 'error' | 'warning' | 'neutral';

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: Tone;
  children: ReactNode;
}

const toneClasses: Record<Tone, string> = {
  gold: 'bg-horizon-gold/15 text-horizon-gold border-horizon-gold/30',
  indigo: 'bg-indigo/15 text-indigo-soft border-indigo/30',
  cyan: 'bg-photon-cyan/15 text-photon-cyan border-photon-cyan/30',
  success: 'bg-success/15 text-success border-success/30',
  error: 'bg-error/15 text-error border-error/30',
  warning: 'bg-warning/15 text-warning border-warning/30',
  neutral: 'bg-white/10 text-text-secondary border-white/15',
};

export function Badge({ tone = 'neutral', className, children, ...rest }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-sm border px-2.5 py-1 font-body text-xs font-medium',
        toneClasses[tone],
        className,
      )}
      {...rest}
    >
      {children}
    </span>
  );
}
