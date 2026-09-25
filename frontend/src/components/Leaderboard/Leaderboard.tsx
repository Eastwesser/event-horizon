// frontend/src/components/Leaderboard/Leaderboard.tsx
import { useEffect, useState, useRef } from 'react';
import { getLeaderboard } from '../../services/api';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  score: number;
}

type GameId = 'hexagon' | 'memory' | 'flappy' | 'towers' | 'hanoi';

const GAME_TABS: { id: GameId; label: string }[] = [
  { id: 'hexagon', label: '🥞 Pancaker' },
  { id: 'memory', label: '🎴 Memonia' },
  { id: 'flappy', label: '🐦 Flappy Bird' },
  { id: 'towers', label: '🗼 Builder' },
  { id: 'hanoi', label: '🪈 Hanoi' },
];

interface LeaderboardProps {
  /**
   * When set (in-game widget), show ONLY this game — no cross-game tabs.
   * Omit on global surfaces if a multi-game picker is desired.
   */
  gameId?: GameId;
}

export function Leaderboard({ gameId }: LeaderboardProps) {
  const locked = Boolean(gameId);
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedGame, setSelectedGame] = useState<GameId>(gameId ?? 'hexagon');
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (gameId) setSelectedGame(gameId);
  }, [gameId]);

  const fetchLeaderboard = async (gid: GameId = selectedGame) => {
    try {
      const { data } = await getLeaderboard(gid, 10);
      const raw = data?.entries;
      setEntries(Array.isArray(raw) ? raw.filter(Boolean) : []);
    } catch (err) {
      console.error('Failed to fetch leaderboard:', err);
      setEntries([]);
    }
  };

  useEffect(() => {
    if (!isOpen) return;

    const ws = new WebSocket(`ws://${window.location.host}/ws/leaderboard`);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (Array.isArray(data)) {
          setEntries(data.filter(Boolean));
        } else if (data.entries && Array.isArray(data.entries)) {
          setEntries(data.entries.filter(Boolean));
        } else {
          void fetchLeaderboard(selectedGame);
        }
      } catch (e) {
        console.error('Failed to parse:', e);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.close();
      }
    };
  }, [selectedGame, isOpen]);

  useEffect(() => {
    if (isOpen) void fetchLeaderboard(selectedGame);
  }, [selectedGame, isOpen]);

  const handleOpen = () => {
    setIsOpen(true);
  };

  const getMedal = (rank: number) => {
    switch (rank) {
      case 1: return '🥇';
      case 2: return '🥈';
      case 3: return '🥉';
      default: return null;
    }
  };

  const titleGame = GAME_TABS.find((t) => t.id === selectedGame)?.label ?? 'Игра';

  return (
    <>
      <Button variant="secondary" size="sm" onClick={handleOpen}>
        🏆 Топ-10
      </Button>

      <Modal
        open={isOpen}
        onClose={() => setIsOpen(false)}
        title={locked ? `Топ-10 — ${titleGame}` : 'Топ-10'}
      >
        {!locked && (
          <div className="mb-4 flex flex-wrap gap-2">
            {GAME_TABS.map((tab) => (
              <FilterChip
                key={tab.id}
                active={selectedGame === tab.id}
                onClick={() => setSelectedGame(tab.id)}
              >
                {tab.label}
              </FilterChip>
            ))}
          </div>
        )}

        {entries.length === 0 ? (
          <p className="py-8 text-center text-text-secondary">Нет данных</p>
        ) : (
          <div className="flex flex-col gap-2">
            {entries.map((entry, idx) => {
              if (!entry) return null;
              const rank = entry.rank ?? idx + 1;
              const medal = getMedal(rank);
              const score =
                typeof entry.score === 'number' && Number.isFinite(entry.score)
                  ? entry.score
                  : 0;

              return (
                <div
                  key={entry.userId || idx}
                  className="flex items-center gap-3 rounded-sm border border-white/10 bg-nebula-elevated/40 px-3 py-2"
                >
                  <div className="w-6 text-center font-hud text-sm text-text-secondary">
                    {medal || rank}
                  </div>
                  <div className="flex flex-1 items-center gap-2 overflow-hidden">
                    <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-horizon-gold to-horizon-ember text-xs font-semibold text-void">
                      {entry.user_email?.split('@')[0]?.charAt(0).toUpperCase() || '?'}
                    </div>
                    <span className="truncate text-sm text-text-primary">
                      {entry.user_email?.split('@')[0] || 'Аноним'}
                    </span>
                  </div>
                  <div className="font-hud text-sm tabular-nums text-horizon-gold">
                    {score.toLocaleString()}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </Modal>
    </>
  );
}
