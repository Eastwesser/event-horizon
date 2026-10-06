// frontend/src/store/towerStore.ts
import { create } from 'zustand';
import api from '../services/api';

interface TowerState {
  towerBlocks: number[];
  towerHeight: number;
  currentBlockX: number;
  blockWidth: number;
  direction: 1 | -1;
  score: number;
  level: number;
  combo: number;
  gameOver: boolean;
  started: boolean;
  boosted: boolean;
  boostId: string | null;
  lastSubmitRanked: boolean | null;
  lastSubmitMessage: string | null;
  GAME_WIDTH: number;
  GAME_HEIGHT: number;
  BASE_SPEED: number;
  startGame: (opts?: { boosted?: boolean; boostId?: string | null }) => void;
  stopLoop: () => void;
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
  towerBlocks: [],
  towerHeight: 0,
  currentBlockX: (GAME_WIDTH - INITIAL_BLOCK_WIDTH) / 2,
  blockWidth: INITIAL_BLOCK_WIDTH,
  direction: 1,
  score: 0,
  level: 1,
  combo: 0,
  gameOver: false,
  started: false,
  boosted: false,
  boostId: null,
  lastSubmitRanked: null,
  lastSubmitMessage: null,
  GAME_WIDTH,
  GAME_HEIGHT,
  BASE_SPEED,

  stopLoop: () => {
    if ((get() as { gameLoop?: ReturnType<typeof setInterval> }).gameLoop) {
      clearInterval((get() as { gameLoop?: ReturnType<typeof setInterval> }).gameLoop);
      (get() as { gameLoop?: ReturnType<typeof setInterval> }).gameLoop = undefined;
    }
  },

  startGame: (opts) => {
    get().stopLoop();

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
      started: true,
      boosted: opts?.boosted ?? false,
      boostId: opts?.boostId ?? null,
      lastSubmitRanked: null,
      lastSubmitMessage: null,
    });

    const gameLoop = setInterval(() => {
      const { gameOver, started } = get();
      if (!gameOver && started) {
        get().update();
      }
    }, 1000 / 60);

    (get() as { gameLoop?: ReturnType<typeof setInterval> }).gameLoop = gameLoop;
  },

  update: () => {
    const { currentBlockX, blockWidth, direction, BASE_SPEED, GAME_WIDTH, level, boosted } = get();
    // Boost: horizontal block speed ×0.8 (easier aim)
    const speedMul = boosted ? 0.8 : 1;
    const speed = (BASE_SPEED + Math.floor(level / 5)) * speedMul;

    let newX = currentBlockX + direction * speed;
    let newDirection = direction;

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
    const { currentBlockX, blockWidth, towerBlocks, GAME_WIDTH, score, combo, gameOver, started } =
      get();

    if (gameOver || !started) return;

    const lastBlockWidth = towerBlocks[towerBlocks.length - 1];
    const towerLeft = (GAME_WIDTH - lastBlockWidth) / 2;
    const towerRight = towerLeft + lastBlockWidth;
    const blockLeft = currentBlockX;
    const blockRight = currentBlockX + blockWidth;

    const overlapLeft = Math.max(blockLeft, towerLeft);
    const overlapRight = Math.min(blockRight, towerRight);
    const overlap = overlapRight - overlapLeft;

    if (overlap <= 0) {
      set({ gameOver: true });
      void get().submitScore();
      return;
    }

    const newBlockWidth = Math.max(overlap, 5);
    const blockScore = get().calculateBlockScore();
    const newScore = score + blockScore;
    const newCombo = combo + 1;
    const newLevel = Math.floor(newScore / 100) + 1;
    const newTowerBlocks = [...towerBlocks, newBlockWidth];

    set({
      towerBlocks: newTowerBlocks,
      towerHeight: newTowerBlocks.length,
      blockWidth: newBlockWidth,
      currentBlockX: (GAME_WIDTH - newBlockWidth) / 2,
      score: newScore,
      level: newLevel,
      combo: newCombo,
    });

    if (newBlockWidth < 10) {
      set({ gameOver: true });
      void get().submitScore();
    }
  },

  calculateBlockScore: () => {
    const { level, combo } = get();
    let multiplier = 1;
    if (combo >= 5) multiplier = combo - 2;
    else if (combo >= 4) multiplier = 3;
    else if (combo >= 3) multiplier = 2;
    return 10 * level * multiplier;
  },

  submitScore: async () => {
    const { score, boosted, boostId } = get();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const nickname = localStorage.getItem('nickname') || userEmail?.split('@')[0] || 'Игрок';

    if (!userId || !userEmail) return;

    try {
      const body: Record<string, unknown> = {
        user_id: userId,
        game_id: 'towers',
        level: 1,
        score,
        user_email: userEmail,
        nickname,
        seed: `towers_seed_${Date.now()}`,
        moves: [],
      };
      if (boosted && boostId) body.boost_id = boostId;

      const response = await api.post('/game/submit', body);

      if (response.status >= 200 && response.status < 300) {
        const ranked =
          response.data?.ranked === true &&
          !boosted &&
          !String(response.data?.message || '').includes('not ranked');
        set({
          lastSubmitRanked: ranked,
          lastSubmitMessage: String(response.data?.message || ''),
        });

        if (ranked) {
          const storageKey = `gameScores_${userId}`;
          const totalScoreKey = `totalScore_${userId}`;
          const playedKey = `towersGamesPlayed_${userId}`;
          const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
          const currentBest = savedScores.towers || 0;
          if (score > currentBest) {
            savedScores.towers = score;
            localStorage.setItem(storageKey, JSON.stringify(savedScores));
          }
          const played = parseInt(localStorage.getItem(playedKey) || '0', 10);
          localStorage.setItem(playedKey, String(played + 1));
          const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0', 10);
          localStorage.setItem(totalScoreKey, String(totalScore + score));
          void import('../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
        }
      }
    } catch (err) {
      console.error('Failed to submit towers score:', err);
    }
  },
}));
