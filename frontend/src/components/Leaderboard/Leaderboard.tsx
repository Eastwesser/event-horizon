// frontend/src/components/Leaderboard/Leaderboard.tsx
import { useEffect, useState, useRef } from 'react';
import { getLeaderboard } from '../../services/api';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';
import { Icon, IconLabel } from '../ui/Icon';
import { GAME_LB_TABS, type GameTabId } from '../../lib/gameIcons';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  score: number;
}

type GameId = GameTabId;

const GAME_TABS = GAME_LB_TABS;

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
  const [selectedLevel, setSelectedLevel] = useState(1);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (gameId) setSelectedGame(gameId);
  }, [gameId]);

  const fetchLeaderboard = async (gid: GameId = selectedGame, level = selectedLevel) => {
    try {
      const lv = gid === 'flappy' ? level : 1;
      const { data } = await getLeaderboard(gid, 10, lv);
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
    if (isOpen) void fetchLeaderboard(selectedGame, selectedLevel);
  }, [selectedGame, selectedLevel, isOpen]);

  const handleOpen = () => {
    setIsOpen(true);
  };

  const titleGame = GAME_TABS.find((t) => t.id === selectedGame)?.label ?? 'Игра';

  return (
    <>
      <Button variant="secondary" size="sm" onClick={handleOpen}>
        <IconLabel name="trophy" iconClassName="h-3.5 w-3.5">
          Топ-10
        </IconLabel>
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
                <IconLabel name={tab.icon}>{tab.label}</IconLabel>
              </FilterChip>
            ))}
          </div>
        )}

        {selectedGame === 'flappy' && (
          <label className="mb-3 flex items-center gap-2 text-sm text-text-primary">
            Уровень
            <select
              className="rounded-sm border border-white/15 bg-void px-2 py-1"
              value={selectedLevel}
              onChange={(e) =>
                setSelectedLevel(Math.min(10, Math.max(1, parseInt(e.target.value, 10) || 1)))
              }
            >
              {Array.from({ length: 10 }, (_, i) => i + 1).map((lv) => (
                <option key={lv} value={lv}>
                  {lv}
                </option>
              ))}
            </select>
          </label>
        )}

        {entries.length === 0 ? (
          <p className="py-8 text-center text-text-secondary">Нет данных</p>
        ) : (
          <div className="flex flex-col gap-2">
            {entries.map((entry, idx) => {
              if (!entry) return null;
              const rank = entry.rank ?? idx + 1;
              const score =
                typeof entry.score === 'number' && Number.isFinite(entry.score)
                  ? entry.score
                  : 0;

              return (
                <div
                  key={entry.userId || idx}
                  className="flex items-center gap-3 rounded-sm border border-white/10 bg-nebula-elevated/40 px-3 py-2"
                >
                  <div className="flex w-8 items-center justify-center gap-0.5 font-hud text-sm text-text-secondary">
                    {rank <= 3 ? <Icon name="medal" className="h-3.5 w-3.5 text-horizon-gold" /> : null}
                    {rank}
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
