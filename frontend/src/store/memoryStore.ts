
// frontend/src/store/memoryStore.ts
import { create } from 'zustand';

export interface Card {
  id: number;
  emoji: string;
  flipped: boolean;
  matched: boolean;
}

interface MemoryState {
  cards: Card[];
  flippedIndices: number[];
  matchedPairs: number;
  moves: number;
  combo: number;
  multiplier: number;
  gameOver: boolean;
  score: number;
  
  initGame: () => void;
  flipCard: (index: number) => void;
  checkMatch: () => void;
  resetGame: () => void;
  calculateFinalScore: () => number;
}

const FRUIT_EMOJIS = [
  '🍎', '🍒', '🍊', '🍋', '🍉',
  '🥝', '🍓', '🍑', '🥥', '🥑',
  '🍇', '🍐', '🍈', '🫐', '🍌'
];

// Перемешать массив (Fisher-Yates)
function shuffleArray<T>(arr: T[]): T[] {
  const shuffled = [...arr];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

// Создать колоду из 30 карт (15 пар)
function createDeck(): Card[] {
  const pairs: Card[] = [];
  let id = 0;
  
  for (const emoji of FRUIT_EMOJIS) {
    // Каждая пара — две одинаковые карты
    pairs.push({ id: id++, emoji, flipped: false, matched: false });
    pairs.push({ id: id++, emoji, flipped: false, matched: false });
  }
  
  return shuffleArray(pairs);
}

// Формула очков (честная и прозрачная)
function calculateScore(moves: number, totalPairs: number): number {
  const MIN_SCORE = 100;
  const MAX_SCORE = 1000;
  const IDEAL_MOVES = totalPairs; // 15 ходов = идеально
  
  if (moves <= IDEAL_MOVES) {
    return MAX_SCORE;
  }
  
  // Каждый лишний ход снижает на 20 очков, но не ниже минимума
  const penalty = (moves - IDEAL_MOVES) * 20;
  const score = Math.max(MIN_SCORE, MAX_SCORE - penalty);
  
  return score;
}

export const useMemoryStore = create<MemoryState>((set, get) => ({
  cards: [],
  flippedIndices: [],
  matchedPairs: 0,
  moves: 0,
  combo: 0,
  multiplier: 1,
  gameOver: false,
  score: 0,
  
  initGame: () => {
    const deck = createDeck();
    set({
      cards: deck,
      flippedIndices: [],
      matchedPairs: 0,
      moves: 0,
      combo: 0,
      multiplier: 1,
      gameOver: false,
      score: 0,
    });
  },
  
  flipCard: (index: number) => {
    const { cards, flippedIndices, gameOver } = get();
    
    // Нельзя кликать если игра окончена, карта уже совпала или уже открыта
    if (gameOver) return;
    if (cards[index].matched) return;
    if (flippedIndices.includes(index)) return;
    
    // Нельзя открыть больше 2 карт за раз
    if (flippedIndices.length >= 2) return;
    
    // Переворачиваем карту
    const newCards = [...cards];
    newCards[index] = { ...newCards[index], flipped: true };
    
    const newFlippedIndices = [...flippedIndices, index];
    
    set({
      cards: newCards,
      flippedIndices: newFlippedIndices,
    });
    
    // Если открыли 2 карты, проверяем совпадение
    if (newFlippedIndices.length === 2) {
      setTimeout(() => {
        get().checkMatch();
      }, 500); // Даём время посмотреть на вторую карту
    }
  },
  
  checkMatch: () => {
    const { cards, flippedIndices, moves, combo, matchedPairs } = get();
    
    if (flippedIndices.length !== 2) return;
    
    const [idx1, idx2] = flippedIndices;
    const card1 = cards[idx1];
    const card2 = cards[idx2];
    
    const isMatch = card1.emoji === card2.emoji;
    const newMoves = moves + 1;
    
    let newCards = [...cards];
    let newMatchedPairs = matchedPairs;
    let newCombo = combo;
    let newMultiplier = 1;
    
    if (isMatch) {
      // Карты совпали — помечаем как matched (они исчезнут)
      newCards[idx1] = { ...newCards[idx1], matched: true, flipped: false };
      newCards[idx2] = { ...newCards[idx2], matched: true, flipped: false };
      newMatchedPairs = matchedPairs + 1;
      
      // Увеличиваем комбо за правильную пару
      newCombo = combo + 1;
      
      // Множитель: 2 пары подряд = x2, 4+ подряд = x3
      if (newCombo >= 4) {
        newMultiplier = 3;
      } else if (newCombo >= 2) {
        newMultiplier = 2;
      } else {
        newMultiplier = 1;
      }
    } else {
      // Карты не совпали — переворачиваем обратно
      newCards[idx1] = { ...newCards[idx1], flipped: false };
      newCards[idx2] = { ...newCards[idx2], flipped: false };
      
      // Сбрасываем комбо при ошибке
      newCombo = 0;
      newMultiplier = 1;
    }
    
    // Проверяем, все ли пары найдены
    const totalPairs = FRUIT_EMOJIS.length;
    const gameOver = newMatchedPairs === totalPairs;
    
    let finalScore = 0;
    if (gameOver) {
      finalScore = calculateScore(newMoves, totalPairs);
    }
    
    set({
      cards: newCards,
      flippedIndices: [],
      moves: newMoves,
      matchedPairs: newMatchedPairs,
      combo: newCombo,
      multiplier: newMultiplier,
      gameOver,
      score: finalScore,
    });
    
    // Если игра окончена, можно потом вызвать submitScore
    if (gameOver) {
      console.log(`🎉 Game Over! Moves: ${newMoves}, Score: ${finalScore}`);
      // отправка в лидерборд
    }
  },
  
  resetGame: () => {
    get().initGame();
  },
  
  calculateFinalScore: () => {
    const { moves } = get();
    return calculateScore(moves, FRUIT_EMOJIS.length);
  },
}));