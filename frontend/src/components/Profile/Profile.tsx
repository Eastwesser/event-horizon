// frontend/src/components/Profile/Profile.tsx
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { StatCard } from '../ui/StatCard';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';
import { Icon, IconLabel, type IconName } from '../ui/Icon';
import { gameIcon } from '../../lib/gameIcons';
import { LAMP_HINT, TICKET_HINT } from '../../lib/currencyGlossary';
import {
  achievementIcon,
  syncAchievements,
  type ProfileAchievement,
} from '../../lib/achievements';

const GAME_ROUTES: Record<string, string> = {
  hexagon: '/game/hexagon',
  memory: '/game/memory',
  flappy: '/game/flappy',
  towers: '/game/towers',
  hanoi: '/game/hanoi',
  twenty48: '/game/twenty48',
  gears: '/game/gears',
  companion: '/game/companion',
};

const EMPTY_SCORES = {
  hexagon: 0,
  memory: 0,
  flappy: 0,
  towers: 0,
  hanoi: 0,
  twenty48: 0,
  gears: 0,
  companion: 0,
};

export function Profile() {
  const navigate = useNavigate();
  const email = localStorage.getItem('userEmail') || 'unknown@example.com';
  const userId = localStorage.getItem('userId') || '';
  const nicknameKey = `nickname_${userId}`;

  const [stats, setStats] = useState({
    nickname: localStorage.getItem(nicknameKey) || email.split('@')[0],
    totalScore: 0,
    bestScores: { ...EMPTY_SCORES },
    achievements: [] as ProfileAchievement[],
  });
  const [balance, setBalance] = useState({ lamps: 0, tickets: 0 });
  const [showResetModal, setShowResetModal] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token) return;

    // Silent seed of seen-set (option a) + load badges / scores from API
    syncAchievements({ toast: false })
      .then(({ profile, achievements }) => {
        const best = profile.best_scores || {};
        setStats((prev) => ({
          ...prev,
          nickname:
            profile.nickname ||
            localStorage.getItem(nicknameKey) ||
            email.split('@')[0],
          totalScore: profile.total_score ?? 0,
          bestScores: {
            hexagon: best.hexagon || 0,
            memory: best.memory || 0,
            flappy: best.flappy || 0,
            towers: best.towers || 0,
            hanoi: best.hanoi || 0,
            twenty48: best.twenty48 || 0,
            gears: best.gears || 0,
            companion: best.companion || 0,
          },
          achievements,
        }));
        if (profile.nickname) {
          localStorage.setItem(nicknameKey, profile.nickname);
        }
      })
      .catch(() => {
        /* keep local nickname fallback */
      });

    api
      .get('/billing/balance/all', { headers: { Authorization: `Bearer ${token}` } })
      .then((res) => {
        const data = res.data || {};
        setBalance({
          lamps: data.lamps ?? data.Lamps ?? 0,
          tickets: data.tickets ?? data.Tickets ?? 0,
        });
      })
      .catch(() => {
        /* ignore */
      });
  }, [nicknameKey, email]);

  const handleSetNickname = async () => {
    const newNick = window.prompt('Новый никнейм:', stats.nickname);
    if (!newNick || !newNick.trim()) return;

    try {
      const token = localStorage.getItem('accessToken');
      await api.post(
        '/auth/update-nickname',
        { nickname: newNick.trim() },
        { headers: { Authorization: `Bearer ${token}` } },
      );

      localStorage.setItem(nicknameKey, newNick.trim());
      setStats((prev) => ({ ...prev, nickname: newNick.trim() }));
      alert('Никнейм обновлён!');
    } catch (err) {
      alert('Ошибка при обновлении ника');
      console.error(err);
    }
  };

  const confirmResetStats = () => {
    localStorage.removeItem(`gameScores_${userId}`);
    localStorage.removeItem(`gamesPlayed_${userId}`);
    localStorage.removeItem(`totalScore_${userId}`);
    localStorage.removeItem(nicknameKey);
    setShowResetModal(false);
    window.location.reload();
  };

  const handleBack = () => {
    navigate('/');
  };

  const bestScoreRows = [
    { key: 'hexagon', label: 'Pancaker', value: stats.bestScores.hexagon },
    { key: 'memory', label: 'Memonia', value: stats.bestScores.memory },
    { key: 'flappy', label: 'Flappy Bird', value: stats.bestScores.flappy },
    { key: 'towers', label: 'Builder', value: stats.bestScores.towers },
    { key: 'hanoi', label: 'Hanoi', value: stats.bestScores.hanoi },
    { key: 'twenty48', label: '2048', value: stats.bestScores.twenty48 },
    { key: 'gears', label: 'Орбиты', value: stats.bestScores.gears },
    { key: 'companion', label: 'Компаньон', value: stats.bestScores.companion },
  ];

  const avatarIcon: IconName = stats.achievements.length >= 2 ? 'crown' : 'hex';

  return (
    <PageShell width="narrow">
      <PageHeader title="Профиль" onBack={handleBack} backLabel="На главную" />

      <div className="eh-ring mb-8 flex flex-wrap items-center gap-4 rounded-lg border border-white/10 bg-nebula p-6">
        <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-horizon-gold to-horizon-ember text-void shadow-glow-gold">
          <Icon name={avatarIcon} className="h-8 w-8" />
        </div>
        <div className="min-w-0">
          <button
            type="button"
            onClick={handleSetNickname}
            className="inline-flex items-center gap-1.5 font-display text-xl font-semibold text-text-primary transition-colors hover:text-indigo-soft"
          >
            {stats.nickname}
            <Icon name="pen" className="h-4 w-4 text-text-muted" />
          </button>
          <p className="text-sm text-text-secondary">{email}</p>
          {userId ? (
            <details className="mt-1">
              <summary className="cursor-pointer text-xs text-text-muted hover:text-text-secondary">
                Показать ID
              </summary>
              <p className="mt-1 break-all font-hud text-xs text-text-muted">{userId}</p>
            </details>
          ) : null}
        </div>
      </div>

      <div className="mb-8 grid grid-cols-2 gap-5 sm:grid-cols-3">
        <StatCard
          size="md"
          value={stats.totalScore}
          label={<IconLabel name="trophy">Всего очков</IconLabel>}
        />
        {bestScoreRows.map((row) => (
          <StatCard
            key={row.key}
            size="md"
            value={row.value}
            label={<IconLabel name={gameIcon(row.key)}>{row.label}</IconLabel>}
          />
        ))}
      </div>

      {stats.achievements.length > 0 && (
        <div className="mb-8">
          <h3 className="mb-3 font-display text-lg font-semibold text-text-primary">
            <IconLabel name="medal" iconClassName="h-5 w-5 text-horizon-gold">
              Достижения
            </IconLabel>
          </h3>
          <div className="flex flex-wrap gap-2">
            {stats.achievements.map((ach) => (
              <Badge key={ach.code} tone="gold" title={ach.description || undefined}>
                <IconLabel name={achievementIcon(ach.icon)} iconClassName="h-3 w-3">
                  {ach.title}
                </IconLabel>
              </Badge>
            ))}
          </div>
        </div>
      )}

      <div className="mb-8 flex flex-wrap gap-3">
        <span
          className="inline-flex items-center gap-1.5 rounded-sm border border-horizon-gold/30 bg-horizon-gold/10 px-3 py-1.5 font-hud text-sm tabular-nums text-horizon-gold"
          title={LAMP_HINT}
          aria-label={`${LAMP_HINT}: ${balance.lamps}`}
        >
          <Icon name="lamp" className="h-4 w-4" />
          {balance.lamps}
        </span>
        <span
          className="inline-flex items-center gap-1.5 rounded-sm border border-photon-cyan/30 bg-photon-cyan/10 px-3 py-1.5 font-hud text-sm tabular-nums text-photon-cyan"
          title={TICKET_HINT}
          aria-label={`${TICKET_HINT}: ${balance.tickets}`}
        >
          <Icon name="ticket" className="h-4 w-4" />
          {balance.tickets}
        </span>
      </div>

      <div className="mb-8">
        <h3 className="mb-3 font-display text-lg font-semibold text-text-primary">
          <IconLabel name="chart" iconClassName="h-5 w-5">
            Рекорды по играм
          </IconLabel>
        </h3>
        <div className="divide-y divide-white/10 rounded-md border border-white/10 bg-nebula">
          {bestScoreRows.map((row) => (
            <button
              key={row.key}
              type="button"
              onClick={() => navigate(GAME_ROUTES[row.key])}
              className="flex w-full items-center justify-between px-4 py-3 text-left transition-colors hover:bg-white/5"
            >
              <IconLabel name={gameIcon(row.key)} className="text-text-secondary">
                {row.label}
              </IconLabel>
              <span className="font-hud tabular-nums text-horizon-gold">{row.value}</span>
            </button>
          ))}
        </div>
      </div>

      <Button variant="ghost" onClick={() => setShowResetModal(true)}>
        Сбросить статистику
      </Button>

      <Modal
        open={showResetModal}
        onClose={() => setShowResetModal(false)}
        title="Сбросить статистику?"
      >
        <p className="text-sm leading-relaxed text-text-secondary">
          Будут удалены локальные рекорды, счётчики игр и никнейм на этом устройстве. Это действие
          нельзя отменить.
        </p>
        <div className="mt-6 flex flex-wrap justify-end gap-3">
          <Button variant="ghost" onClick={() => setShowResetModal(false)}>
            Отмена
          </Button>
          <Button variant="danger" onClick={confirmResetStats}>
            Да, сбросить
          </Button>
        </div>
      </Modal>
    </PageShell>
  );
}
