// frontend/src/components/Games/Twenty48/Twenty48Game.tsx
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
  type TouchEvent,
} from 'react';
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
import { GameHelp } from '../../../lib/GameHelp';
import './twenty48.css';

type Board = number[][];

const SIZE = 4;
const SWIPE_MIN_PX = 28;
function tileClass(v: number): string {
  if (v <= 0) return 't48-tile';
  if (v >= 2048) return 't48-tile t48-tile--2048';
  return `t48-tile t48-tile--${v}`;
}

function spawnedCell(beforeSpawn: Board, after: Board): string | null {
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      if (beforeSpawn[r][c] === 0 && after[r][c] !== 0) return `${r}-${c}`;
    }
  }
  return null;
}


function emptyBoard(): Board {
  return Array.from({ length: SIZE }, () => Array(SIZE).fill(0));
}

function clone(b: Board): Board {
  return b.map((row) => [...row]);
}

function spawn(b: Board): Board {
  const empty: [number, number][] = [];
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      if (b[r][c] === 0) empty.push([r, c]);
    }
  }
  if (empty.length === 0) return b;
  const [r, c] = empty[Math.floor(Math.random() * empty.length)];
  const next = clone(b);
  next[r][c] = Math.random() < 0.9 ? 2 : 4;
  return next;
}

function slideRow(row: number[]): { row: number[]; gained: number } {
  const nums = row.filter((n) => n !== 0);
  let gained = 0;
  const merged: number[] = [];
  for (let i = 0; i < nums.length; i++) {
    if (i + 1 < nums.length && nums[i] === nums[i + 1]) {
      const v = nums[i] * 2;
      merged.push(v);
      gained += v;
      i++;
    } else {
      merged.push(nums[i]);
    }
  }
  while (merged.length < SIZE) merged.push(0);
  return { row: merged, gained };
}

function moveLeft(b: Board): { board: Board; gained: number; moved: boolean } {
  let gained = 0;
  let moved = false;
  const next = b.map((row) => {
    const res = slideRow(row);
    gained += res.gained;
    if (row.some((v, i) => v !== res.row[i])) moved = true;
    return res.row;
  });
  return { board: next, gained, moved };
}

function rotateCW(b: Board): Board {
  const n = emptyBoard();
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) n[c][SIZE - 1 - r] = b[r][c];
  }
  return n;
}

function move(b: Board, dir: 'left' | 'right' | 'up' | 'down') {
  let board = clone(b);
  let rotates = 0;
  if (dir === 'up') rotates = 3;
  if (dir === 'right') rotates = 2;
  if (dir === 'down') rotates = 1;
  for (let i = 0; i < rotates; i++) board = rotateCW(board);
  const res = moveLeft(board);
  board = res.board;
  for (let i = 0; i < (4 - rotates) % 4; i++) board = rotateCW(board);
  return { board, gained: res.gained, moved: res.moved };
}

function canMove(b: Board): boolean {
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      if (b[r][c] === 0) return true;
      if (c + 1 < SIZE && b[r][c] === b[r][c + 1]) return true;
      if (r + 1 < SIZE && b[r][c] === b[r + 1][c]) return true;
    }
  }
  return false;
}

function maxTile(b: Board): number {
  return Math.max(...b.flat());
}

