// frontend/src/store/towerStore.ts
import { create } from 'zustand';
import api from '../../../services/api';
import { getNickname } from '../../../lib/nickname';

export type Slab = { x: number; z: number; w: number; d: number };
export type TravelAxis = 'x' | 'z';
export type Scrap = Slab & { dy: number; life: number; lift: number; hue: number };

interface TowerState {
  tower: Slab[];
  towerHeight: number;
  mover: Slab;
  axis: TravelAxis;
  direction: 1 | -1;
  scraps: Scrap[];
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

/** Flush enough that a sliver must not show. World units, about 2 screen px. */
const PERFECT_PX = 3;
export const LAYER = 16;

function baseSlab(size: number): Slab {
  return { x: -size / 2, z: -size / 2, w: size, d: size };
}

function moverFrom(placed: Slab, axis: TravelAxis): Slab {
  if (axis === 'x') return { ...placed, x: placed.x - placed.w - 28 };
  return { ...placed, z: placed.z - placed.d - 28 };
}

function tickScraps(scraps: Scrap[]): Scrap[] {
  return scraps
    .map((s) => ({ ...s, dy: s.dy + 5, life: s.life - 1 }))
    .filter((s) => s.life > 0 && s.dy < 240);
}

export const useTowerStore = create<TowerState>((set, get) => ({
  tower: [],
  towerHeight: 0,
  mover: baseSlab(100),
  axis: 'x',
  direction: 1,
  scraps: [],
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
    const base = baseSlab(startWidthForDifficulty(difficulty));

    set({
      tower: [base],
      towerHeight: 1,
      mover: moverFrom(base, 'x'),
      axis: 'x',
      direction: 1,
      scraps: [],
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
      const { gameOver, started, scraps } = get();
      if (started && (!gameOver || scraps.length > 0)) {
        get().update();
      }
    }, 1000 / 60);

    (get() as { gameLoop?: ReturnType<typeof setInterval> }).gameLoop = gameLoop;
  },

  update: () => {
    const { mover, axis, direction, tower, BASE_SPEED, difficulty, boosted, gameOver, scraps } =
      get();
    const nextScraps = tickScraps(scraps);
    if (gameOver || tower.length === 0) {
      set({ scraps: nextScraps });
      return;
    }

    const speedMul = boosted ? 0.55 : 1;
    const speed = (BASE_SPEED * 0.85 + (difficulty - 1) * 0.35) * speedMul;
    const last = tower[tower.length - 1];
    const size = axis === 'x' ? mover.w : mover.d;
    const anchor = axis === 'x' ? last.x : last.z;
    const anchorSize = axis === 'x' ? last.w : last.d;
    const min = anchor - size - 36;
    const max = anchor + anchorSize + 36;

    let pos = (axis === 'x' ? mover.x : mover.z) + direction * speed;
    let nextDir = direction;
    if (pos <= min) {
      pos = min;
      nextDir = 1;
    } else if (pos + size >= max) {
      pos = max - size;
      nextDir = -1;
    }

    const next = { ...mover };
    if (axis === 'x') next.x = pos;
    else next.z = pos;
    set({ mover: next, direction: nextDir, scraps: nextScraps });
  },

  dropBlock: () => {
    const { mover, axis, tower, scraps, score, combo, gameOver, started, difficulty } = get();

    if (gameOver || !started || tower.length === 0) return;

    const last = tower[tower.length - 1];
    const m0 = axis === 'x' ? mover.x : mover.z;
    const span = axis === 'x' ? mover.w : mover.d;
    const m1 = m0 + span;
    const t0 = axis === 'x' ? last.x : last.z;
    const tSpan = axis === 'x' ? last.w : last.d;
    const t1 = t0 + tSpan;
    const grace = softFailPx(difficulty);

    const flush =
      Math.abs(m0 - t0) <= PERFECT_PX && Math.abs(span - tSpan) <= 1;
    if (flush) {
      const nextAxis: TravelAxis = axis === 'x' ? 'z' : 'x';
      const placed = { ...last };
      const stack = [...tower, placed];
      set({
        tower: stack,
        towerHeight: stack.length,
        mover: moverFrom(placed, nextAxis),
        axis: nextAxis,
        direction: 1,
        score: score + get().calculateBlockScore(),
        floor: stack.length,
        combo: combo + 1,
      });
      return;
    }

    const overlap = Math.min(m1, t1 + grace) - Math.max(m0, t0 - grace);
    if (overlap <= 0) {
      const miss: Scrap = {
        ...mover,
        dy: 0,
        life: 42,
        lift: tower.length * LAYER,
        hue: tower.length + 1,
      };
      set({ gameOver: true, scraps: [...scraps, miss] });
      void get().submitScore();
      return;
    }

    const real0 = Math.max(m0, t0);
    const real1 = Math.min(m1, t1);
    const real = real1 - real0;
    const kept = Math.max(real > 0 ? real : Math.min(overlap, 8), 5);
    const kept0 = real > 0 ? real0 : Math.max(t0, Math.min(m0, t1 - kept));

    const cuts: Scrap[] = [];
    const lift = tower.length * LAYER;
    const hue = tower.length + 1;
    const scrap = (piece: Slab): Scrap => ({ ...piece, dy: 0, life: 42, lift, hue });
    if (m0 < kept0 - 0.5) {
      const cut = kept0 - m0;
      cuts.push(
        scrap(
          axis === 'x'
            ? { x: m0, z: mover.z, w: cut, d: mover.d }
            : { x: mover.x, z: m0, w: mover.w, d: cut },
        ),
      );
    }
    if (m1 > kept0 + kept + 0.5) {
      const cut = m1 - (kept0 + kept);
      cuts.push(
        scrap(
          axis === 'x'
            ? { x: kept0 + kept, z: mover.z, w: cut, d: mover.d }
            : { x: mover.x, z: kept0 + kept, w: mover.w, d: cut },
        ),
      );
    }

    const placed: Slab =
      axis === 'x'
        ? { x: kept0, z: last.z, w: kept, d: last.d }
        : { x: last.x, z: kept0, w: last.w, d: kept };
    const stack = [...tower, placed];
    const nextAxis: TravelAxis = axis === 'x' ? 'z' : 'x';
    const tooThin = kept < minWidthForDifficulty(difficulty);

    set({
      tower: stack,
      towerHeight: stack.length,
      mover: moverFrom(placed, nextAxis),
      axis: nextAxis,
      direction: 1,
      scraps: [...scraps, ...cuts],
      score: score + get().calculateBlockScore(),
      floor: stack.length,
      combo: combo + 1,
      gameOver: tooThin,
    });

    if (tooThin) void get().submitScore();
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
    const nickname = getNickname() || userEmail?.split('@')[0] || 'Игрок';

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
          void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
        }
      }
    } catch (err) {
      console.error('Failed to submit towers score:', err);
    }
  },
}));
