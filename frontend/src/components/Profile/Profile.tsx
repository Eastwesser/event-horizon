// frontend/src/components/Profile/Profile.tsx
import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { StatCard } from '../ui/StatCard';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';

const GAME_ROUTES: Record<string, string> = {
  hexagon: '/game/hexagon',
  memory: '/game/memory',
  flappy: '/game/flappy',
  towers: '/game/towers',
  hanoi: '/game/hanoi',
};

export function Profile() {
  const navigate = useNavigate();
  const email = localStorage.getItem('userEmail') || 'unknown@example.com';
  const userId = localStorage.getItem('userId') || '';

  const storageKey = `gameScores_${userId}`;
  const playedKey = `gamesPlayed_${userId}`;
  const totalScoreKey = `totalScore_${userId}`;
  const nicknameKey = `nickname_${userId}`;

  const [stats, setStats] = useState({
    nickname: localStorage.getItem(nicknameKey) || email.split('@')[0],
    totalScore: 0,
    bestScores: { hexagon: 0, memory: 0, flappy: 0, towers: 0, hanoi: 0 },
    gamesPlayed: { hexagon: 0, memory: 0, flappy: 0, towers: 0, hanoi: 0 },
    achievements: [] as string[]
  });
  const [balance, setBalance] = useState({ lamps: 0, tickets: 0 });
  const [showResetModal, setShowResetModal] = useState(false);

  useEffect(() => {
    const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
    const played = JSON.parse(localStorage.getItem(playedKey) || '{}');

    const hexagonBest = savedScores.hexagon || 0;
    const memoryBest = savedScores.memory || 0;
    const flappyBest = savedScores.flappy || 0;
    const towersBest = savedScores.towers || 0;
    const hanoiBest = savedScores.hanoi || 0;
    
    const hexagonPlayed = played.hexagon || 0;
    const memoryPlayed = played.memory || 0;
    const flappyPlayed = played.flappy || 0;
    const towersPlayed = played.towers || 0;
    const hanoiPlayed = played.hanoi || 0;
    
    const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0');

    const achievements: string[] = [];
    if (hexagonBest >= 100) achievements.push('🥞 100 блинов');
    if (hexagonBest >= 500) achievements.push('👑 Master Pancaker');
    if (memoryBest >= 500) achievements.push('🎴 Memonia Master');
    if (towersBest >= 100) achievements.push('🗼 Builder Master');
    if (flappyBest >= 100) achievements.push('🐦 Flappy Master');
    if (hanoiBest >= 900) achievements.push('🪈 Hanoi Master');
    if (hexagonPlayed + memoryPlayed + flappyPlayed + towersPlayed + hanoiPlayed >= 10) achievements.push('🎮 10+ игр позади');
    if (hexagonPlayed + memoryPlayed + flappyPlayed + towersPlayed + hanoiPlayed >= 50) achievements.push('🔥 Одержимый');

    const fetchUserData = async () => {
      try {
        const token = localStorage.getItem('accessToken');
        const response = await api.get('/auth/user', {
          headers: { Authorization: `Bearer ${token}` }
        });
        
        if (response.data.nickname) {
          localStorage.setItem(nicknameKey, response.data.nickname);
          setStats(prev => ({ ...prev, nickname: response.data.nickname }));
        }
        
        if (response.data.best_scores) {
          const uid = localStorage.getItem('userId');
          const scoresKey = `gameScores_${uid}`;
          const scores = JSON.parse(localStorage.getItem(scoresKey) || '{}');
          
          Object.keys(response.data.best_scores).forEach(gameId => {
            scores[gameId] = response.data.best_scores[gameId];
          });
          
          localStorage.setItem(scoresKey, JSON.stringify(scores));
          
          setStats(prev => ({
            ...prev,
            bestScores: {
              hexagon: scores.hexagon || 0,
              memory: scores.memory || 0,
              flappy: scores.flappy || 0,
              towers: scores.towers || 0,
              hanoi: scores.hanoi || 0,
            }
          }));
        }
        
        if (response.data.total_score) {
          const uid = localStorage.getItem('userId');
          const tKey = `totalScore_${uid}`;
          localStorage.setItem(tKey, String(response.data.total_score));
          setStats(prev => ({ ...prev, totalScore: response.data.total_score }));
        }
        
      } catch (err) {
        console.error('Failed to fetch user data', err);
      }
    };
    
    fetchUserData();

    const fetchBalance = async () => {
        try {
          const token = localStorage.getItem('accessToken');
          const response = await api.get('/billing/balance/all', {
            headers: { Authorization: `Bearer ${token}` }
          });
          setBalance({ lamps: response.data.lamps || 0, tickets: response.data.tickets || 0 });
        } catch (err) {
          console.error('Failed to fetch balance', err);
        }
    };
      
    fetchBalance();

    setStats(prev => ({
      ...prev,
      totalScore: totalScore,
      bestScores: { hexagon: hexagonBest, memory: memoryBest, flappy: flappyBest, towers: towersBest, hanoi: hanoiBest },
      gamesPlayed: { hexagon: hexagonPlayed, memory: memoryPlayed, flappy: flappyPlayed, towers: towersPlayed, hanoi: hanoiPlayed },
      achievements
    }));
  }, [userId, storageKey, playedKey, totalScoreKey, nicknameKey]);

  const handleSetNickname = async () => {
    const newNick = prompt('Введите ваш ник:', stats.nickname);
    if (!newNick || !newNick.trim()) return;

    try {
      const token = localStorage.getItem('accessToken');
      await api.post('/auth/update-nickname', 
        { nickname: newNick.trim() },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      
      localStorage.setItem(nicknameKey, newNick.trim());
      setStats(prev => ({ ...prev, nickname: newNick.trim() }));
      alert('✅ Никнейм обновлён!');
    } catch (err) {
      alert('❌ Ошибка при обновлении ника');
      console.error(err);
    }
  };

  const confirmResetStats = () => {
    localStorage.removeItem(storageKey);
    localStorage.removeItem(playedKey);
    localStorage.removeItem(totalScoreKey);
    localStorage.removeItem(nicknameKey);
    setShowResetModal(false);
    window.location.reload();
  };

  const handleBack = () => {
    navigate('/');
  };

  const bestScoreRows = [
    { key: 'hexagon', icon: '🥞', label: 'Pancaker', value: stats.bestScores.hexagon },
    { key: 'memory', icon: '🎴', label: 'Memonia', value: stats.bestScores.memory },
    { key: 'flappy', icon: '🐦', label: 'Flappy Bird', value: stats.bestScores.flappy },
    { key: 'towers', icon: '🗼', label: 'Builder', value: stats.bestScores.towers },
    { key: 'hanoi', icon: '🪈', label: 'Hanoi', value: stats.bestScores.hanoi },
  ];

  return (
    <PageShell width="narrow">
      <PageHeader title="Профиль" onBack={handleBack} backLabel="На главную" />

      <div className="eh-ring mb-8 flex flex-wrap items-center gap-4 rounded-lg border border-white/10 bg-nebula p-6">
        <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-horizon-gold to-horizon-ember text-3xl shadow-glow-gold">
          {stats.achievements.length >= 2 ? '👑' : '🥞'}
        </div>
        <div className="min-w-0">
          <button
            onClick={handleSetNickname}
            className="font-display text-xl font-semibold text-text-primary transition-colors hover:text-indigo-soft"
          >
            {stats.nickname} ✏️
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
        <StatCard size="md" value={stats.totalScore} label="🏆 Всего очков" />
        <StatCard size="md" value={stats.bestScores.hexagon} label="🥞 Pancaker" />
        <StatCard size="md" value={stats.bestScores.memory} label="🎴 Memonia" />
        <StatCard size="md" value={stats.bestScores.flappy} label="🐦 Flappy Bird" />
        <StatCard size="md" value={stats.bestScores.towers} label="🗼 Builder" />
        <StatCard size="md" value={stats.bestScores.hanoi} label="🪈 Hanoi" />
      </div>

      {stats.achievements.length > 0 && (
        <div className="mb-8">
          <h3 className="mb-3 font-display text-lg font-semibold text-text-primary">🏅 Достижения</h3>
          <div className="flex flex-wrap gap-2">
            {stats.achievements.map((ach, i) => (
              <Badge key={i} tone="gold">{ach}</Badge>
            ))}
          </div>
        </div>
      )}

      <div className="mb-8 flex flex-wrap gap-3">
        <span className="inline-flex items-center gap-1.5 rounded-sm border border-horizon-gold/30 bg-horizon-gold/10 px-3 py-1.5 font-hud text-sm tabular-nums text-horizon-gold">
          💡 {balance.lamps} лампочек
        </span>
        <span className="inline-flex items-center gap-1.5 rounded-sm border border-photon-cyan/30 bg-photon-cyan/10 px-3 py-1.5 font-hud text-sm tabular-nums text-photon-cyan">
          🎫 {balance.tickets} билетиков
        </span>
      </div>

      <div className="mb-8">
        <h3 className="mb-3 font-display text-lg font-semibold text-text-primary">📊 Рекорды по играм</h3>
        <div className="divide-y divide-white/10 rounded-md border border-white/10 bg-nebula">
          {bestScoreRows.map((row) => (
            <button
              key={row.key}
              type="button"
              onClick={() => navigate(GAME_ROUTES[row.key])}
              className="flex w-full items-center justify-between px-4 py-3 text-left transition-colors hover:bg-white/5"
            >
              <span className="text-text-secondary">{row.icon} {row.label}</span>
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
          Будут удалены локальные рекорды, счётчики игр и никнейм на этом устройстве.
          Это действие нельзя отменить.
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
