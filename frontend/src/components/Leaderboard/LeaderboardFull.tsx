// frontend/src/components/Leaderboard/LeaderboardFull.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getLeaderboard } from '../../services/api';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';
import { cn } from '../../lib/cn';
import { Icon, IconLabel, type IconName } from '../ui/Icon';
import { gameIcon } from '../../lib/gameIcons';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  nickname: string;
  score: number;
  game_id?: string;
}

type GameId = 'hexagon' | 'memory' | 'flappy' | 'towers' | 'hanoi' | 'twenty48' | 'gears' | 'companion';

const GAME_TABS: { id: GameId; label: string; icon: IconName }[] = [
  { id: 'hexagon', label: 'Pancaker', icon: gameIcon('hexagon') },
  { id: 'flappy', label: 'Flappy Bird', icon: gameIcon('flappy') },
  { id: 'memory', label: 'Memonia', icon: gameIcon('memory') },
  { id: 'towers', label: 'Builder', icon: gameIcon('towers') },
  { id: 'hanoi', label: 'Hanoi', icon: gameIcon('hanoi') },
  { id: 'twenty48', label: '2048', icon: gameIcon('twenty48') },
  { id: 'gears', label: 'Орбиты', icon: gameIcon('gears') },
  { id: 'companion', label: 'Компаньон', icon: gameIcon('companion') },
];

const RANK_TONE: Record<number, string> = {
  1: 'text-horizon-gold',
  2: 'text-text-primary',
  3: 'text-horizon-ember',
};

export function LeaderboardFull() {
  const navigate = useNavigate();
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedGame, setSelectedGame] = useState<GameId>('hexagon');

  useEffect(() => {
    setLoading(true);
    getLeaderboard(selectedGame, 50)
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
  }, [selectedGame]);

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
            <IconLabel name={tab.icon}>{tab.label}</IconLabel>
          </FilterChip>
        ))}
      </div>

      {loading ? (
        <div className="flex flex-col items-center gap-4 py-20">
          <Spinner size={48} />
          <p className="text-text-secondary">Загрузка рекордов...</p>
        </div>
      ) : entries.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-20 text-center">
          <p className="text-text-secondary">Пока нет рекордов в этой игре</p>
          <p className="text-text-muted">Стань первым!</p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-md border border-white/10 bg-nebula">
          <table className="w-full">
            <thead>
              <tr className="border-b border-white/10 text-left text-xs uppercase tracking-wide text-text-muted">
                <th className="px-4 py-3 font-normal">#</th>
                <th className="px-4 py-3 font-normal">Ник</th>
                <th className="px-4 py-3 text-right font-normal">Очки</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {entries.map((entry, idx) => {
                const rank = idx + 1;
                return (
                  <tr key={entry.userId || `row-${idx}`} className="hover:bg-white/5">
                    <td
                      className={cn(
                        'px-4 py-3 font-hud tabular-nums',
                        RANK_TONE[rank] || 'text-text-secondary',
                      )}
                    >
                      {rank <= 3 ? (
                        <Icon name="medal" className="inline h-4 w-4" title={`${rank}`} />
                      ) : (
                        rank
                      )}
                      {rank <= 3 ? (
                        <span className="ml-1 align-middle">{rank}</span>
                      ) : null}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2.5">
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-horizon-gold to-horizon-ember text-sm font-semibold text-void">
                          {entry.nickname?.charAt(0).toUpperCase() ||
                            entry.user_email?.charAt(0).toUpperCase() ||
                            '?'}
                        </div>
                        <span className="truncate text-text-primary">
                          {entry.nickname || entry.user_email?.split('@')[0] || 'Аноним'}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right font-hud tabular-nums text-horizon-gold">
                      <span className="inline-flex items-center justify-end gap-1.5">
                        {(entry.score ?? 0).toLocaleString()}
                        <Icon name={activeIcon} className="h-3.5 w-3.5" />
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </PageShell>
  );
}
