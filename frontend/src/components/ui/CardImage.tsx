import { useState, type ReactNode } from 'react';
import { cn } from '../../lib/cn';

interface CardImageProps {
  src?: string | null;
  alt: string;
  /** Extra classes on the frame (width etc.). */
  className?: string;
  /** Shown when src missing or load fails. */
  fallback?: ReactNode;
  /**
   * cover — fills 5/7 frame (grid + modal default).
   * contain — full card visible (detail hero).
   */
  fit?: 'contain' | 'cover';
  /** Skip fixed 5/7 when caller sets custom aspect (detail hero). */
  fixedAspect?: boolean;
}

/**
 * KKI card frame: aspect 5/7, rounded, nebula bg.
 * Default fit=cover so previews fill evenly.
 * Detail heroes: fit=contain + fixedAspect=false — letterbox is nebula;
 * img itself is rounded so white art corners don't show.
 */
export function CardImage({
  src,
  alt,
  className,
  fallback = <span className="text-4xl">🃏</span>,
  fit = 'cover',
  fixedAspect = true,
}: CardImageProps) {
  const [failed, setFailed] = useState(false);
  const showImg = Boolean(src) && !failed;

  return (
    <div
      className={cn(
        'flex items-center justify-center overflow-hidden rounded-md bg-nebula',
        fixedAspect && 'aspect-[5/7]',
        className
      )}
    >
      {showImg ? (
        <img
          src={src!}
          alt={alt}
          loading="lazy"
          decoding="async"
          className={cn(
            fit === 'contain'
              ? 'max-h-full max-w-full rounded-md object-contain'
              : 'h-full w-full object-cover'
          )}
          onError={() => setFailed(true)}
        />
      ) : (
        fallback
      )}
    </div>
  );
}
