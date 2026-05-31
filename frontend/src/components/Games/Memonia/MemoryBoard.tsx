// frontend/src/components/Games/Memonia/MemoryBoard.tsx
import { useMemoryStore } from '../../../store/memoryStore';
import { MemoryCard } from './MemoryCard';

export function MemoryBoard() {
  const { cards, flipCard, gameOver } = useMemoryStore();
  
  if (!cards.length) {
    return <div className="memory-board__empty">Загрузка...</div>;
  }
  
  return (
    <div className="memory-board">
      {cards.map((card, index) => (
        <MemoryCard
          key={card.id}
          emoji={card.emoji}
          flipped={card.flipped}
          matched={card.matched}
          onClick={() => flipCard(index)}
          disabled={gameOver}
        />
      ))}
    </div>
  );
}
