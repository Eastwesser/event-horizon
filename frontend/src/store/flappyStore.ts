// frontend/src/store/flappyStore.ts
import { create } from 'zustand';
import api from '../services/api';

export interface Pipe {
  id: number;
  x: number;
  topHeight: number;
  bottomY: number;
  passed: boolean;
}

interface FlappyState {
  // Игровое состояние
  birdY: number;
  birdVelocity: number;
  pipes: Pipe[];
  score: number;
  gameOver: boolean;
  started: boolean;
  
  // Константы
  GRAVITY: number;
  JUMP_FORCE: number;
  PIPE_WIDTH: number;
  PIPE_GAP: number;
  PIPE_SPACING: number;
  PIPE_SPEED: number;
  
  // Actions
  startGame: () => void;
  jump: () => void;
  updateGame: () => void;
  resetGame: () => void;
  generatePipe: () => Pipe;
  checkCollisions: () => boolean;
  submitScore: () => Promise<void>;
}

const GAME_HEIGHT = 500;
const GAME_WIDTH = 800;
const BIRD_SIZE = 30;

export const useFlappyStore = create<FlappyState>((set, get) => ({
  // Начальные значения
  birdY: GAME_HEIGHT / 2,
  birdVelocity: 0,
  pipes: [],
  score: 0,
  gameOver: false,
  started: false,
  
  // Константы физики
  GRAVITY: 0.3,
  JUMP_FORCE: -6.5,
  PIPE_WIDTH: 60,
  PIPE_GAP: 150,
  PIPE_SPACING: 300,
  PIPE_SPEED: 3,
  
  startGame: () => {
    // Останавливаем предыдущий цикл если есть
    if ((get() as any).gameLoop) {
      clearInterval((get() as any).gameLoop);
    }
    
    set({
      started: true,
      gameOver: false,
      birdY: GAME_HEIGHT / 2,
      birdVelocity: 0,
      pipes: [],
      score: 0,
    });
    
    // Генерируем первую трубу
    const firstPipe = get().generatePipe();
    set({ pipes: [firstPipe] });
    
    // Запускаем игровой цикл
    const gameLoop = setInterval(() => {
      const { gameOver, started } = get();
      if (!gameOver && started) {
        get().updateGame();
      } else if (gameOver) {
        clearInterval(gameLoop);
      }
    }, 1000 / 60); // 60 FPS
    
    // Сохраняем gameLoop для очистки
    (get() as any).gameLoop = gameLoop;
  },
  
  jump: () => {
    const { gameOver, started } = get();
    if (!gameOver && started) {
      set({ birdVelocity: get().JUMP_FORCE });
    } else if (!started) {
      get().startGame();
    }
  },
  
  updateGame: () => {
    const state = get();
    const { birdVelocity, GRAVITY, pipes, PIPE_SPEED, PIPE_WIDTH, birdY, score } = state;
    
    // Обновляем физику птички
    const newVelocity = birdVelocity + GRAVITY;
    const newBirdY = birdY + newVelocity;
    
    // Обновляем позиции труб
    const updatedPipes = pipes
      .map(pipe => ({
        ...pipe,
        x: pipe.x - PIPE_SPEED,
      }))
      .filter(pipe => pipe.x + PIPE_WIDTH > 0);
    
    // Проверяем прохождение труб (подсчёт очков)
    let newScore = score;
    const pipesWithScore = updatedPipes.map(pipe => {
      if (!pipe.passed && pipe.x + PIPE_WIDTH < 100) {
        newScore = newScore + 10;
        return { ...pipe, passed: true };
      }
      return pipe;
    });
    
    // Генерируем новую трубу если нужно
    let newPipes = pipesWithScore;
    if (pipesWithScore.length === 0 || 
        pipesWithScore[pipesWithScore.length - 1].x < GAME_WIDTH - get().PIPE_SPACING) {
      const newPipe = get().generatePipe();
      newPipes = [...pipesWithScore, newPipe];
    }
    
    // Обновляем состояние
    set({
      birdY: newBirdY,
      birdVelocity: newVelocity,
      pipes: newPipes,
      score: newScore,
    });
    
    // Проверяем коллизии
    const hasCollision = get().checkCollisions();
    if (hasCollision || newBirdY > GAME_HEIGHT - BIRD_SIZE || newBirdY < 0) {
      set({ gameOver: true });
      get().submitScore();
    }
  },
  
  resetGame: () => {
    // Останавливаем текущий цикл
    if ((get() as any).gameLoop) {
      clearInterval((get() as any).gameLoop);
    }
    get().startGame();
  },
  
  generatePipe: () => {
    const { PIPE_GAP } = get();
    const minTop = 50;
    const maxTop = GAME_HEIGHT - PIPE_GAP - 50;
    const topHeight = Math.random() * (maxTop - minTop) + minTop;
    const bottomY = topHeight + PIPE_GAP;
    
    return {
      id: Date.now(),
      x: GAME_WIDTH,
      topHeight,
      bottomY,
      passed: false,
    };
  },
  
  checkCollisions: () => {
    const { birdY, pipes, PIPE_WIDTH } = get();
    const birdX = 100;
    const birdSize = BIRD_SIZE;
    
    for (const pipe of pipes) {
      if (birdX + birdSize > pipe.x && birdX < pipe.x + PIPE_WIDTH) {
        if (birdY < pipe.topHeight || birdY + birdSize > pipe.bottomY) {
          return true;
        }
      }
    }
    
    return false;
  },
  
  submitScore: async () => {
    const { score } = get();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const nickname = localStorage.getItem('nickname') || userEmail?.split('@')[0] || 'Игрок';

    if (!userId || !userEmail) return;
    
    try {
        const response = await api.post('/game/submit', {
                user_id: userId,
                game_id: 'flappy',
                level: 1,
                score: score,
                user_email: userEmail,
                nickname: nickname,
                seed: `flappy_seed_${Date.now()}`,
                moves: [],
        });
        
        if (response.status >= 200 && response.status < 300) {
            console.log(`✅ Flappy score submitted: ${score}`);
            
            // 💾 Сохраняем статистику с привязкой к userId
            // Используем существующий userId
            const storageKey = `gameScores_${userId}`;
            const totalScoreKey = `totalScore_${userId}`;
            const playedKey = `flappyGamesPlayed_${userId}`;
            
            const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
            const currentBest = savedScores.flappy || 0;
            
            if (score > currentBest) {
                savedScores.flappy = score;
                localStorage.setItem(storageKey, JSON.stringify(savedScores));
            }
            
            const played = parseInt(localStorage.getItem(playedKey) || '0');
            localStorage.setItem(playedKey, String(played + 1));
            
            const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0');
            localStorage.setItem(totalScoreKey, String(totalScore + score));
        }
    } catch (err) {
        console.error('Failed to submit flappy score:', err);
    }
  },
}));
