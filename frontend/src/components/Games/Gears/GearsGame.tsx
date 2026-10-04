// frontend/src/components/Games/Gears/GearsGame.tsx
// Merge-up: drop levelled orbs (1→8). Two same levels merge into next. Goal: make an 8.
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';

type Orb = { id: number; x: number; y: number; vx: number; vy: number; level: number; settled: boolean };

const W = 320;
const H = 420;
const MAX_LEVEL = 8;
const RADIUS = [0, 14, 18, 22, 26, 30, 34, 38, 44];
const COLORS = [
  '',
  '#6b8cae',
  '#5a9e8a',
  '#c4a35a',
  '#c47a5a',
  '#a66bb5',
  '#5a8fc4',
  '#d4c05a',
  '#e8e8f0',
];

let idSeq = 1;

function spawnLevel(): number {
  // Bias small; never spawn 7/8 (no scam)
  const r = Math.random();
  if (r < 0.45) return 1;
  if (r < 0.75) return 2;
  if (r < 0.92) return 3;
  return 4;
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
  const scoreRef = useRef(0);

  const reset = () => {
    orbsRef.current = [];
    nextRef.current = spawnLevel();
    setNextPreview(nextRef.current);
    scoreRef.current = 0;
    setScore(0);
    setWon(false);
    setLost(false);
    droppingRef.current = false;
  };

  const drop = useCallback(() => {
    if (droppingRef.current || won || lost) return;
    const level = nextRef.current;
    const r = RADIUS[level];
    orbsRef.current.push({
      id: idSeq++,
      x: Math.min(W - r, Math.max(r, aimXRef.current)),
      y: r + 4,
      vx: 0,
      vy: 0.4,
      level,
      settled: false,
    });
    nextRef.current = spawnLevel();
    setNextPreview(nextRef.current);
    droppingRef.current = true;
    setTimeout(() => {
      droppingRef.current = false;
    }, 350);
  }, [won, lost]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    let raf = 0;
    const dangerY = 56;

    const tick = () => {
      const orbs = orbsRef.current;
      // physics
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
      // collisions + merges
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
            if (a.level === b.level && a.level < MAX_LEVEL) {
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

      // lose: settled orb above danger line
      if (
        !won &&
        orbsRef.current.some((o) => o.settled && o.y - RADIUS[o.level] < dangerY)
      ) {
        setLost(true);
      }

      // draw
      ctx.clearRect(0, 0, W, H);
      ctx.fillStyle = 'rgba(12,14,22,0.95)';
      ctx.fillRect(0, 0, W, H);
      ctx.strokeStyle = 'rgba(255,255,255,0.12)';
      ctx.beginPath();
      ctx.moveTo(0, dangerY);
      ctx.lineTo(W, dangerY);
      ctx.stroke();
      // aim ghost
      if (!droppingRef.current && !lost && !won) {
        const lvl = nextRef.current;
        const r = RADIUS[lvl];
        ctx.globalAlpha = 0.35;
        ctx.beginPath();
        ctx.arc(aimXRef.current, r + 8, r, 0, Math.PI * 2);
        ctx.fillStyle = COLORS[lvl];
        ctx.fill();
        ctx.globalAlpha = 1;
      }
      for (const o of orbsRef.current) {
        const r = RADIUS[o.level];
        ctx.beginPath();
        ctx.arc(o.x, o.y, r, 0, Math.PI * 2);
        ctx.fillStyle = COLORS[o.level];
        ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.25)';
        ctx.stroke();
        ctx.fillStyle = '#0c0e16';
        ctx.font = 'bold 12px ui-monospace, monospace';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(String(o.level), o.x, o.y);
      }
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [won, lost]);

  const onPointer = (clientX: number, rect: DOMRect) => {
    const x = ((clientX - rect.left) / rect.width) * W;
    aimXRef.current = Math.min(W - 20, Math.max(20, x));
  };

  const submit = async () => {
    setSaving(true);
    try {
      await api.post('/game/submit', {
        user_id: localStorage.getItem('userId'),
        game_id: 'gears',
        level: 1,
        score: scoreRef.current,
        user_email: localStorage.getItem('userEmail'),
        nickname: localStorage.getItem('nickname') || '',
        seed: `gears_${Date.now()}`,
        moves: [],
      });
      setNotif({ message: 'Счёт сохранён', type: 'success' });
      void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
    } catch {
      setNotif({ message: 'Не удалось сохранить', type: 'error' });
    } finally {
      setSaving(false);
    }
  };

  return (
    <GameShell
      title="Орбиты"
      onBack={() => navigate('/')}
      width="narrow"
      stats={
        <>
          <ScoreChip label="Счёт" value={score} />
          <ScoreChip label="След." value={nextPreview} />
        </>
      }
      controls={
        <>
          <Button size="sm" variant="ghost" onClick={reset}>
            Заново
          </Button>
          <Button size="sm" onClick={() => void submit()} disabled={saving || score === 0}>
            Сохранить
          </Button>
        </>
      }
      help={
        <p>
          Кликай / тапай — уронить шестерёнку. Две одного уровня сливаются в большую (до 8). Не
          пересекай линию сверху. Большие уровни почти не выпадают — без скама.
        </p>
      }
    >
      <canvas
        ref={canvasRef}
        width={W}
        height={H}
        className="max-h-full w-full max-w-sm touch-none rounded-md border border-white/10"
        onPointerMove={(e) => onPointer(e.clientX, e.currentTarget.getBoundingClientRect())}
        onPointerUp={(e) => {
          onPointer(e.clientX, e.currentTarget.getBoundingClientRect());
          drop();
        }}
      />
      <Modal open={won} onClose={() => setWon(false)} title="Собрана 8!">
        <p className="mb-4 text-text-secondary">Счёт {score}. Можно сохранить.</p>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={reset}>
            Ещё раз
          </Button>
          <Button onClick={() => void submit()} disabled={saving}>
            Сохранить
          </Button>
        </div>
      </Modal>
      <Modal open={lost} onClose={() => setLost(false)} title="Переполнение">
        <p className="mb-4 text-text-secondary">Счёт {score}.</p>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={reset}>
            Заново
          </Button>
          <Button onClick={() => void submit()} disabled={saving}>
            Сохранить
          </Button>
        </div>
      </Modal>
      {notif && (
        <Notification message={notif.message} type={notif.type} onClose={() => setNotif(null)} />
      )}
    </GameShell>
  );
}
