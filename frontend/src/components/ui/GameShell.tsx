import type { ReactNode } from 'react';
import { PageHeader } from './PageHeader';
import { PageShell } from './PageShell';
import { cn } from '../../lib/cn';

interface GameShellProps {
  title: string;
  onBack: () => void;
  actions?: ReactNode;
  stats?: ReactNode;
  controls?: ReactNode;
  children: ReactNode;
  help?: ReactNode;
  width?: 'narrow' | 'wide';
  className?: string;
}

/**
 * Fit-to-viewport: outer shell is exactly 100dvh with no page scroll.
 * Stage flexes into leftover space after header/stats/controls/help.
 *
 * Control order convention (all games):
 *   primary (Старт / Новая) → secondary (Завершить / Сохранить) → extras (LB / skin) → Boost last
 */
export function GameShell({
  title,
  onBack,
  actions,
  stats,
  controls,
  children,
  help,
  width = 'wide',
  className,
}: GameShellProps) {
  return (
    <PageShell
      width={width}
      pad="game"
      className={cn('flex h-dvh max-h-dvh flex-col overflow-hidden', className)}
      innerClassName="flex h-full min-h-0 flex-col"
    >
      <PageHeader
        title={title}
        onBack={onBack}
        backLabel="На главную"
        actions={actions}
        className="mb-1.5 shrink-0 !pb-1.5 sm:mb-2"
      />

      {stats && (
        <div className="mb-1.5 flex shrink-0 flex-wrap items-center justify-center gap-1.5 sm:mb-2 sm:gap-2 sm:justify-start">
          {stats}
        </div>
      )}

      {controls && (
        <div className="relative z-20 mb-1.5 flex shrink-0 flex-wrap items-center justify-center gap-2 sm:mb-2 sm:gap-3 sm:justify-start [&_button]:min-h-10 [&_button]:min-w-10">
          {controls}
        </div>
      )}

      <div className="flex min-h-0 flex-1 flex-col items-center justify-center overflow-hidden">
        <div className="flex max-h-full min-h-0 w-full max-w-full flex-col items-center justify-center overflow-hidden">
          {children}
        </div>
      </div>

      {help && (
        <details className="mt-1.5 shrink-0 rounded-md border border-white/10 bg-nebula p-2.5 text-sm text-text-secondary">
          <summary className="cursor-pointer font-semibold text-indigo-soft">Как играть?</summary>
          <div className="mt-2 max-h-24 space-y-1 overflow-y-auto">{help}</div>
        </details>
      )}
    </PageShell>
  );
}

export function ScoreChip({
  label,
  value,
  className,
}: {
  label: string;
  value: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-nebula px-2.5 py-1 font-hud text-xs tabular-nums sm:gap-2 sm:px-4 sm:py-2 sm:text-sm',
        className,
      )}
    >
      <span className="text-text-secondary">{label}:</span>
      <span className="font-semibold text-indigo-soft">{value}</span>
    </div>
  );
}
