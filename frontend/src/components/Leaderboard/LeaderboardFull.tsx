// frontend/src/components/Leaderboard/LeaderboardFull.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { getLeaderboard } from '../../services/api';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';
import { cn } from '../../lib/cn';
import { Icon, IconLabel } from '../ui/Icon';
import { GAME_LB_TABS, type GameTabId } from '../../lib/gameIcons';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  nickname: string;
  score: number;
  game_id?: string;
}

type GameId = GameTabId;

const GAME_TABS = GAME_LB_TABS;

const RANK_TONE: Record<number, string> = {
  1: 'text-horizon-gold',
  2: 'text-text-primary',
  3: 'text-horizon-ember',
};

function isGameId(v: string | null): v is GameId {
  return !!v && GAME_TABS.some((t) => t.id === v);
}

export function LeaderboardFull() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const highlightRef = useRef<HTMLLIElement | null>(null);

  const paramGame = searchParams.get('game');
  const paramLevel = Number(searchParams.get('level') || '1');
  const highlightUser = searchParams.get('highlight') || searchParams.get('user_id') || '';

  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedGame, setSelectedGame] = useState<GameId>(
    isGameId(paramGame) ? paramGame : 'hexagon',
  );
  const [selectedLevel, setSelectedLevel] = useState(
    Number.isFinite(paramLevel) && paramLevel >= 1 ? paramLevel : 1,
  );

  useEffect(() => {
    if (isGameId(paramGame)) setSelectedGame(paramGame);
    if (Number.isFinite(paramLevel) && paramLevel >= 1) setSelectedLevel(paramLevel);
  }, [paramGame, paramLevel]);

  useEffect(() => {
    setLoading(true);
    // Product: single board per game (level=1). Multi-level Flappy UI is commented out.
    getLeaderboard(selectedGame, 50, 1)
      .then(({ data }) => {
        const raw = Array.isArray(data?.entries) ? data.entries : [];
        setEntries(
          raw.map((e: Partial<LeaderboardEntry>, i: number) => ({
            rank: e.rank ?? i + 1,
            userId: e.userId || (e as { user_id?: string }).user_id || `anon-${i}`,
            user_email: e.user_email || '',
            nickname: e.nickname || '',
            score: typeof e.score === 'number' ? e.score : 0,
            game_id: e.game_id,
          })),
        );
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to fetch leaderboard:', err);
        setLoading(false);
      });
  }, [selectedGame, selectedLevel]);

  useEffect(() => {
    if (!highlightUser || loading) return;
    highlightRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }, [highlightUser, loading, entries]);

  const activeIcon = GAME_TABS.find((t) => t.id === selectedGame)?.icon || 'trophy';
  const handleBack = () => navigate('/');

  return (
    <PageShell width="narrow">
      <PageHeader
        title={
          <IconLabel name="trophy" iconClassName="h-7 w-7 text-horizon-gold">
            Лидерборд
          </IconLabel>
        }
        onBack={handleBack}
        backLabel="На главную"
      />

      <div className="mb-6 flex flex-wrap gap-2">
        {GAME_TABS.map((tab) => (
          <FilterChip
            key={tab.id}
            active={selectedGame === tab.id}
            onClick={() => setSelectedGame(tab.id)}
          >
            <IconLabel name={tab.icon} iconClassName="h-4 w-4">
              {tab.label}
            </IconLabel>
          </FilterChip>
        ))}
      </div>

      {/* Flappy levels 2–3 unused in product — keep single global board (level=1).
      {selectedGame === 'flappy' && (
        <div className="mb-4 flex flex-wrap gap-2">
          {[1, 2, 3].map((lvl) => (
            <FilterChip
              key={lvl}
              active={selectedLevel === lvl}
              onClick={() => setSelectedLevel(lvl)}
            >
              Уровень {lvl}
            </FilterChip>
          ))}
        </div>
      )}
      */}

      {loading ? (
        <div className="flex justify-center py-16">
          <Spinner size={48} />
        </div>
      ) : entries.length === 0 ? (
        <p className="py-12 text-center text-text-secondary">Пока нет результатов</p>
      ) : (
        <ul className="space-y-2">
          {entries.map((entry) => {
            const highlighted = highlightUser && entry.userId === highlightUser;
            return (
              <li
                key={`${entry.userId}-${entry.rank}`}
                ref={highlighted ? highlightRef : undefined}
                className={cn(
                  'flex items-center gap-3 rounded-md border px-4 py-3',
                  highlighted
                    ? 'border-horizon-gold/60 bg-horizon-gold/10'
                    : 'border-white/10 bg-nebula',
                )}
              >
                <span
                  className={cn(
                    'w-8 font-hud text-lg font-semibold',
                    RANK_TONE[entry.rank] || 'text-text-muted',
                  )}
                >
                  {entry.rank}
                </span>
                <Icon name={activeIcon} className="h-5 w-5 text-text-muted" />
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium text-text-primary">
                    {entry.nickname || entry.user_email || entry.userId.slice(0, 8)}
                  </p>
                </div>
                <span className="font-hud text-horizon-gold">{entry.score}</span>
              </li>
            );
          })}
        </ul>
      )}
    </PageShell>
  );
}
