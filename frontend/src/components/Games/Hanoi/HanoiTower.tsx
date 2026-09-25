// frontend/src/components/Games/Hanoi/HanoiTower.tsx
import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';
import './HanoiTower.css';

type Pegs = [number[], number[], number[]];

/** Board geometry always assumes this many rings (rod height + width scale). */
const MAX_DISKS = 8;

const RING_COLORS = [
  '#ff6b6b',
  '#ffa94d',
  '#ffd43b',
  '#69db7c',
  '#4dabf7',
  '#9775fa',
  '#f783ac',
  '#20c997',
];

const PEG_LABELS = ['A', 'B', 'C'];

const DISK_OPTIONS = [3, 4, 5, 6, 7, 8] as const;

/** Absolute 1..MAX_DISKS width as % of peg.
 *  Peg has a solid min-width so rings stay chunky (not "beads") under GameShell. */
function ringWidthPercent(size: number): number {
  const t = Math.min(Math.max(size, 1), MAX_DISKS) / MAX_DISKS;
  // size 1 ≈ 42%, size 8 ≈ 92% of peg — matches previous comfortable proportions
  return 42 + t * 50;
}

function ringWidthCss(size: number): string {
  return `${ringWidthPercent(size)}%`;
}

/** Pixel width for the floating ring — % of viewport stretches rings during drag. */
function ringWidthPx(size: number, pegWidth: number): number {
  return (pegWidth * ringWidthPercent(size)) / 100;
}

function pluralMoves(n: number): string {
  const abs = Math.abs(n) % 100;
  const d = abs % 10;
  if (abs > 10 && abs < 20) return `${n} ходов`;
  if (d === 1) return `${n} ход`;
  if (d >= 2 && d <= 4) return `${n} хода`;
  return `${n} ходов`;
}

function createInitialPegs(diskCount: number): Pegs {
  const disks = Array.from({ length: diskCount }, (_, i) => diskCount - i);
  return [disks, [], []];
}

function canMove(pegs: Pegs, from: number, to: number): boolean {
  if (from === to) return false;
  const fromStack = pegs[from];
  const toStack = pegs[to];
  if (fromStack.length === 0) return false;
  const disk = fromStack[fromStack.length - 1];
  if (toStack.length === 0) return true;
  return disk < toStack[toStack.length - 1];
}

/** Рекурсивный алгоритм Ханоя — возвращает список ходов [from, to] */
function solveHanoi(n: number, from: number, to: number, aux: number): [number, number][] {
  if (n <= 0) return [];
  return [
    ...solveHanoi(n - 1, from, aux, to),
    [from, to],
    ...solveHanoi(n - 1, aux, to, from),
  ];
}

/** Очки: 1000 - (лишние ходы × 20), минимум 100 — та же формула, что у Мемонии. */
function calculateScore(moves: number, minMoves: number): number {
  const excess = Math.max(0, moves - minMoves);
  return Math.max(100, 1000 - excess * 20);
}

