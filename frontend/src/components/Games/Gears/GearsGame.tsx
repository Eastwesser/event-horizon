// frontend/src/components/Games/Gears/GearsGame.tsx
// Merge-up: drop levelled gears (1→MAX). Two same levels merge into next.
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import { getNickname } from '../../../lib/nickname';
import { useGameBoost } from '../../../hooks/useGameBoost';
import { useSkins } from '../../../hooks/useSkins';
import { Balance } from '../../Billing/Balance';
import { Spinner } from '../../ui/Spinner';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';
import { Icon } from '../../ui/Icon';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';
import { GameOverActions } from '../GameOverActions';
import { boostHelpLines } from '../../../lib/gameBoostCopy';
import { drawAim, drawGear, drawMergeFlash, drawWell } from './gearsDraw';
import './gears.css';

type Orb = {
  id: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  level: number;
  settled: boolean;
  rot: number;
};

type Flash = { x: number; y: number; r: number; life: number };

const QUEUE_LEN = 5;
const FLASH_LIFE = 16;

const W = 320;
const H = 420;
const MAX_LEVEL = 10;
/** Top 10% is the danger / spawn band. */
const DANGER_Y = Math.round(H * 0.1);
const RADIUS = [0, 14, 17, 20, 23, 26, 29, 32, 36, 40, 44];
const COLOR_PALETTES: Record<string, string[]> = {
  metal: [
    '',
    '#6b8cae',
    '#5a9e8a',
    '#c4a35a',
    '#c47a5a',
    '#a66bb5',
    '#5a8fc4',
    '#d4c05a',
    '#c9c9d8',
    '#e0c070',
    '#f2f2f8',
  ],
  sunflower: [
    '',
    '#f4d35e',
    '#eea73b',
    '#e07a2f',
    '#c45c26',
    '#8b5a2b',
    '#6b8f3c',
    '#f7e8a4',
    '#d4a017',
    '#ffe066',
    '#fff4c2',
  ],
  rose: [
    '',
    '#f2a6b8',
    '#e86b8a',
    '#c93a5a',
    '#9b2745',
    '#6b1f3a',
    '#d48aa8',
    '#f5c6d0',
    '#b83b5e',
    '#ff8fab',
    '#ffe0e9',
  ],
  pansy: [
    '',
    '#b388ff',
    '#7c4dff',
    '#651fff',
    '#ffd54f',
    '#ffb300',
    '#ce93d8',
    '#5e35b1',
    '#ea80fc',
    '#fff59d',
    '#f3e5f5',
  ],
};

type FlowerSkin = keyof typeof COLOR_PALETTES;
const SKIN_KEY = 'eh_gears_skin_v1';

let idSeq = 1;

function spawnLevel(): number {
  const r = Math.random();
  if (r < 0.42) return 1;
  if (r < 0.72) return 2;
  if (r < 0.9) return 3;
  return 4;
}

function makeQueue(): number[] {
  return Array.from({ length: QUEUE_LEN }, () => spawnLevel());
}

function TrayGear({
  level,
  color,
  size,
  flower,
}: {
  level: number;
  color: string;
  size: number;
  flower: boolean;
}) {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.clearRect(0, 0, size, size);
    drawGear(ctx, size / 2, size / 2, size / 2 - 2, color, level, 1, flower, 0);
  }, [level, color, size, flower]);
  return <canvas ref={ref} width={size} height={size} aria-hidden />;
}

