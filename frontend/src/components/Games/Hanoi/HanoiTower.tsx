// frontend/src/components/Games/Hanoi/HanoiTower.tsx
import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from 'react';
import { useNavigate } from 'react-router-dom';
import './HanoiTower.css';

type Pegs = [number[], number[], number[]];

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

  const formatTime = (ms: number) => {
    const sec = Math.floor(ms / 1000);
    const min = Math.floor(sec / 60);
    const s = sec % 60;
    const cs = Math.floor((ms % 1000) / 10);
    return `${min}:${String(s).padStart(2, '0')}.${String(cs).padStart(2, '0')}`;
  };

  const ringWidth = (size: number) => `${30 + size * 14}%`;

  const getFloatingStyle = (): CSSProperties | undefined => {
    if (!animating) return undefined;
    const toEl = pegRefs.current[animating.to];
    if (!toEl) return undefined;
    const toRect = toEl.getBoundingClientRect();
    const width = ringWidth(animating.disk);
    return {
      width,
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
    const width = ringWidth(animating.disk);
    return {
      width,
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

  return (
    <div className="hanoi-container">
      <div className="hanoi-header">
        <div className="hanoi-title">
          <h1>🗼 Ханойская башня</h1>
          <p>Переместите все кольца на стержень C</p>
        </div>
        <div className="hanoi-stats">
          <div className="hanoi-stat">
            <span className="stat-label">Время</span>
            <span className="stat-value">{formatTime(elapsedMs)}</span>
          </div>
          <div className="hanoi-stat">
            <span className="stat-label">Ходы</span>
            <span className="stat-value">{moves}</span>
          </div>
          <div className={`hanoi-stat hanoi-stat--moves-${moves > 0 && moves <= minMoves ? 'optimal' : moves > minMoves ? 'over' : ''}`}>
            <span className="stat-label">Минимум</span>
            <span className="stat-value">{minMoves}</span>
          </div>
        </div>
      </div>

      <div className="hanoi-controls">
        <label>
          Колец:
          <select
            value={diskCount}
            disabled={autoSolving || moves > 0}
            onChange={(e) => handleDiskCountChange(Number(e.target.value))}
          >
            {[3, 4, 5, 6, 7, 8].map((n) => (
              <option key={n} value={n}>{n}</option>
            ))}
          </select>
        </label>
        <button className="hanoi-btn hanoi-btn--reset" onClick={() => resetGame()} disabled={autoSolving}>
          🔄 Сброс
        </button>
        <button className="hanoi-btn hanoi-btn--auto" onClick={runAutoSolve} disabled={autoSolving || won}>
          {autoSolving ? '⏳ Решаю...' : '🤖 Авто-решение'}
        </button>
        <button className="hanoi-btn hanoi-btn--back" onClick={() => navigate('/')}>
          ← На главную
        </button>
      </div>

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
                      width: ringWidth(disk),
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

      {won && (
        <div className="hanoi-win-overlay" onClick={() => setWon(false)}>
          <div className="hanoi-win-modal" onClick={(e) => e.stopPropagation()}>
            <h2>🎉 Победа!</h2>
            <p>Все кольца на месте!</p>
            <p>⏱ Время: <strong>{formatTime(elapsedMs)}</strong></p>
            <p>🎯 Ходов: <strong>{moves}</strong> (минимум {minMoves})</p>
            <p className={moves <= minMoves ? 'win-optimal' : 'win-over'}>
              {moves <= minMoves
                ? '✨ Идеально! Вы уложились в оптимум!'
                : `📈 Превышение на ${moves - minMoves} ход(ов)`}
            </p>
            <button className="hanoi-btn hanoi-btn--reset" onClick={() => resetGame()}>
              🔄 Играть снова
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
