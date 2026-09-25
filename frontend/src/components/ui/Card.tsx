import type { ElementType, HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/cn';

interface CardProps extends HTMLAttributes<HTMLElement> {
  children: ReactNode;
  /** Adds a gentle lift used for clickable cards (game tiles, shop items). */
  interactive?: boolean;
  /** Gold glow reserved for flagship / accent selection — use sparingly. */
  glow?: boolean;
  /** Renders as a different element (e.g. 'li' inside a list) while keeping card styling. */
  as?: ElementType;
}

export function Card({
  children,
  interactive = false,
  glow = false,
  as: Component = 'div',
  className,
  ...rest
}: CardProps) {
  return (
    <Component
      className={cn(
        'rounded-md border border-white/10 bg-nebula p-6',
        'transition-[transform,box-shadow,border-color] duration-300 ease-warp',
        interactive &&
          'cursor-pointer hover:-translate-y-1 hover:border-indigo/50 hover:shadow-glow-indigo',
        glow && 'border-horizon-gold/50 shadow-glow-gold',
        className,
      )}
      {...rest}
    >
      {children}
    </Component>
  );
}
