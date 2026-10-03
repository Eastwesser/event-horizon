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

const BASE_PIPE_SPEED = 3;
const BASE_GRAVITY = 0.3;
const BOOST_DURATION_MS = 5000;

interface FlappyState {
  birdY: number;
  birdVelocity: number;
  pipes: Pipe[];
  score: number;
  gameOver: boolean;
  started: boolean;

  GRAVITY: number;
  JUMP_FORCE: number;
  PIPE_WIDTH: number;
  PIPE_GAP: number;
  PIPE_SPACING: number;
  PIPE_SPEED: number;

  boostId: string | null;
  boosted: boolean;
  lastSubmitRanked: boolean | null;
  lastSubmitMessage: string | null;

  startGame: (opts?: { boostId?: string | null; boosted?: boolean }) => void;
  jump: () => void;
  updateGame: () => void;
  resetGame: () => void;
  clearBoost: () => void;
  generatePipe: () => Pipe;
  checkCollisions: () => boolean;
  submitScore: () => Promise<void>;
}

const GAME_HEIGHT = 500;
const GAME_WIDTH = 800;
const BIRD_SIZE = 30;

export const useFlappyStore = create<FlappyState>((set, get) => ({
  birdY: GAME_HEIGHT / 2,
  birdVelocity: 0,
  pipes: [],
  score: 0,
  gameOver: false,
  started: false,

  GRAVITY: BASE_GRAVITY,
  JUMP_FORCE: -6.5,
  PIPE_WIDTH: 60,
  PIPE_GAP: 150,
  PIPE_SPACING: 300,
  PIPE_SPEED: BASE_PIPE_SPEED,

  boostId: null,
  boosted: false,
  lastSubmitRanked: null,
  lastSubmitMessage: null,

  clearBoost: () => {
    set({
      boostId: null,
      boosted: false,
      PIPE_SPEED: BASE_PIPE_SPEED,
      GRAVITY: BASE_GRAVITY,
    });
  },

  startGame: (opts) => {
    if ((get() as any).gameLoop) {
      clearInterval((get() as any).gameLoop);
    }
    if ((get() as any).boostTimer) {
      clearTimeout((get() as any).boostTimer);
    }

    const boosted = Boolean(opts?.boosted);
    const boostId = opts?.boostId ?? null;

    set({
      started: true,
      gameOver: false,
      birdY: GAME_HEIGHT / 2,
      birdVelocity: 0,
      pipes: [],
      score: 0,
      boostId,
      boosted,
      lastSubmitRanked: null,
      lastSubmitMessage: null,
      PIPE_SPEED: boosted ? BASE_PIPE_SPEED / 2 : BASE_PIPE_SPEED,
      GRAVITY: boosted ? BASE_GRAVITY / 2 : BASE_GRAVITY,
    });

    const firstPipe = get().generatePipe();
    set({ pipes: [firstPipe] });

    if (boosted) {
      const boostTimer = setTimeout(() => {
        set({ PIPE_SPEED: BASE_PIPE_SPEED, GRAVITY: BASE_GRAVITY });
      }, BOOST_DURATION_MS);
      (get() as any).boostTimer = boostTimer;
    }

    const gameLoop = setInterval(() => {
      const { gameOver, started } = get();
      if (!gameOver && started) {
        get().updateGame();
      } else if (gameOver) {
        clearInterval(gameLoop);
      }
    }, 1000 / 60);

    (get() as any).gameLoop = gameLoop;
  },

  jump: () => {
    const { gameOver, started } = get();
    if (!gameOver && started) {
      set({ birdVelocity: get().JUMP_FORCE });
    }
  },

  updateGame: () => {
    const state = get();
    const { birdVelocity, GRAVITY, pipes, PIPE_SPEED, PIPE_WIDTH, birdY, score } = state;

    const newVelocity = birdVelocity + GRAVITY;
    const newBirdY = birdY + newVelocity;

    const updatedPipes = pipes
      .map((pipe) => ({
        ...pipe,
        x: pipe.x - PIPE_SPEED,
      }))
      .filter((pipe) => pipe.x + PIPE_WIDTH > 0);

    let newScore = score;
    const pipesWithScore = updatedPipes.map((pipe) => {
      if (!pipe.passed && pipe.x + PIPE_WIDTH < 100) {
        newScore = newScore + 10;
        return { ...pipe, passed: true };
      }
      return pipe;
    });

    let newPipes = pipesWithScore;
    if (
      pipesWithScore.length === 0 ||
      pipesWithScore[pipesWithScore.length - 1].x < GAME_WIDTH - get().PIPE_SPACING
    ) {
      const newPipe = get().generatePipe();
      newPipes = [...pipesWithScore, newPipe];
    }

    set({
      birdY: newBirdY,
      birdVelocity: newVelocity,
      pipes: newPipes,
      score: newScore,
    });

    const hasCollision = get().checkCollisions();
    if (hasCollision || newBirdY > GAME_HEIGHT - BIRD_SIZE || newBirdY < 0) {
      set({ gameOver: true });
      get().submitScore();
    }
  },

  resetGame: () => {
    if ((get() as any).gameLoop) {
      clearInterval((get() as any).gameLoop);
    }
    if ((get() as any).boostTimer) {
      clearTimeout((get() as any).boostTimer);
    }
    get().clearBoost();
    set({
      started: false,
      gameOver: false,
      birdY: GAME_HEIGHT / 2,
      birdVelocity: 0,
      pipes: [],
      score: 0,
      lastSubmitRanked: null,
      lastSubmitMessage: null,
    });
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
    const { score, boostId, boosted } = get();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const nickname = localStorage.getItem('nickname') || userEmail?.split('@')[0] || 'Игрок';

    if (!userId || !userEmail) return;

    try {
      const body: Record<string, unknown> = {
        user_id: userId,
        game_id: 'flappy',
        level: 1,
        score: score,
        user_email: userEmail,
        nickname: nickname,
        seed: `flappy_seed_${Date.now()}`,
        moves: [],
      };
      if (boosted && boostId) {
        body.boost_id = boostId;
      }

      const response = await api.post('/game/submit', body);

        if (response.status >= 200 && response.status < 300) {
        const ranked =
          response.data?.ranked === true &&
          !boosted &&
          !String(response.data?.message || '').includes('not ranked');
        set({
          lastSubmitRanked: ranked,
          lastSubmitMessage: response.data?.message || null,
        });
        console.log(`✅ Flappy score submitted: ${score} ranked=${ranked}`);

        const storageKey = `gameScores_${userId}`;
        const totalScoreKey = `totalScore_${userId}`;
        const playedKey = `flappyGamesPlayed_${userId}`;

        const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
        const currentBest = savedScores.flappy || 0;

        if (ranked && score > currentBest) {
          savedScores.flappy = score;
          localStorage.setItem(storageKey, JSON.stringify(savedScores));
        }

        const played = parseInt(localStorage.getItem(playedKey) || '0');
        localStorage.setItem(playedKey, String(played + 1));

        if (ranked) {
          const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0');
          localStorage.setItem(totalScoreKey, String(totalScore + score));
        }
      }
    } catch (err) {
      console.error('Failed to submit flappy score:', err);
    }
  },
}));
