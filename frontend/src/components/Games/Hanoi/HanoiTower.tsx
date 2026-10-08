// frontend/src/components/Games/Hanoi/HanoiTower.tsx
import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
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
  const [started, setStarted] = useState(false);
  const [startTime, setStartTime] = useState(0);
  const [elapsedMs, setElapsedMs] = useState(0);
  const [autoSolving, setAutoSolving] = useState(false);
  const [animating, setAnimating] = useState<{ disk: number; from: number; to: number } | null>(null);
  const [scoreSaved, setScoreSaved] = useState(false);
  const [lastRanked, setLastRanked] = useState<boolean | null>(null);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [hoverPeg, setHoverPeg] = useState<number | null>(null);
  const [boardShake, setBoardShake] = useState(false);
  const [boardPulse, setBoardPulse] = useState(false);
  const [runBoosted, setRunBoosted] = useState(false);
  const [runBoostId, setRunBoostId] = useState<string | null>(null);
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('hanoi');

  const pegRefs = useRef<(HTMLDivElement | null)[]>([]);
  const autoTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const minMoves = useMemo(() => Math.pow(2, diskCount) - 1, [diskCount]);

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  const handleBack = () => navigate('/#games');

  useEffect(() => {
    if (boostError) {
      setSaveMessage({ type: 'error', text: boostError });
      setBoostError(null);
    }
  }, [boostError, setBoostError]);

  useEffect(() => {
    if (!started || won) return;
    const id = setInterval(() => setElapsedMs(performance.now() - startTime), 100);
    return () => clearInterval(id);
  }, [startTime, won, started]);

  useEffect(() => {
    if (!started || pegs[2].length !== diskCount || diskCount <= 0 || animating) return;
    setWon(true);
    setElapsedMs(performance.now() - startTime);
  }, [pegs, diskCount, animating, startTime, started]);

  useEffect(() => {
    return () => {
      if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    };
  }, []);

  const beginRun = useCallback(async (count = diskCount) => {
    if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    setAutoSolving(false);
    setAnimating(null);
    try {
      const { boostId, boosted } = await armBoost();
      setRunBoosted(boosted);
      setRunBoostId(boostId);
    } catch {
      return;
    }
    setPegs(createInitialPegs(count));
    setSelectedPeg(null);
    setHoverPeg(null);
    setMoves(0);
    setWon(false);
    setScoreSaved(false);
    setLastRanked(null);
    setStarted(true);
    setStartTime(performance.now());
    setElapsedMs(0);
  }, [diskCount, armBoost]);

  const handleDiskCountChange = (count: number) => {
    // Pre-game only: pick rings, then «Старт» starts the timer.
    if (started) return;
    setDiskCount(count);
    if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    setAutoSolving(false);
    setAnimating(null);
    setPegs(createInitialPegs(count));
    setSelectedPeg(null);
    setHoverPeg(null);
    setMoves(0);
    setWon(false);
    setElapsedMs(0);
    setStartTime(0);
  };

  const stopAutoSolve = () => {
    if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    autoTimerRef.current = null;
    setAutoSolving(false);
    setAnimating(null);
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

  const triggerInvalidFeedback = () => {
    setBoardShake(true);
    setBoardPulse(true);
    window.setTimeout(() => setBoardShake(false), 420);
    window.setTimeout(() => setBoardPulse(false), 180);
  };

  const handlePegClick = async (pegIndex: number) => {
    if (!started || autoSolving || animating || won) return;

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
      triggerInvalidFeedback();
      if (pegs[pegIndex].length > 0) setSelectedPeg(pegIndex);
      else setSelectedPeg(null);
    }
  };

  const pegTargetClass = (pegIndex: number): string => {
    if (selectedPeg === null || selectedPeg === pegIndex || autoSolving || animating || won) {
      return '';
    }
    if (hoverPeg !== pegIndex) return '';
    return canMove(pegs, selectedPeg, pegIndex)
      ? 'hanoi-peg--target-valid'
      : 'hanoi-peg--target-invalid';
  };

  const runAutoSolve = async () => {
    if (autoSolving || won || !runBoosted) return;
    if (autoTimerRef.current) clearTimeout(autoTimerRef.current);
    setAnimating(null);
    setPegs(createInitialPegs(diskCount));
    setSelectedPeg(null);
    setHoverPeg(null);
    setMoves(0);
    setWon(false);
    setStartTime(performance.now());
    setElapsedMs(0);
    setAutoSolving(true);

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
      const body: Record<string, unknown> = {
        user_id: userId,
        game_id: 'hanoi',
        level: diskCount,
        score,
        user_email: userEmail,
        seed: `hanoi_${diskCount}_${Date.now()}`,
        moves: [],
      };
      if (runBoosted && runBoostId) body.boost_id = runBoostId;

      const response = await api.post('/game/submit', body);
      const ranked =
        response.data?.ranked === true &&
        !runBoosted &&
        !String(response.data?.message || '').includes('not ranked');
      setScoreSaved(true);
      setLastRanked(ranked);
      setSaveMessage({
        type: 'success',
        text: ranked ? 'Счёт сохранён · рекорд в лидерборд' : boostUnrankedToast(),
      });
      if (ranked) {
        void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
      }
    } catch {
      setSaveMessage({ type: 'error', text: 'Ошибка при сохранении' });
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
      onBack={handleBack}
      width="narrow"
      actions={<Balance />}
      stats={
        <>
          <ScoreChip label="Время" value={formatTime(elapsedMs)} />
          <ScoreChip label="Ходы" value={moves} className={movesToneClass} />
          <ScoreChip label="Минимум" value={minMoves} />
        </>
      }
      controls={
        <>
          {!started && (
            <label className="flex items-center gap-2 text-sm text-text-secondary">
              Колец:
              <select
                value={diskCount}
                disabled={boostBusy}
                onChange={(e) => handleDiskCountChange(Number(e.target.value))}
                className="eh-hanoi-disk-select rounded-sm border border-white/15 bg-void px-3 py-1.5 text-text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-horizon-gold"
              >
                {DISK_OPTIONS.map((n) => (
                  <option key={n} value={n} className="bg-void text-text-primary">
                    {n}
                  </option>
                ))}
              </select>
            </label>
          )}
          {started && (
            <span className="rounded-sm border border-white/10 px-3 py-1.5 text-sm text-text-secondary">
              Колец: <span className="text-text-primary">{diskCount}</span>
            </span>
          )}
          <Button
            variant="primary"
            size="sm"
            onClick={() => void beginRun()}
            disabled={autoSolving || boostBusy}
          >
            {boostBusy ? 'Старт…' : started ? 'Сброс' : 'Старт'}
          </Button>
          {runBoosted && started && !autoSolving && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => void runAutoSolve()}
              disabled={won}
            >
              Авто-решение
            </Button>
          )}
          {autoSolving && (
            <Button variant="secondary" size="sm" onClick={stopAutoSolve}>
              Стоп
            </Button>
          )}
          {!started && (
            <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} />
          )}
        </>
      }
      help={
        <>
          <p>Выберите число колец (3–8), затем нажмите «Старт» — таймер пойдёт с этой кнопки.</p>
          <p>Перенесите все кольца со стержня A на стержень C</p>
          <p>Нельзя класть большее кольцо на меньшее</p>
          {boostHelpLines('hanoi').map((line) => (
            <p key={line}>{line}</p>
          ))}
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

      {!started && (
        <p className="mb-3 text-center text-sm text-text-secondary">
          Выберите число колец и нажмите «Старт» — таймер запустится с этой кнопки.
        </p>
      )}
      <div
        className={[
          'hanoi-board',
          !started ? 'hanoi-board--pregame' : '',
          boardShake ? 'hanoi-board--shake' : '',
          boardPulse ? 'hanoi-board--pulse' : '',
        ]
          .filter(Boolean)
          .join(' ')}
      >
        {pegs.map((stack, pegIndex) => (
          <div
            key={pegIndex}
            ref={(el) => { pegRefs.current[pegIndex] = el; }}
            className={[
              'hanoi-peg',
              selectedPeg === pegIndex ? 'hanoi-peg--selected' : '',
              pegTargetClass(pegIndex),
              autoSolving ? 'hanoi-peg--disabled' : '',
            ]
              .filter(Boolean)
              .join(' ')}
            onClick={() => handlePegClick(pegIndex)}
            onMouseEnter={() => setHoverPeg(pegIndex)}
            onMouseLeave={() => setHoverPeg((h) => (h === pegIndex ? null : h))}
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

      <Modal open={won} onClose={() => setWon(false)} title="Победа">
        <p className="text-text-secondary">Все кольца на месте!</p>

        <div className="mt-4 space-y-1.5 text-text-secondary">
          <p>
            Время: <strong className="text-text-primary">{formatTime(elapsedMs)}</strong>
          </p>
          <p>
            {pluralMoves(moves)}
            {minMoves > 0 ? (
              <span className="text-text-muted"> (минимум {minMoves})</span>
            ) : null}
          </p>
          <p className={`text-sm font-medium ${moves <= minMoves ? 'text-success' : 'text-warning'}`}>
            {moves <= minMoves
              ? 'Идеально! Вы уложились в оптимум.'
              : `Превышение на ${pluralMoves(moves - minMoves)}`}
          </p>
        </div>

        <div className="mt-5 text-center">
          <p className="text-xs font-medium uppercase tracking-wide text-text-muted">Очки</p>
          <p className="font-hud text-3xl font-bold text-horizon-gold">{finalScore}</p>
        </div>

        {runBoosted && (
          <p className="mt-3 text-sm text-horizon-gold">{boostUnrankedToast()}</p>
        )}
        {!runBoosted && lastRanked === true && (
          <p className="mt-3 inline-flex items-center gap-2 text-sm text-success">
            <Icon name="check" className="h-4 w-4" aria-hidden />
            Счёт сохранён · рекорд в лидерборд
          </p>
        )}
        <GameOverActions
          onNewGame={() => void resetGame()}
          onHome={handleBack}
          onSave={() => {
            if (!scoreSaved) void handleSubmitScore();
          }}
          saveLabel={scoreSaved ? 'Сохранено' : 'Сохранить рекорд'}
          busy={boostBusy}
        />
      </Modal>
    </GameShell>
  );
}
