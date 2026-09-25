import { cn } from '../../lib/cn';

interface SpinnerProps {
  size?: number;
  className?: string;
  label?: string;
}

/** Event-horizon loading motif: rotating indigo photon ring around a dark core. */
export function Spinner({ size = 40, className, label = 'Загрузка' }: SpinnerProps) {
  return (
    <div
      role="status"
      aria-label={label}
      className={cn('relative inline-block', className)}
      style={{ width: size, height: size }}
    >
      <div
        className="absolute inset-0 animate-spin rounded-full"
        style={{
          background:
            'conic-gradient(from 0deg, transparent 0%, var(--color-indigo) 20%, var(--color-indigo-soft) 30%, transparent 55%)',
          WebkitMask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), #000 calc(100% - 3px))',
          mask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), #000 calc(100% - 3px))',
        }}
      />
      <div className="absolute inset-[6px] animate-pulse rounded-full bg-indigo-soft/20" />
    </div>
  );
}
