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
  /** In-run floor height (UI only — never sent as submit level). */
  floor: number;
  /** Difficulty 1–10 (leaderboard partition + physics). */
  difficulty: number;
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
  startGame: (opts?: {
    boosted?: boolean;
    boostId?: string | null;
    difficulty?: number;
  }) => void;
  stopLoop: () => void;
  dropBlock: () => void;
  update: () => void;
  calculateBlockScore: () => number;
  submitScore: () => Promise<void>;
}

const GAME_WIDTH = 400;
const GAME_HEIGHT = 500;
const BASE_SPEED = 3;

function clampDiff(d: number): number {
  if (!Number.isFinite(d)) return 1;
  return Math.min(10, Math.max(1, Math.round(d)));
}

/** L1 ≈ 100px start, L10 ≈ 55px. */
function startWidthForDifficulty(d: number): number {
  return Math.max(55, 100 - (clampDiff(d) - 1) * 5);
}

/** Extra overhang (px) still counted as a hit. L1=14, L10=0. */
function softFailPx(d: number): number {
  return Math.max(0, Math.round(14 - (clampDiff(d) - 1) * 1.55));
}

/** Min block width before hard fail. L1=6, L10=12. */
function minWidthForDifficulty(d: number): number {
  return Math.min(12, 6 + Math.floor((clampDiff(d) - 1) * 0.67));
}

export const useTowerStore = create<TowerState>((set, get) => ({
  towerBlocks: [],
  towerHeight: 0,
  currentBlockX: (GAME_WIDTH - 100) / 2,
  blockWidth: 100,
  direction: 1,
  score: 0,
  floor: 1,
  difficulty: 1,
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
    const difficulty = clampDiff(opts?.difficulty ?? get().difficulty);
    const w = startWidthForDifficulty(difficulty);

    set({
      towerBlocks: [w],
      towerHeight: 1,
      currentBlockX: (GAME_WIDTH - w) / 2,
      blockWidth: w,
      direction: 1,
      score: 0,
      floor: 1,
      difficulty,
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
    const { currentBlockX, blockWidth, direction, BASE_SPEED, GAME_WIDTH, difficulty, boosted } =
      get();
    const speedMul = boosted ? 0.55 : 1;
    const speed = (BASE_SPEED + (difficulty - 1) * 0.4) * speedMul;

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
    const {
      currentBlockX,
      blockWidth,
      towerBlocks,
      GAME_WIDTH,
      score,
      combo,
      gameOver,
      started,
      difficulty,
    } = get();

    if (gameOver || !started) return;

    const lastBlockWidth = towerBlocks[towerBlocks.length - 1];
    const towerLeft = (GAME_WIDTH - lastBlockWidth) / 2;
    const towerRight = towerLeft + lastBlockWidth;
    const grace = softFailPx(difficulty);
    const blockLeft = currentBlockX;
    const blockRight = currentBlockX + blockWidth;

    const overlapLeft = Math.max(blockLeft, towerLeft - grace);
    const overlapRight = Math.min(blockRight, towerRight + grace);
    const overlap = overlapRight - overlapLeft;

    if (overlap <= 0) {
      set({ gameOver: true });
      void get().submitScore();
      return;
    }

    // Soft-fail: grace hit keeps a thin slab; perfect stack keeps real overlap.
    const realOverlapLeft = Math.max(blockLeft, towerLeft);
    const realOverlapRight = Math.min(blockRight, towerRight);
    const realOverlap = realOverlapRight - realOverlapLeft;
    const newBlockWidth = Math.max(realOverlap > 0 ? realOverlap : Math.min(overlap, 8), 5);

    const blockScore = get().calculateBlockScore();
    const newScore = score + blockScore;
    const newCombo = combo + 1;
    const newTowerBlocks = [...towerBlocks, newBlockWidth];

    set({
      towerBlocks: newTowerBlocks,
      towerHeight: newTowerBlocks.length,
      blockWidth: newBlockWidth,
      currentBlockX: (GAME_WIDTH - newBlockWidth) / 2,
      score: newScore,
      floor: newTowerBlocks.length,
      combo: newCombo,
    });

    if (newBlockWidth < minWidthForDifficulty(difficulty)) {
      set({ gameOver: true });
      void get().submitScore();
    }
  },

  calculateBlockScore: () => {
    const { difficulty, combo } = get();
    let multiplier = 1;
    if (combo >= 5) multiplier = combo - 2;
    else if (combo >= 4) multiplier = 3;
    else if (combo >= 3) multiplier = 2;
    return 10 * difficulty * multiplier;
  },

  submitScore: async () => {
    const { score, boosted, boostId, difficulty } = get();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const nickname = localStorage.getItem('nickname') || userEmail?.split('@')[0] || 'Игрок';

    if (!userId || !userEmail) return;

    try {
      const body: Record<string, unknown> = {
        user_id: userId,
        game_id: 'towers',
        level: difficulty,
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