export function Twenty48Game() {
  const navigate = useNavigate();
  const [board, setBoard] = useState<Board>(() => spawn(spawn(emptyBoard())));
  const [score, setScore] = useState(0);
  const [best, setBest] = useState(() => Number(localStorage.getItem('eh_twenty48_best') || 0));
  const [over, setOver] = useState(false);
  const [won, setWon] = useState(false);
  const [notif, setNotif] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
  const [saving, setSaving] = useState(false);
  const [lastRanked, setLastRanked] = useState<boolean | null>(null);
  const [runBoosted, setRunBoosted] = useState(false);
  const [runBoostId, setRunBoostId] = useState<string | null>(null);
  const [undoSnap, setUndoSnap] = useState<{ board: Board; score: number } | null>(null);
  const [undoUsed, setUndoUsed] = useState(false);
  const [spawnKey, setSpawnKey] = useState<string | null>(null);
  const touchStart = useRef<{ x: number; y: number } | null>(null);
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('twenty48');

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
      setBoard(spawn(spawn(emptyBoard())));
      setScore(0);
      setOver(false);
      setWon(false);
      setUndoSnap(null);
      setUndoUsed(false);
      setSpawnKey(null);
      setLastRanked(null);
    } catch {
      /* handled */
    }
  };

  const applyMove = useCallback(
    (dir: 'left' | 'right' | 'up' | 'down') => {
      if (over) return;
      const res = move(board, dir);
      if (!res.moved) return;
      if (runBoosted && !undoUsed) {
        setUndoSnap({ board: clone(board), score });
      }
      const next = spawn(res.board);
      setSpawnKey(spawnedCell(res.board, next));
      const nextScore = score + res.gained;
      setBoard(next);
      setScore(nextScore);
      if (nextScore > best) {
        setBest(nextScore);
        localStorage.setItem('eh_twenty48_best', String(nextScore));
      }
      if (!won && maxTile(next) >= 2048) setWon(true);
      if (!canMove(next)) setOver(true);
    },
    [board, score, best, over, won, runBoosted, undoUsed],
  );

  const undoOnce = () => {
    if (!runBoosted || undoUsed || !undoSnap) return;
    setBoard(undoSnap.board);
    setScore(undoSnap.score);
    setUndoSnap(null);
    setUndoUsed(true);
    setOver(false);
  };

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const map: Record<string, 'left' | 'right' | 'up' | 'down'> = {
        ArrowLeft: 'left',
        ArrowRight: 'right',
        ArrowUp: 'up',
        ArrowDown: 'down',
        a: 'left',
        d: 'right',
        w: 'up',
        s: 'down',
      };
      const dir = map[e.key];
      if (!dir) return;
      e.preventDefault();
      applyMove(dir);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [applyMove]);

  const finishSwipe = (x: number, y: number) => {
    const start = touchStart.current;
    touchStart.current = null;
    if (!start) return;
    const dx = x - start.x;
    const dy = y - start.y;
    if (Math.max(Math.abs(dx), Math.abs(dy)) < SWIPE_MIN_PX) return;
    if (Math.abs(dx) > Math.abs(dy)) {
      applyMove(dx > 0 ? 'right' : 'left');
    } else {
      applyMove(dy > 0 ? 'down' : 'up');
    }
  };

  const onTouchStart = (e: TouchEvent) => {
    const t = e.touches[0];
    if (!t) return;
    touchStart.current = { x: t.clientX, y: t.clientY };
  };

  const onTouchEnd = (e: TouchEvent) => {
    const t = e.changedTouches[0];
    if (!t) {
      touchStart.current = null;
      return;
    }
    finishSwipe(t.clientX, t.clientY);
  };

  const onPointerDown = (e: ReactPointerEvent) => {
    if (e.pointerType === 'touch') return; // touch handlers cover mobile
    touchStart.current = { x: e.clientX, y: e.clientY };
  };

  const onPointerUp = (e: ReactPointerEvent) => {
    if (e.pointerType === 'touch') return;
    finishSwipe(e.clientX, e.clientY);
  };

  const submit = async () => {
    setSaving(true);
    try {
      const body: Record<string, unknown> = {
        user_id: localStorage.getItem('userId'),
        game_id: 'twenty48',
        level: 1,
        score,
        user_email: localStorage.getItem('userEmail'),
        nickname: getNickname(),
        seed: `twenty48_${Date.now()}`,
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
      setNotif({ message: 'Не удалось сохранить счёт', type: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const handleBack = () => navigate('/#games');

  return (
    <GameShell
      title="2048"
      onBack={handleBack}
      gameId="twenty48"
      width="wide"
      actions={<Balance />}
      stats={
        <>
          <ScoreChip label="Счёт" value={score} />
          <ScoreChip label="Рекорд" value={best} />
        </>
      }
      controls={
        <>
          <Button size="sm" variant="ghost" onClick={() => void reset()} disabled={boostBusy}>
            Заново
          </Button>
          {runBoosted && !undoUsed && (
            <Button size="sm" variant="secondary" onClick={undoOnce} disabled={!undoSnap}>
              Undo (1)
            </Button>
          )}
          <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} gameId={'twenty48'} />
        </>
      }
      help={
        <>
          <GameHelp gameId="twenty48" />
        </>
      }
    >
      <div className="t48-stage">
        <div
          className="t48-board"
          onTouchStart={onTouchStart}
          onTouchEnd={onTouchEnd}
          onPointerDown={onPointerDown}
          onPointerUp={onPointerUp}
          onPointerLeave={() => {
            touchStart.current = null;
          }}
        >
          {board.flatMap((row, r) =>
            row.map((v, c) => {
              const id = `${r}-${c}`;
              return (
                <div
                  key={id}
                  className={`${tileClass(v)}${spawnKey === id ? ' t48-tile--spawn' : ''}`}
                >
                  {v || ''}
                </div>
              );
            }),
          )}
        </div>
      </div>
      <div className="mt-3 grid grid-cols-4 gap-2 sm:hidden">
        {(['up', 'left', 'down', 'right'] as const).map((d) => (
          <Button
            key={d}
            size="sm"
            variant="ghost"
            className="min-h-11 w-full justify-center text-lg"
            onClick={() => applyMove(d)}
          >
            {d === 'up' ? '↑' : d === 'down' ? '↓' : d === 'left' ? '←' : '→'}
          </Button>
        ))}
      </div>

      <Modal open={over} onClose={() => setOver(false)} title="Ходов больше нет">
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
          newLabel="Ещё раз"
          busy={saving || boostBusy}
        />
      </Modal>
      <Modal open={won && !over} onClose={() => setWon(false)} title="2048!">
        <p className="mb-4 text-text-secondary">Можно продолжать или сохранить счёт {score}.</p>
        <GameOverActions
          onNewGame={() => setWon(false)}
          onHome={handleBack}
          onSave={() => void submit()}
          newLabel="Играть дальше"
          busy={saving || boostBusy}
        />
      </Modal>
      {notif && (
        <Notification message={notif.message} type={notif.type} onClose={() => setNotif(null)} />
      )}
    </GameShell>
  );
}
