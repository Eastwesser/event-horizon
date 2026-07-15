// frontend/src/components/Games/Memonia/MemoryBoard.tsx
import { useMemoryStore } from '../../../store/memoryStore';
import { MemoryCard } from './MemoryCard';

interface MemoryBoardProps {
  skin?: 'default' | 'animals';
}

// Эмодзи для разных скинов
// const defaultEmojis = ['🍎', '🍊', '🍋', '🍇', '🍓', '🍑', '🍒', '🍉', '🥝', '🍍', '🥭', '🍌', '🍈', '🍏', '🍐'];
const defaultEmojis = ['🍎', '🍊', '🍋', '🍇', '🍓', '🍑', '🍒', '🍉', '🥝', '🍍', '🥭', '🍌', '🍈', '🍏', '🍐', '🥑', '🥥', '🫐'];
// const animalEmojis = ['🐶', '🐱', '🐭', '🐹', '🐰', '🦊', '🐻', '🐼', '🐨', '🐯', '🦁', '🐮', '🐷', '🐸', '🐵'];
const animalEmojis = ['🐶', '🐱', '🐭', '🐹', '🐰', '🦊', '🐻', '🐼', '🐨', '🐯', '🦁', '🐮', '🐷', '🐸', '🐵', '🦝', '🦊', '🐺'];

export function MemoryBoard({ skin = 'default' }: MemoryBoardProps) {
  const { cards, flipCard, gameOver } = useMemoryStore();
  
  // Подменяем эмодзи на скиновые, если они есть
  const getCardEmoji = (originalEmoji: string) => {
    if (skin === 'animals') {
      const index = defaultEmojis.indexOf(originalEmoji);
      if (index !== -1 && index < animalEmojis.length) {
        return animalEmojis[index];
      }
      // 🆕 Если эмодзи не найден в списке - возвращаем его как есть
      console.warn('⚠️ Эмодзи не найден в списке:', originalEmoji);
      return originalEmoji;
    }
    return originalEmoji;
  };
  
  if (!cards.length) {
    return <div className="memory-board__empty">Загрузка...</div>;
  }
  
  return (
    <div className="memory-board">
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