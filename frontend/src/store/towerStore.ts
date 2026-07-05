// frontend/src/store/towerStore.ts
import { create } from 'zustand';

interface TowerState {
  // Башня
  towerBlocks: number[];
  towerHeight: number;
  
  // Текущий блок
  currentBlockX: number;
  blockWidth: number;
  direction: 1 | -1;
  
  // Игровое состояние
  score: number;
  level: number;
  combo: number;
  gameOver: boolean;
  
  // Константы
  GAME_WIDTH: number;
  GAME_HEIGHT: number;
  BASE_SPEED: number;
  
  // Actions
  startGame: () => void;
  dropBlock: () => void;
  update: () => void;
  calculateBlockScore: () => number;
  submitScore: () => Promise<void>;
}

const GAME_WIDTH = 400;
const GAME_HEIGHT = 500;
const BASE_SPEED = 3;
const INITIAL_BLOCK_WIDTH = 100;

export const useTowerStore = create<TowerState>((set, get) => ({
  // Начальные значения
  towerBlocks: [],
  towerHeight: 0,
  currentBlockX: (GAME_WIDTH - INITIAL_BLOCK_WIDTH) / 2,
  blockWidth: INITIAL_BLOCK_WIDTH,
  direction: 1,
  score: 0,
  level: 1,
  combo: 0,
  gameOver: false,
  
  GAME_WIDTH,
  GAME_HEIGHT,
  BASE_SPEED,
  
  startGame: () => {
    // Останавливаем предыдущий цикл
    if ((get() as any).gameLoop) {
      clearInterval((get() as any).gameLoop);
    }
    
    set({
      towerBlocks: [INITIAL_BLOCK_WIDTH],
      towerHeight: 1,
      currentBlockX: (GAME_WIDTH - INITIAL_BLOCK_WIDTH) / 2,
      blockWidth: INITIAL_BLOCK_WIDTH,
      direction: 1,
      score: 0,
      level: 1,
      combo: 0,
      gameOver: false,
    });
    
    // Запускаем игровой цикл
    const gameLoop = setInterval(() => {
      const { gameOver } = get();
      if (!gameOver) {
        get().update();
      }
    }, 1000 / 60);
    
    (get() as any).gameLoop = gameLoop;
  },
  
  update: () => {
    const { currentBlockX, blockWidth, direction, BASE_SPEED, GAME_WIDTH, level } = get();
    
    // Скорость увеличивается с уровнем
    const speed = BASE_SPEED + Math.floor(level / 5);
    
    let newX = currentBlockX + (direction * speed);
    let newDirection = direction;
    
    // Отскок от границ
    if (newX <= 0) {
      newX = 0;
      newDirection = 1;
    } else if (newX + blockWidth >= GAME_WIDTH) {
      newX = GAME_WIDTH - blockWidth;
      newDirection = -1;
    }
    
    set({
      currentBlockX: newX,
      direction: newDirection,
    });
  },
  
  dropBlock: () => {
    const { currentBlockX, blockWidth, towerBlocks, GAME_WIDTH, score, combo, gameOver } = get();
    
    if (gameOver) return;
    
    const lastBlockWidth = towerBlocks[towerBlocks.length - 1];
    const towerLeft = (GAME_WIDTH - lastBlockWidth) / 2;
    const towerRight = towerLeft + lastBlockWidth;
    const blockLeft = currentBlockX;
    const blockRight = currentBlockX + blockWidth;
    
    // Вычисляем перекрытие между блоком и башней
    const overlapLeft = Math.max(blockLeft, towerLeft);
    const overlapRight = Math.min(blockRight, towerRight);
    const overlap = overlapRight - overlapLeft;
    
    console.log('🔍 Drop debug:', {
      blockLeft, blockRight,
      towerLeft, towerRight,
      overlap
    });
    
    // Если нет перекрытия — полный промах
    if (overlap <= 0) {
      console.log('💀 Полный промах! Game Over');
      set({ gameOver: true });
      get().submitScore();
      return;
    }
    
    // Новая ширина блока = перекрытие
    const newBlockWidth = Math.max(overlap, 5);
    
    // Рассчитываем очки
    const blockScore = get().calculateBlockScore();
    const newScore = score + blockScore;
    
    // Обновляем комбо (если есть перекрытие, комбо увеличивается)
    let newCombo = combo + 1;
    
    // Обновляем уровень (каждые 100 очков)
    const newLevel = Math.floor(newScore / 100) + 1;
    
    // Добавляем блок в башню
    const newTowerBlocks = [...towerBlocks, newBlockWidth];
    
    console.log('✅ Блок добавлен:', {
      oldWidth: lastBlockWidth,
      newWidth: newBlockWidth,
      overlap,
      score: blockScore,
      totalScore: newScore,
      combo: newCombo
    });
    
    set({
      towerBlocks: newTowerBlocks,
      towerHeight: newTowerBlocks.length,
      blockWidth: newBlockWidth,
      currentBlockX: (GAME_WIDTH - newBlockWidth) / 2,
      score: newScore,
      level: newLevel,
      combo: newCombo,
    });
    
    // Проверка на Game Over (слишком узкий блок)
    if (newBlockWidth < 10) {
      console.log('💀 Блок слишком узкий! Game Over');
      set({ gameOver: true });
      get().submitScore();
    }
  },
  
  calculateBlockScore: () => {
    const { level, combo } = get();
    let multiplier = 1;
    
    if (combo >= 5) multiplier = combo - 2;
    else if (combo >= 4) multiplier = 3;
    else if (combo >= 3) multiplier = 2;
    
    const score = 10 * level * multiplier;
    console.log(`📊 Очки за блок: 10 × ${level} × ${multiplier} = ${score}`);
    return score;
  },
  
  submitScore: async () => {
    const { score } = get();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const nickname = localStorage.getItem('nickname') || userEmail?.split('@')[0] || 'Игрок';
    
    console.log('💾 Сохранение рекорда:', { userId, userEmail, nickname, score });
    
    if (!userId || !userEmail) {
        console.log('❌ Нет userId или userEmail');
        return;
    }
    
    try {
        const response = await fetch('/api/game/submit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: userId,
                game_id: 'towers',
                level: 1,
                score: score,
                user_email: userEmail,
                nickname: nickname,
                seed: `towers_seed_${Date.now()}`,
                moves: [],
            }),
        });
        
        if (response.ok) {
            console.log(`✅ Towers score submitted: ${score}`);
            
            // 💾 Сохраняем статистику с привязкой к userId
            // Используем существующий userId
            const storageKey = `gameScores_${userId}`;
            const totalScoreKey = `totalScore_${userId}`;
            const playedKey = `towersGamesPlayed_${userId}`;
            
            const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
            const currentBest = savedScores.towers || 0;
            
            if (score > currentBest) {
                savedScores.towers = score;
                localStorage.setItem(storageKey, JSON.stringify(savedScores));
            }
            
            const played = parseInt(localStorage.getItem(playedKey) || '0');
            localStorage.setItem(playedKey, String(played + 1));
            
            const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0');
            localStorage.setItem(totalScoreKey, String(totalScore + score));
        }
    } catch (err) {
        console.error('Failed to submit towers score:', err);
    }
  },
}));