export function GearsGame() {
  const navigate = useNavigate();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const orbsRef = useRef<Orb[]>([]);
  const queueRef = useRef<number[]>(makeQueue());
  const flashesRef = useRef<Flash[]>([]);
  const aimXRef = useRef(W / 2);
  const droppingRef = useRef(false);
  const [score, setScore] = useState(0);
  const [queue, setQueue] = useState<number[]>(() => queueRef.current);
  const [won, setWon] = useState(false);
  const [lost, setLost] = useState(false);
  const [notif, setNotif] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
  const [saving, setSaving] = useState(false);
  const [lastRanked, setLastRanked] = useState<boolean | null>(null);
  const scoreRef = useRef(0);
  const boostedRef = useRef(false);
  const [runBoosted, setRunBoosted] = useState(false);
  const [runBoostId, setRunBoostId] = useState<string | null>(null);
  const [skin, setSkin] = useState<FlowerSkin>(() => {
    const raw = localStorage.getItem(SKIN_KEY) as FlowerSkin | null;
    return raw && COLOR_PALETTES[raw] ? raw : 'metal';
  });
  const colorsRef = useRef(COLOR_PALETTES.metal);
  colorsRef.current = COLOR_PALETTES[skin] || COLOR_PALETTES.metal;
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('gears');
  const { skins, loading: skinsLoading } = useSkins();
  const ownedFlower = {
    sunflower: skins.gears.hasSunflower,
    rose: skins.gears.hasRose,
    pansy: skins.gears.hasPansy,
  };
  const hasFlowerMerch = ownedFlower.sunflower || ownedFlower.rose || ownedFlower.pansy;

  useEffect(() => {
    localStorage.setItem(SKIN_KEY, skin);
  }, [skin]);

  useEffect(() => {
    if (skinsLoading) return;
    if (skin !== 'metal' && !ownedFlower[skin as 'sunflower' | 'rose' | 'pansy']) {
      setSkin('metal');
    }
  }, [skinsLoading, skin, ownedFlower.sunflower, ownedFlower.rose, ownedFlower.pansy]);

  useEffect(() => {
    if (boostError) {
      setNotif({ message: boostError, type: 'error' });
      setBoostError(null);
    }
  }, [boostError, setBoostError]);

  const reset = async () => {
    try {
      const { boostId, boosted } = await armBoost();
      setRunBoosted(boosted);
      setRunBoostId(boostId);
      boostedRef.current = boosted;
      orbsRef.current = [];
      queueRef.current = makeQueue();
      setQueue([...queueRef.current]);
      flashesRef.current = [];
      scoreRef.current = 0;
      setScore(0);
      setWon(false);
      setLost(false);
      setLastRanked(null);
      droppingRef.current = false;
    } catch {
      /* handled */
    }
  };

  const drop = useCallback(() => {
    if (won || lost || droppingRef.current) return;
    const level = queueRef.current[0];
    const r = RADIUS[level];
    const jitter = (Math.random() - 0.5) * 8;
    orbsRef.current.push({
      id: idSeq++,
      x: Math.min(W - r, Math.max(r, aimXRef.current + jitter)),
      y: r + 6,
      vx: 0,
      vy: 0.4,
      level,
      settled: false,
      rot: 0,
    });
    if (!boostedRef.current) {
      queueRef.current = [...queueRef.current.slice(1), spawnLevel()];
    }
    setQueue([...queueRef.current]);
    droppingRef.current = true;
    setTimeout(() => {
      droppingRef.current = false;
    }, 480);
  }, [won, lost]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    let raf = 0;
    const dangerY = DANGER_Y;

    const tick = () => {
      const orbs = orbsRef.current;
      for (const o of orbs) {
        o.vy += 0.18;
        o.x += o.vx;
        o.y += o.vy;
        o.rot += o.vx * 0.05;
        const r = RADIUS[o.level];
        if (o.x < r) {
          o.x = r;
          o.vx *= -0.4;
        }
        if (o.x > W - r) {
          o.x = W - r;
          o.vx *= -0.4;
        }
        if (o.y > H - r) {
          o.y = H - r;
          o.vy *= -0.25;
          o.vx *= 0.92;
          if (Math.abs(o.vy) < 0.35) {
            o.vy = 0;
            o.settled = true;
          }
        }
      }

      const remove = new Set<number>();
      const add: Orb[] = [];
      for (let i = 0; i < orbs.length; i++) {
        for (let j = i + 1; j < orbs.length; j++) {
          const a = orbs[i];
          const b = orbs[j];
          if (remove.has(a.id) || remove.has(b.id)) continue;
          const dx = b.x - a.x;
          const dy = b.y - a.y;
          const dist = Math.hypot(dx, dy) || 0.001;
          const min = RADIUS[a.level] + RADIUS[b.level];
          if (dist < min) {
            const aSlow = Math.hypot(a.vx, a.vy) < 1.35;
            const bSlow = Math.hypot(b.vx, b.vy) < 1.35;
            // Merge on contact (same level) — no multi-touch delay; slow gate only avoids mid-air chain abuse.
            if (a.level === b.level && a.level < MAX_LEVEL && (aSlow || bSlow || a.settled || b.settled)) {
              remove.add(a.id);
              remove.add(b.id);
              const nl = a.level + 1;
              const mx = (a.x + b.x) / 2;
              const my = (a.y + b.y) / 2;
              add.push({
                id: idSeq++,
                x: mx,
                y: my,
                vx: 0,
                vy: -1.2,
                level: nl,
                settled: false,
                rot: 0,
              });
              flashesRef.current.push({ x: mx, y: my, r: RADIUS[nl], life: FLASH_LIFE });
              scoreRef.current += nl * 10;
              setScore(scoreRef.current);
              if (nl >= MAX_LEVEL) setWon(true);
            } else {
              const overlap = (min - dist) / 2;
              const nx = dx / dist;
              const ny = dy / dist;
              a.x -= nx * overlap;
              a.y -= ny * overlap;
              b.x += nx * overlap;
              b.y += ny * overlap;
              const dvx = b.vx - a.vx;
              const dvy = b.vy - a.vy;
              const impact = dvx * nx + dvy * ny;
              if (impact < 0) {
                a.vx += nx * impact;
                a.vy += ny * impact;
                b.vx -= nx * impact;
                b.vy -= ny * impact;
              }
            }
          }
        }
      }
      if (remove.size) {
        orbsRef.current = orbs.filter((o) => !remove.has(o.id)).concat(add);
      } else if (add.length) {
        orbsRef.current = orbs.concat(add);
      }

      if (
        !won &&
        orbsRef.current.some((o) => {
          const top = o.y - RADIUS[o.level];
          return o.settled && top < dangerY;
        })
      ) {
        setLost(true);
      }

      ctx.clearRect(0, 0, W, H);
      drawWell(ctx, W, H, dangerY);
      const flower = skin !== 'metal';
      const COLORS = colorsRef.current;
      if (!droppingRef.current && !lost && !won) {
        const lvl = queueRef.current[0];
        const r = RADIUS[lvl];
        const gx = aimXRef.current;
        const gy = r + 8;
        drawAim(ctx, gx, gy + r, H - 10);
        drawGear(ctx, gx, gy, r, COLORS[lvl], lvl, 0.4, flower, 0);
      }
      for (const o of orbsRef.current) {
        drawGear(ctx, o.x, o.y, RADIUS[o.level], COLORS[o.level], o.level, 1, flower, o.rot);
      }
      flashesRef.current = flashesRef.current
        .map((f) => ({ ...f, life: f.life - 1 }))
        .filter((f) => f.life > 0);
      for (const f of flashesRef.current) {
        drawMergeFlash(ctx, f.x, f.y, f.r, f.life / FLASH_LIFE);
      }
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [won, lost, skin]);

  const onPointer = (clientX: number, rect: DOMRect) => {
    const x = ((clientX - rect.left) / rect.width) * W;
    aimXRef.current = Math.min(W - 20, Math.max(20, x));
  };

  const submit = async () => {
    setSaving(true);
    try {
      const body: Record<string, unknown> = {
        user_id: localStorage.getItem('userId'),
        game_id: 'gears',
        level: 1,
        score: scoreRef.current,
        user_email: localStorage.getItem('userEmail'),
        nickname: getNickname(),
        seed: `gears_${Date.now()}`,
        moves: [],
      };
      if (runBoosted && runBoostId) body.boost_id = runBoostId;
      const response = await api.post('/game/submit', body);
      const ranked =
        response.data?.ranked === true &&
        !runBoosted &&
        !String(response.data?.message || '').includes('not ranked');
      setLastRanked(ranked);
      setNotif({
        message: ranked ? 'Счёт сохранён · рекорд в лидерборд' : boostUnrankedToast(),
        type: 'success',
      });
      if (ranked) {
        void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
      }
    } catch {
      setNotif({ message: 'Не удалось сохранить', type: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const handleBack = () => navigate('/#games');

  if (skinsLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-void">
        <Spinner size={56} />
      </div>
    );
  }

  return (
    <GameShell
      title="Gears"
      onBack={handleBack}
      width="narrow"
      actions={<Balance />}
      stats={
        <>
          <ScoreChip label="Счёт" value={score} />
          <ScoreChip label="След." value={queue[0]} />
          <ScoreChip label="Цель" value={MAX_LEVEL} />
        </>
      }
      controls={
        <>
          <Button size="sm" variant="ghost" onClick={() => void reset()} disabled={boostBusy}>
            Заново
          </Button>
          <Button size="sm" onClick={() => void submit()} disabled={saving || score === 0}>
            Сохранить
          </Button>
          <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} />
        </>
      }
      help={
        <>
          <p>
            Соединяй шестерёнки одного уровня — они сливаются в следующий. Цель: собрать самую
            большую (уровень {MAX_LEVEL}).
          </p>
          <p>Не заполняй поле выше красной линии — иначе переполнение.</p>
          {boostHelpLines('gears').map((line) => (
            <p key={line}>{line}</p>
          ))}
        </>
      }
    >
      <div className="gears-stage">
        {hasFlowerMerch && (
          <div className="gears-skins">
            {(
              [
                ['metal', 'Металл', true],
                ['sunflower', 'Подсолнух', ownedFlower.sunflower],
                ['rose', 'Роза', ownedFlower.rose],
                ['pansy', 'Анютины глазки', ownedFlower.pansy],
              ] as const
            )
              .filter(([, , show]) => show)
              .map(([id, label]) => (
                <Button
                  key={id}
                  size="sm"
                  variant={skin === id ? 'secondary' : 'ghost'}
                  onClick={() => setSkin(id)}
                >
                  {label}
                </Button>
              ))}
          </div>
        )}
        <div className="gears-well">
          <canvas
            ref={canvasRef}
            width={W}
            height={H}
            onPointerMove={(e) => onPointer(e.clientX, e.currentTarget.getBoundingClientRect())}
            onClick={() => drop()}
          />
        </div>
        <div className="gears-tray" aria-label="Следующие шестерни">
          {queue.map((lvl, i) => (
            <div
              key={`${i}-${lvl}`}
              className={['gears-tray-slot', i === 0 ? 'gears-tray-slot--next' : ''].filter(Boolean).join(' ')}
            >
              <TrayGear
                level={lvl}
                color={(COLOR_PALETTES[skin] || COLOR_PALETTES.metal)[lvl]}
                size={i === 0 ? 48 : 32}
                flower={skin !== 'metal'}
              />
            </div>
          ))}
        </div>
      </div>
      <Modal open={won} onClose={() => setWon(false)} title={`Собрана ${MAX_LEVEL}!`}>
        <p className="mb-4 text-text-secondary">Счёт: {score}.</p>
        {runBoosted && <p className="mb-4 text-sm text-horizon-gold">{boostUnrankedToast()}</p>}
        {!runBoosted && lastRanked === true && (
          <p className="mb-4 inline-flex items-center gap-2 text-sm text-success">
            <Icon name="check" className="h-4 w-4" aria-hidden />
            Счёт сохранён · рекорд в лидерборд
          </p>
        )}
        <GameOverActions
          onNewGame={() => void reset()}
          onHome={handleBack}
          onSave={() => void submit()}
          busy={saving || boostBusy}
        />
      </Modal>
      <Modal open={lost} onClose={() => setLost(false)} title="Переполнение">
        <p className="mb-4 text-text-secondary">Поле забито до линии спавна. Счёт: {score}.</p>
        {runBoosted && <p className="mb-4 text-sm text-horizon-gold">{boostUnrankedToast()}</p>}
        <GameOverActions
          onNewGame={() => void reset()}
          onHome={handleBack}
          onSave={() => void submit()}
          busy={saving || boostBusy}
        />
      </Modal>
      {notif && (
        <Notification message={notif.message} type={notif.type} onClose={() => setNotif(null)} />
      )}
    </GameShell>
  );
}
