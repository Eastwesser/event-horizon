import type { ReactNode } from 'react';
import { cn } from '../../lib/cn';

type Width = 'narrow' | 'wide';
type Pad = 'default' | 'game';

interface PageShellProps {
  children: ReactNode;
  /**
   * narrow = max-w-3xl (Profile, Leaderboard, forms, Subscription)
   * wide   = max-w-6xl (Shop, Inventory, Analytics, Home content)
   */
  width?: Width;
  /** default = standard page; game = HUD chrome vertical rhythm */
  pad?: Pad;
  className?: string;
  innerClassName?: string;
}

const widthClasses: Record<Width, string> = {
  narrow: 'max-w-3xl',
  wide: 'max-w-6xl',
};

/**
 * ROOT CAUSE of "glued to left edge" (from screenshot review):
 * 1. Home never used PageShell — footer/header lived outside any max-width.
 * 2. PageShell used only px-4 (16px) — on large monitors that reads as "stuck".
 * 3. Card grids used auto-fill left-aligned → content hugs the left of a wide shell.
 *
 * Fix: stronger horizontal inset (px-6 / sm:px-8), always mx-auto + w-full,
 * and pages must put ALL chrome (header/footer) inside the same shell edges.
 */
const padClasses: Record<Pad, string> = {
  default: 'box-border px-6 py-6 sm:px-8 sm:py-8',
  game: 'box-border px-4 pb-3 pt-3 sm:px-6 sm:pt-4',
};

export function PageShell({
  children,
  width = 'wide',
  pad = 'default',
  className,
  innerClassName,
}: PageShellProps) {
  return (
    <div className={cn('box-border min-h-screen w-full bg-void text-text-primary', padClasses[pad], className)}>
      <div className={cn('mx-auto w-full', widthClasses[width], innerClassName)}>{children}</div>
    </div>
  );
}
