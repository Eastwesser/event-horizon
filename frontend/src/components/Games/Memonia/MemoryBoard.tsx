// frontend/src/components/Games/Memonia/MemoryBoard.tsx
import { useEffect, useRef, useState } from 'react';
import { useMemoryStore } from '../../../store/memoryStore';
import { MemoryCard } from './MemoryCard';

interface MemoryBoardProps {
  skin?: 'default' | 'animals';
}

// Эмодзи для разных скинов
const defaultEmojis = ['🍎', '🍊', '🍋', '🍇', '🍓', '🍑', '🍒', '🍉', '🥝', '🍍', '🥭', '🍌', '🍈', '🍏', '🍐', '🥑', '🥥', '🫐'];
const animalEmojis = ['🐶', '🐱', '🐭', '🐹', '🐰', '🦊', '🐻', '🐼', '🐨', '🐯', '🦁', '🐮', '🐷', '🐸', '🐵', '🦝', '🦉', '🐺'];

export function MemoryBoard({ skin = 'default' }: MemoryBoardProps) {
  const { cards, flipCard, gameOver, flippedIndices, matchedPairs } = useMemoryStore();
  const [boardMismatch, setBoardMismatch] = useState(false);
  const prevRef = useRef({ flippedLen: 0, matchedPairs: 0 });

  // UI-local mismatch juice: observe flip clear without a new match.
  useEffect(() => {
    const prev = prevRef.current;
    const flippedLen = flippedIndices.length;

    if (prev.flippedLen === 2 && flippedLen === 0 && matchedPairs === prev.matchedPairs) {
      setBoardMismatch(true);
      const timer = window.setTimeout(() => setBoardMismatch(false), 420);
      prevRef.current = { flippedLen, matchedPairs };
      return () => window.clearTimeout(timer);
    }

    prevRef.current = { flippedLen, matchedPairs };
  }, [flippedIndices, matchedPairs]);

  const getCardEmoji = (originalEmoji: string) => {
    if (skin === 'animals') {
      const index = defaultEmojis.indexOf(originalEmoji);
      if (index !== -1 && index < animalEmojis.length) {
        return animalEmojis[index];
      }
      return originalEmoji;
    }
    return originalEmoji;
  };

  if (!cards.length) {
    return <div className="memory-board__empty">Загрузка...</div>;
  }

  return (
    <div
      className={['memory-board', boardMismatch ? 'memory-board--mismatch' : '']
        .filter(Boolean)
        .join(' ')}
    >
      {cards.map((card, index) => (
        <MemoryCard
          key={card.id}
          emoji={getCardEmoji(card.emoji)}
          flipped={card.flipped}
          matched={card.matched}
          onClick={() => flipCard(index)}
          disabled={gameOver}
          skin={skin}
        />
      ))}
    </div>
  );
}