export function HanoiTower() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');

  const [diskCount, setDiskCount] = useState(5);
  const [pegs, setPegs] = useState<Pegs>(() => createInitialPegs(5));
  const [selectedPeg, setSelectedPeg] = useState<number | null>(null);
  const [moves, setMoves] = useState(0);
  const [won, setWon] = useState(false);
  const [startTime, setStartTime] = useState<number>(() => performance.now());
  const [elapsedMs, setElapsedMs] = useState(0);
  const [autoSolving, setAutoSolving] = useState(false);
  const [animating, setAnimating] = useState<{ disk: number; from: number; to: number } | null>(null);
  const [scoreSaved, setScoreSaved] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const pegRefs = useRef<(HTMLDivElement | null)[]>([]);
  const autoTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const minMoves = useMemo(() => Math.pow(2, diskCount) - 1, [diskCount]);

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  useEffect(() => {
    if (won) return;
    const id = setInterval(() => setElapsedMs(performance.now() - startTime), 100);
    return () => clearInterval(id);
  }, [startTime, won]);

  useEffect(() => {
    if (pegs[2].length === diskCount && diskCount > 0 && !animating) {
      setWon(true);
      setElapsedMs(performance.now() - startTime);
    }
  }, [pegs, diskCount, animating, startTime]);

  useEffect(() => {
    return () => {
      if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    };
  }, []);

  const resetGame = useCallback((count = diskCount) => {
    if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    setAutoSolving(false);
    setAnimating(null);
    setPegs(createInitialPegs(count));
    setSelectedPeg(null);
    setMoves(0);
    setWon(false);
    setScoreSaved(false);
    setStartTime(performance.now());
    setElapsedMs(0);
  }, [diskCount]);

  const handleDiskCountChange = (count: number) => {
    setDiskCount(count);
    resetGame(count);
  };

  const applyMove = useCallback((from: number, to: number, countMove = true) => {
    setPegs((prev) => {
      const next: Pegs = [ [...prev[0]], [...prev[1]], [...prev[2]] ];
      const disk = next[from].pop()!;
      next[to].push(disk);
      return next;
    });
    if (countMove) setMoves((m) => m + 1);
    setSelectedPeg(null);
  }, []);

  const animateAndMove = useCallback(
    (from: number, to: number, countMove = true): Promise<void> => {
      return new Promise((resolve) => {
        if (!canMove(pegs, from, to)) {
          resolve();
          return;
        }
        const disk = pegs[from][pegs[from].length - 1];
        setAnimating({ disk, from, to });
        setTimeout(() => {
          applyMove(from, to, countMove);
          setAnimating(null);
          resolve();
        }, 250);
      });
    },
    [pegs, applyMove]
  );

  const handlePegClick = async (pegIndex: number) => {
    if (autoSolving || animating || won) return;

    if (selectedPeg === null) {
      if (pegs[pegIndex].length > 0) setSelectedPeg(pegIndex);
      return;
    }

    if (selectedPeg === pegIndex) {
      setSelectedPeg(null);
      return;
    }

    if (canMove(pegs, selectedPeg, pegIndex)) {
      await animateAndMove(selectedPeg, pegIndex);
    } else {
      if (pegs[pegIndex].length > 0) setSelectedPeg(pegIndex);
      else setSelectedPeg(null);
    }
  };

  const runAutoSolve = async () => {
    if (autoSolving || won) return;
    resetGame();
    setAutoSolving(true);
    setSelectedPeg(null);

    const solution = solveHanoi(diskCount, 0, 2, 1);

    const runStep = (index: number) => {
      if (index >= solution.length) {
        setAutoSolving(false);
        return;
      }
      const [from, to] = solution[index];
      setPegs((prev) => {
        if (!canMove(prev, from, to)) return prev;
        const next: Pegs = [ [...prev[0]], [...prev[1]], [...prev[2]] ];
        const disk = next[from].pop()!;
        next[to].push(disk);
        return next;
      });
      setMoves((m) => m + 1);
      autoTimerRef.current = setTimeout(() => runStep(index + 1), 500);
    };

    autoTimerRef.current = setTimeout(() => runStep(0), 500);
  };

  const handleSubmitScore = async () => {
    const score = calculateScore(moves, minMoves);
    try {
      const userId = localStorage.getItem('userId');
      const userEmail = localStorage.getItem('userEmail');

      await api.post('/game/submit', {
        user_id: userId,
        game_id: 'hanoi',
        level: diskCount,
        score,
        user_email: userEmail,
        seed: `hanoi_${diskCount}_${Date.now()}`,
        moves: [],
      });

      setScoreSaved(true);
      setSaveMessage({ type: 'success', text: '✅ Рекорд сохранён!' });
    } catch (err) {
      setSaveMessage({ type: 'error', text: '❌ Ошибка при сохранении' });
    }
  };

  const formatTime = (ms: number) => {
    const sec = Math.floor(ms / 1000);
    const min = Math.floor(sec / 60);
    const s = sec % 60;
    const cs = Math.floor((ms % 1000) / 10);
    return `${min}:${String(s).padStart(2, '0')}.${String(cs).padStart(2, '0')}`;
  };

  const getFloatingStyle = (): CSSProperties | undefined => {
    if (!animating) return undefined;
    const toEl = pegRefs.current[animating.to];
    if (!toEl) return undefined;
    const toRect = toEl.getBoundingClientRect();
    return {
      width: `${ringWidthPx(animating.disk, toRect.width)}px`,
      left: toRect.left + toRect.width / 2,
      top: toRect.top + 40,
      transform: 'translateX(-50%)',
      background: RING_COLORS[(animating.disk - 1) % RING_COLORS.length],
    };
  };

  const initialFloatingStyle = (): CSSProperties | undefined => {
    if (!animating) return undefined;
    const fromEl = pegRefs.current[animating.from];
    if (!fromEl) return undefined;
    const fromRect = fromEl.getBoundingClientRect();
    return {
      width: `${ringWidthPx(animating.disk, fromRect.width)}px`,
      left: fromRect.left + fromRect.width / 2,
      top: fromRect.top + 40,
      transform: 'translateX(-50%)',
      background: RING_COLORS[(animating.disk - 1) % RING_COLORS.length],
    };
  };

  const [floatStyle, setFloatStyle] = useState<CSSProperties | undefined>();

  useEffect(() => {
    if (!animating) {
      setFloatStyle(undefined);
      return;
    }
    setFloatStyle(initialFloatingStyle());
    requestAnimationFrame(() => {
      requestAnimationFrame(() => setFloatStyle(getFloatingStyle()));
    });
  }, [animating]);

  const movesToneClass =
    moves > 0 && moves <= minMoves
      ? 'border-success/30 [&_span:last-child]:text-success'
      : moves > minMoves
        ? 'border-warning/30 [&_span:last-child]:text-warning'
        : '';
  const finalScore = calculateScore(moves, minMoves);

  return (
    <GameShell
      title="Hanoi"
      onBack={() => navigate('/')}
      width="narrow"
      stats={
        <>
          <ScoreChip label="Время" value={formatTime(elapsedMs)} />
          <ScoreChip label="Ходы" value={moves} className={movesToneClass} />
          <ScoreChip label="Минимум" value={minMoves} />
        </>
      }
      controls={
        <>
          <label className="flex items-center gap-2 text-sm text-text-secondary">
            Колец:
            <select
              value={diskCount}
              disabled={autoSolving || moves > 0}
              onChange={(e) => handleDiskCountChange(Number(e.target.value))}
              className="rounded-sm border border-horizon-gold/30 bg-black/30 px-3 py-1.5 text-text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-photon-cyan"
            >
              {DISK_OPTIONS.map((n) => (
                <option key={n} value={n}>{n}</option>
              ))}
            </select>
          </label>
          <Button variant="secondary" size="sm" onClick={() => resetGame()} disabled={autoSolving}>
            🔄 Сброс
          </Button>
          <Button variant="ghost" size="sm" onClick={runAutoSolve} disabled={autoSolving || won}>
            {autoSolving ? '⏳ Решаю...' : '🤖 Авто-решение'}
          </Button>
        </>
      }
      help={
        <>
          <p>🎯 Перенесите все кольца со стержня A на стержень C</p>
          <p>🚫 Нельзя класть большее кольцо на меньшее</p>
          <p>🏆 Очки: 1000 − (лишние ходы × 20), минимум 100</p>
          <p>🤖 Кнопка «Авто-решение» покажет оптимальный путь</p>
        </>
      }
    >
      {saveMessage && (
        <Notification
          type={saveMessage.type}
          message={saveMessage.text}
          onClose={() => setSaveMessage(null)}
        />
      )}

      <div className="hanoi-board">
        {pegs.map((stack, pegIndex) => (
          <div
            key={pegIndex}
            ref={(el) => { pegRefs.current[pegIndex] = el; }}
            className={`hanoi-peg ${selectedPeg === pegIndex ? 'hanoi-peg--selected' : ''} ${autoSolving ? 'hanoi-peg--disabled' : ''}`}
            onClick={() => handlePegClick(pegIndex)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => e.key === 'Enter' && handlePegClick(pegIndex)}
          >
            <span className="hanoi-peg-label">Стержень {PEG_LABELS[pegIndex]}</span>
            <div className="hanoi-rod" />
            <div className="hanoi-base" />
            <div className="hanoi-peg-stack">
              {stack.map((disk, i) => {
                const isTop = i === stack.length - 1;
                const isMoving =
                  animating &&
                  animating.from === pegIndex &&
                  isTop &&
                  animating.disk === disk;
                return (
                  <div
                    key={`${pegIndex}-${disk}-${i}`}
                    className={`hanoi-ring ${isTop && selectedPeg === pegIndex ? 'hanoi-ring--top-selected' : ''} ${isMoving ? 'hanoi-ring--hidden' : ''}`}
                    style={{
                      width: ringWidthCss(disk),
                      background: RING_COLORS[(disk - 1) % RING_COLORS.length],
                    }}
                  >
                    {disk}
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      {animating && floatStyle && (
        <div className="hanoi-floating-ring" style={floatStyle}>
          {animating.disk}
        </div>
      )}

      <Modal open={won} onClose={() => setWon(false)} title="🎉 Победа!">
        <p className="text-text-secondary">Все кольца на месте!</p>

        <div className="mt-4 space-y-1.5 text-text-secondary">
          <p>
            ⏱ Время: <strong className="text-text-primary">{formatTime(elapsedMs)}</strong>
          </p>
          <p>
            🎯 {pluralMoves(moves)}
            {minMoves > 0 ? (
              <span className="text-text-muted"> (минимум {minMoves})</span>
            ) : null}
          </p>
          <p className={`text-sm font-medium ${moves <= minMoves ? 'text-success' : 'text-warning'}`}>
            {moves <= minMoves
              ? '✨ Идеально! Вы уложились в оптимум!'
              : `📈 Превышение на ${pluralMoves(moves - minMoves)}`}
          </p>
        </div>

        <div className="mt-5 text-center">
          <p className="text-xs font-medium uppercase tracking-wide text-text-muted">Очки</p>
          <p className="font-hud text-3xl font-bold text-horizon-gold">{finalScore}</p>
        </div>

        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <Button
            variant="primary"
            className="w-full min-w-0"
            onClick={handleSubmitScore}
            disabled={scoreSaved}
          >
            {scoreSaved ? '✅ Сохранено' : '📤 Сохранить рекорд'}
          </Button>
          <Button variant="ghost" className="w-full min-w-0" onClick={() => resetGame()}>
            🔄 Играть снова
          </Button>
        </div>
      </Modal>
    </GameShell>
  );
}
