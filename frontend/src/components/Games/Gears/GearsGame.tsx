// frontend/src/components/Games/Gears/GearsGame.tsx
// Merge-up: drop levelled gears (1→MAX). Two same levels merge into next.
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import { getNickname } from '../../../lib/nickname';
import { useGameBoost } from '../../../hooks/useGameBoost';
import { Balance } from '../../Billing/Balance';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';
import { Icon } from '../../ui/Icon';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';
import { GameOverActions } from '../GameOverActions';
import { boostHelpLines } from '../../../lib/gameBoostCopy';

type Orb = { id: number; x: number; y: number; vx: number; vy: number; level: number; settled: boolean };

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

function drawGear(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  color: string,
  level: number,
  alpha = 1,
  flower = false,
) {
  ctx.save();
  ctx.globalAlpha = alpha;
  if (flower) {
    // Petal blob for sunflower/rose/pansy skins only — metal stays a real gear.
    const petals = 6 + (level % 3);
    for (let i = 0; i < petals; i++) {
      const ang = (i / petals) * Math.PI * 2 - Math.PI / 2;
      const px = x + Math.cos(ang) * r * 0.55;
      const py = y + Math.sin(ang) * r * 0.55;
      ctx.beginPath();
      ctx.ellipse(px, py, r * 0.42, r * 0.28, ang, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.fill();
    }
    ctx.beginPath();
    ctx.arc(x, y, r * 0.38, 0, Math.PI * 2);
    ctx.fillStyle = '#3a2a10';
    ctx.fill();
  } else {
    // Classic spur gear: more teeth + deeper valleys so it doesn't read as a star/prize.
    const teeth = 10 + Math.min(level, 6);
    const tip = r;
    const valley = r * 0.68;
    const root = r * 0.62;
    ctx.beginPath();
    for (let i = 0; i < teeth; i++) {
      const a0 = (i / teeth) * Math.PI * 2 - Math.PI / 2;
      const a1 = ((i + 0.35) / teeth) * Math.PI * 2 - Math.PI / 2;
      const a2 = ((i + 0.5) / teeth) * Math.PI * 2 - Math.PI / 2;
      const a3 = ((i + 0.85) / teeth) * Math.PI * 2 - Math.PI / 2;
      const pts: [number, number][] = [
        [x + Math.cos(a0) * valley, y + Math.sin(a0) * valley],
        [x + Math.cos(a1) * tip, y + Math.sin(a1) * tip],
        [x + Math.cos(a2) * tip, y + Math.sin(a2) * tip],
        [x + Math.cos(a3) * valley, y + Math.sin(a3) * valley],
        [x + Math.cos(((i + 1) / teeth) * Math.PI * 2 - Math.PI / 2) * root, y + Math.sin(((i + 1) / teeth) * Math.PI * 2 - Math.PI / 2) * root],
      ];
      for (let p = 0; p < pts.length; p++) {
        if (i === 0 && p === 0) ctx.moveTo(pts[p][0], pts[p][1]);
        else ctx.lineTo(pts[p][0], pts[p][1]);
      }
    }
    ctx.closePath();
    ctx.fillStyle = color;
    ctx.fill();
    ctx.strokeStyle = 'rgba(255,255,255,0.35)';
    ctx.lineWidth = 1.25;
    ctx.stroke();
    ctx.beginPath();
    ctx.arc(x, y, r * 0.36, 0, Math.PI * 2);
    ctx.fillStyle = '#12141c';
    ctx.fill();
    ctx.beginPath();
    ctx.arc(x, y, r * 0.14, 0, Math.PI * 2);
    ctx.fillStyle = '#8a90a0';
    ctx.fill();
  }
  ctx.fillStyle = '#e8e8f0';
  ctx.font = 'bold 11px ui-monospace, monospace';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(String(level), x, y);
  ctx.restore();
}

export function GearsGame() {
  const navigate = useNavigate();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const orbsRef = useRef<Orb[]>([]);
  const nextRef = useRef(spawnLevel());
  const aimXRef = useRef(W / 2);
  const droppingRef = useRef(false);
  const [score, setScore] = useState(0);
  const [nextPreview, setNextPreview] = useState(nextRef.current);
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

  useEffect(() => {
    localStorage.setItem(SKIN_KEY, skin);
  }, [skin]);

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
      nextRef.current = spawnLevel();
      setNextPreview(nextRef.current);
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
    const level = nextRef.current;
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
    });
    nextRef.current = boostedRef.current ? level : spawnLevel();
    setNextPreview(nextRef.current);
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
              add.push({
                id: idSeq++,
                x: (a.x + b.x) / 2,
                y: (a.y + b.y) / 2,
                vx: 0,
                vy: -1.2,
                level: nl,
                settled: false,
              });
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
      ctx.fillStyle = 'rgba(12,14,22,0.95)';
      ctx.fillRect(0, 0, W, H);
      ctx.strokeStyle = 'rgba(255,80,80,0.35)';
      ctx.setLineDash([4, 4]);
      ctx.beginPath();
      ctx.moveTo(0, dangerY);
      ctx.lineTo(W, dangerY);
      ctx.stroke();
      ctx.setLineDash([]);
      if (!droppingRef.current && !lost && !won) {
        const lvl = nextRef.current;
        const r = RADIUS[lvl];
        const COLORS = colorsRef.current;
        drawGear(ctx, aimXRef.current, r + 8, r, COLORS[lvl], lvl, 0.35, skin !== 'metal');
      }
      for (const o of orbsRef.current) {
        const COLORS = colorsRef.current;
        drawGear(ctx, o.x, o.y, RADIUS[o.level], COLORS[o.level], o.level, 1, skin !== 'metal');
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

  return (
    <GameShell
      title="Gears"
      onBack={handleBack}
      width="narrow"
      actions={<Balance />}
      stats={
        <>
          <ScoreChip label="Счёт" value={score} />
          <ScoreChip label="След." value={nextPreview} />
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
      <div className="mb-3 flex flex-wrap justify-center gap-2">
        {(
          [
            ['metal', 'Металл'],
            ['sunflower', 'Подсолнух'],
            ['rose', 'Роза'],
            ['pansy', 'Анютины'],
          ] as const
        ).map(([id, label]) => (
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
      <canvas
        ref={canvasRef}
        width={W}
        height={H}
        className="mx-auto max-w-full touch-none rounded-md border border-white/10"
        onPointerMove={(e) => onPointer(e.clientX, e.currentTarget.getBoundingClientRect())}
        onClick={() => drop()}
      />
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
