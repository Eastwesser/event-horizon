import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { cn } from '../../lib/cn';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  children: ReactNode;
  className?: string;
  /** Bottom keeps the playfield visible above the card. */
  align?: 'center' | 'bottom';
}

/** Portaled to document.body so fixed centering is never trapped by GameShell / transformed ancestors. */
export function Modal({ open, onClose, title, children, className, align = 'center' }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onKeyDown);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKeyDown);
      document.body.style.overflow = prev;
    };
  }, [open, onClose]);

  if (!open) return null;

  return createPortal(
    <div
      className={cn(
        'fixed inset-0 z-[100] flex p-4 sm:p-6',
        align === 'bottom' ? 'items-end justify-center bg-void/35' : 'items-center justify-center bg-void/80',
      )}
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'eh-modal-title' : undefined}
        // className={cn(
        //   'eh-ring w-full max-w-lg rounded-lg border border-indigo/25 bg-nebula-elevated p-6 shadow-elevated sm:p-8',
        //   'max-h-[90vh] overflow-y-auto',
        //   className,
        // )}
        className={cn(
          'eh-ring flex w-full max-w-lg flex-col rounded-lg border border-indigo/25 bg-nebula-elevated shadow-elevated',
          'max-h-[90vh]',
          className,
        )}
        onClick={(e) => e.stopPropagation()}
      >
        {/* {title && (
          <h2 id="eh-modal-title" className="mb-4 font-display text-2xl font-semibold text-text-primary">
            {title}
          </h2>
        )}
        {children} */}
        {title && (
          <div className="border-b border-white/10 px-6 pb-4 pt-6 sm:px-8 sm:pt-8">
            <h2 id="eh-modal-title" className="font-display text-2xl font-semibold text-text-primary">
              {title}
            </h2>
          </div>
        )}
        <div className="flex-1 overflow-y-auto px-6 py-4 sm:px-8 sm:py-6">
          {children}
        </div>
        
      </div>
    </div>,
    document.body,
  );
}
