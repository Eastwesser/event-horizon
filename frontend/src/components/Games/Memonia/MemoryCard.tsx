// frontend/src/components/Games/Memonia/MemoryCard.tsx
import { memo } from 'react';

interface MemoryCardProps {
  emoji: string;
  flipped: boolean;
  matched: boolean;
  onClick: () => void;
  disabled?: boolean;
  skin?: 'default' | 'animals';
}

export const MemoryCard = memo(({ 
  emoji, 
  flipped, 
  matched, 
  onClick, 
  disabled,
  skin = 'default' 
}: MemoryCardProps) => {
  if (matched) {
    return <div className="memory-card memory-card--matched" />;
  }
  
  // Стиль для скина - всегда применяем, а не только при flipped!
  const cardStyle = skin === 'animals' ? {
    background: 'linear-gradient(135deg, #2d1b69, #1a1a2e)',
    border: '2px solid #4ADE80'
  } : {};
  
  const frontStyle = skin === 'animals' ? {
    background: 'linear-gradient(135deg, #2d1b69, #1a1a2e)',
    border: '2px solid #4ADE80'
  } : {};
  
  return (
    <div
      className={`memory-card ${flipped ? 'memory-card--flipped' : ''} ${skin === 'animals' ? 'memory-card--animals' : ''}`}
      onClick={() => !disabled && !flipped && onClick()}
    >
      <div className="memory-card__inner">
        <div className="memory-card__front" style={frontStyle}>
          <span className="memory-card__emoji">{emoji}</span>
        </div>
        <div className="memory-card__back" style={cardStyle}>
          <span className="memory-card__question">
            {skin === 'animals' ? '🐾' : '?'}
          </span>
        </div>
      </div>
    </div>
  );
});

MemoryCard.displayName = 'MemoryCard';