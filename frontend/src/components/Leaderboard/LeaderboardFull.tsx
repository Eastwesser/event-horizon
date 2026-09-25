// frontend/src/components/Leaderboard/LeaderboardFull.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getLeaderboard } from '../../services/api';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';
import { cn } from '../../lib/cn';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  nickname: string;
  score: number;
  game_id?: string;
}

type GameId = 'hexagon' | 'memory' | 'flappy' | 'towers' | 'hanoi';

const GAME_TABS: { id: GameId; label: string; icon: string }[] = [
  { id: 'hexagon', label: 'Pancaker', icon: '🥞' },
  { id: 'flappy', label: 'Flappy Bird', icon: '🐦' },
  { id: 'memory', label: 'Memonia', icon: '🎴' },
  { id: 'towers', label: 'Builder', icon: '🗼' },
  { id: 'hanoi', label: 'Hanoi', icon: '🪈' },
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
      .catch(err => {
        console.error('Failed to fetch leaderboard:', err);
        setLoading(false);
      });
  }, [selectedGame]);

  const getMedal = (rank: number) => {
    switch (rank) {
      case 1: return '🥇';
      case 2: return '🥈';
      case 3: return '🥉';
      default: return null;
    }
  };

  const activeIcon = GAME_TABS.find((t) => t.id === selectedGame)?.icon;
  const handleBack = () => navigate('/');

  return (
    <PageShell width="narrow">
      <PageHeader title="🏆 Лидерборд" onBack={handleBack} backLabel="На главную" />

      <div className="mb-6 flex flex-wrap gap-2">
        {GAME_TABS.map((tab) => (
          <FilterChip
            key={tab.id}
            active={selectedGame === tab.id}
            onClick={() => setSelectedGame(tab.id)}
          >
            {tab.icon} {tab.label}
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
          <p className="text-text-secondary">😢 Пока нет рекордов в этой игре</p>
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
                const medal = getMedal(rank);

                return (
                  <tr key={entry.userId || `row-${idx}`} className="hover:bg-white/5">
                    <td className={cn('px-4 py-3 font-hud tabular-nums', RANK_TONE[rank] || 'text-text-secondary')}>
                      {medal || rank}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2.5">
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-horizon-gold to-horizon-ember text-sm font-semibold text-void">
                          {entry.nickname?.charAt(0).toUpperCase() || entry.user_email?.charAt(0).toUpperCase() || '?'}
                        </div>
                        <span className="truncate text-text-primary">
                          {entry.nickname || entry.user_email?.split('@')[0] || 'Аноним'}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right font-hud tabular-nums text-horizon-gold">
                      {(entry.score ?? 0).toLocaleString()} {activeIcon}
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
