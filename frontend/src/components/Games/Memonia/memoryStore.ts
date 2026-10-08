// frontend/src/store/memoryStore.ts
import { create } from 'zustand';
import api from '../../../services/api';

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
  boosted: boolean;
  boostId: string | null;
  lastSubmitRanked: boolean | null;

  initGame: (opts?: { boosted?: boolean; boostId?: string | null }) => void;
  flipCard: (index: number) => void;
  checkMatch: () => void;
  resetGame: (opts?: { boosted?: boolean; boostId?: string | null }) => void;
  flashBoostHint: () => void;
  calculateFinalScore: () => number;
  submitScore: () => Promise<void>;
}

/** Shared deck + animal-skin map (same order → same animal). */
export const FRUIT_EMOJIS = [
  '🍎', '🍒', '🍊', '🍋', '🍉',
  '🥝', '🍓', '🍑', '🥥', '🥑',
  '🍇', '🍐', '🍈', '🫐', '🍌',
] as const;

export const ANIMAL_EMOJIS = [
  '🐶', '🐱', '🐭', '🐹', '🐰',
  '🦊', '🐻', '🐼', '🐨', '🐯',
  '🦁', '🐮', '🐷', '🐸', '🐵',
] as const;

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
  boosted: false,
  boostId: null,
  lastSubmitRanked: null,

  initGame: (opts) => {
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
      boosted: opts?.boosted ?? false,
      boostId: opts?.boostId ?? null,
      lastSubmitRanked: null,
    });
    if (opts?.boosted) {
      // Defer so cards paint face-down first, then flash one pair.
      setTimeout(() => get().flashBoostHint(), 80);
    }
  },

  flashBoostHint: () => {
    const { cards, boosted } = get();
    if (!boosted || cards.length === 0) return;
    const byEmoji = new Map<string, number[]>();
    cards.forEach((c, i) => {
      if (c.matched) return;
      const list = byEmoji.get(c.emoji) || [];
      list.push(i);
      byEmoji.set(c.emoji, list);
    });
    let pair: number[] | null = null;
    for (const idxs of byEmoji.values()) {
      if (idxs.length >= 2) {
        pair = idxs.slice(0, 2);
        break;
      }
    }
    if (!pair) return;
    const [a, b] = pair;
    const shown = cards.map((c, i) =>
      i === a || i === b ? { ...c, flipped: true } : c,
    );
    set({ cards: shown });
    setTimeout(() => {
      const cur = get().cards;
      set({
        cards: cur.map((c, i) =>
          i === a || i === b ? { ...c, flipped: false } : c,
        ),
      });
    }, 1200);
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
      
      // 💾 Сохраняем статистику в localStorage
      const savedScores = JSON.parse(localStorage.getItem('gameScores') || '{}');
      const currentBest = savedScores.memory || 0;
      
      if (finalScore > currentBest) {
        savedScores.memory = finalScore;
        localStorage.setItem('gameScores', JSON.stringify(savedScores));
      }
      
      const played = parseInt(localStorage.getItem('memoryGamesPlayed') || '0');
      localStorage.setItem('memoryGamesPlayed', String(played + 1));
      
      const totalScore = parseInt(localStorage.getItem('totalScore') || '0');
      localStorage.setItem('totalScore', String(totalScore + finalScore));
      
      if (import.meta.env.DEV) {
        console.log(`Memonia game over: moves=${newMoves} score=${finalScore}`);
      }
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
  },
  
  resetGame: (opts) => {
    get().initGame(opts);
  },
  
  calculateFinalScore: () => {
    const { moves } = get();
    return calculateScore(moves, FRUIT_EMOJIS.length);
  },
  
  submitScore: async () => {
    const { score, boosted, boostId } = get();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail') || '';
    const nickname = localStorage.getItem('nickname') || userEmail.split('@')[0] || 'Игрок';

    if (!userId || !userEmail) return;

    try {
      const body: Record<string, unknown> = {
        user_id: userId,
        game_id: 'memory',
        level: 1,
        score,
        user_email: userEmail,
        nickname,
        seed: `memory_seed_${Date.now()}`,
        moves: [],
      };
      if (boosted && boostId) body.boost_id = boostId;

      const response = await api.post('/game/submit', body);

      if (response.status >= 200 && response.status < 300) {
        const ranked =
          response.data?.ranked === true &&
          !boosted &&
          !String(response.data?.message || '').includes('not ranked');
        set({ lastSubmitRanked: ranked });
        if (!ranked) return;

        const storageKey = `gameScores_${userId}`;
        const totalScoreKey = `totalScore_${userId}`;
        const playedKey = `memoryGamesPlayed_${userId}`;
        const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
        const currentBest = savedScores.memory || 0;
        if (score > currentBest) {
          savedScores.memory = score;
          localStorage.setItem(storageKey, JSON.stringify(savedScores));
        }
        const played = parseInt(localStorage.getItem(playedKey) || '0', 10);
        localStorage.setItem(playedKey, String(played + 1));
        const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0', 10);
        localStorage.setItem(totalScoreKey, String(totalScore + score));
        void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
      }
    } catch (err) {
      console.error('Failed to submit memory score:', err);
    }
  },
}));