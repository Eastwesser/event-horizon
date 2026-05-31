// frontend/src/components/Games/Memonia/MemoryCard.tsx
import { memo } from 'react';

interface MemoryCardProps {
  emoji: string;
  flipped: boolean;
  matched: boolean;
  onClick: () => void;
  disabled?: boolean;
}

export const MemoryCard = memo(({ emoji, flipped, matched, onClick, disabled }: MemoryCardProps) => {
  if (matched) {
    return <div className="memory-card memory-card--matched" />;
  }
  
  return (
    <div
      className={`memory-card ${flipped ? 'memory-card--flipped' : ''}`}
      onClick={() => !disabled && !flipped && onClick()}
    >
      <div className="memory-card__inner">
        <div className="memory-card__front">
          <span className="memory-card__emoji">{emoji}</span>
        </div>
        <div className="memory-card__back">
          <span className="memory-card__question">?</span>
        </div>
      </div>
    </div>
  );
});

MemoryCard.displayName = 'MemoryCard';