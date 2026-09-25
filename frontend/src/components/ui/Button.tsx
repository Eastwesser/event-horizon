import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/cn';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  children: ReactNode;
}

const variantClasses: Record<Variant, string> = {
  /* Gold = deliberate accent CTA only */
  primary: 'bg-horizon-gold text-void hover:bg-horizon-gold-hot shadow-glow-gold',
  /* Indigo secondary — brand identity for non-primary actions */
  secondary:
    'bg-transparent border border-indigo/40 text-indigo-soft hover:bg-indigo/10 hover:border-indigo',
  ghost: 'bg-white/5 text-text-primary border border-white/10 hover:bg-white/10',
  danger: 'bg-error/10 border border-error/50 text-error hover:bg-error/20',
};

const sizeClasses: Record<Size, string> = {
  sm: 'px-4 py-2 text-sm rounded-sm',
  md: 'px-6 py-3 text-base rounded-md',
};

export function Button({
  variant = 'primary',
  size = 'md',
  className,
  children,
  disabled,
  ...rest
}: ButtonProps) {
  return (
    <button
      className={cn(
        'font-body font-semibold transition-[transform,background-color,box-shadow] duration-200',
        'disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:scale-100',
        'active:scale-[0.98] hover:scale-[1.02]',
        variantClasses[variant],
        sizeClasses[size],
        className,
      )}
      disabled={disabled}
      {...rest}
    >
      {children}
    </button>
  );
}
