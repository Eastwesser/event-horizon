// frontend/src/components/Profile/Profile.tsx
import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';

export function Profile() {
  const navigate = useNavigate();
  const email = localStorage.getItem('userEmail') || 'unknown@example.com';
  const userId = localStorage.getItem('userId') || '';

  // Ключи localStorage с привязкой к userId
  const storageKey = `gameScores_${userId}`;
  const playedKey = `gamesPlayed_${userId}`;
  const totalScoreKey = `totalScore_${userId}`;
  const nicknameKey = `nickname_${userId}`;

  const [stats, setStats] = useState({
    nickname: localStorage.getItem(nicknameKey) || email.split('@')[0],
    totalScore: 0,
    bestScores: { hexagon: 0, memory: 0, flappy: 0, towers: 0 },
    gamesPlayed: { hexagon: 0, memory: 0, flappy: 0, towers: 0 },
    achievements: [] as string[]
  });
  const [balance, setBalance] = useState({ lamps: 0, tickets: 0 });

  useEffect(() => {
    // Загружаем статистику из localStorage
    const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
    const played = JSON.parse(localStorage.getItem(playedKey) || '{}');

    const hexagonBest = savedScores.hexagon || 0;
    const memoryBest = savedScores.memory || 0;
    const flappyBest = savedScores.flappy || 0;
    const towersBest = savedScores.towers || 0;
    
    const hexagonPlayed = played.hexagon || 0;
    const memoryPlayed = played.memory || 0;
    const flappyPlayed = played.flappy || 0;
    const towersPlayed = played.towers || 0;
    
    const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0');

    // Простые достижения
    const achievements: string[] = [];
    if (hexagonBest >= 100) achievements.push('🥞 100 блинов');
    if (hexagonBest >= 500) achievements.push('👑 Мастер-блинопёк');
    if (memoryBest >= 500) achievements.push('🎴 Мастер памяти');
    if (towersBest >= 100) achievements.push('🗼 Мастер башен');
    if (flappyBest >= 100) achievements.push('🐦 Мастер полёта');
    if (hexagonPlayed + memoryPlayed + flappyPlayed + towersPlayed >= 10) achievements.push('🎮 Заядлый игрок');
    if (hexagonPlayed + memoryPlayed + flappyPlayed + towersPlayed >= 50) achievements.push('🔥 Одержимый');

    // 🔥 Функция загрузки данных с бэкенда
    const fetchUserData = async () => {
      try {
        const token = localStorage.getItem('accessToken');
        const response = await api.get('/auth/user', {
          headers: { Authorization: `Bearer ${token}` }
        });
        
        // Обновляем никнейм
        if (response.data.nickname) {
          localStorage.setItem(nicknameKey, response.data.nickname);
          setStats(prev => ({ ...prev, nickname: response.data.nickname }));
        }
        
        // 🔥 СИНХРОНИЗИРУЕМ РЕКОРДЫ ИЗ БЭКЕНДА
        if (response.data.best_scores) {
          const userId = localStorage.getItem('userId');
          const storageKey = `gameScores_${userId}`;
          const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
          
          // Обновляем рекорды из бэкенда
          Object.keys(response.data.best_scores).forEach(gameId => {
            savedScores[gameId] = response.data.best_scores[gameId];
          });
          
          localStorage.setItem(storageKey, JSON.stringify(savedScores));
          
          // Обновляем состояние компонента
          setStats(prev => ({
            ...prev,
            bestScores: {
              hexagon: savedScores.hexagon || 0,
              memory: savedScores.memory || 0,
              flappy: savedScores.flappy || 0,
              towers: savedScores.towers || 0,
            }
          }));
        }
        
        if (response.data.total_score) {
          const userId = localStorage.getItem('userId');
          const totalScoreKey = `totalScore_${userId}`;
          localStorage.setItem(totalScoreKey, String(response.data.total_score));
          setStats(prev => ({ ...prev, totalScore: response.data.total_score }));
        }
        
      } catch (err) {
        console.error('Failed to fetch user data', err);
      }
    };
    
    fetchUserData();

    // Добавить загрузку баланса
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
      bestScores: { hexagon: hexagonBest, memory: memoryBest, flappy: flappyBest, towers: towersBest },
      gamesPlayed: { hexagon: hexagonPlayed, memory: memoryPlayed, flappy: flappyPlayed, towers: towersPlayed },
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

  const handleResetStats = () => {
    if (confirm('Сбросить всю статистику? Это действие необратимо.')) {
      localStorage.removeItem(storageKey);
      localStorage.removeItem(playedKey);
      localStorage.removeItem(totalScoreKey);
      localStorage.removeItem(nicknameKey);
      window.location.reload();
    }
  };

  const handleBack = () => {
    navigate('/');
  };

  return (
    <div className="profile-container">
      <div className="profile-header">
        <button onClick={handleBack} className="back-btn-small" title="На главную">
          ←
        </button>
        <div className="profile-avatar">
          {stats.achievements.length >= 2 ? '👑' : '🥞'}
        </div>
        <div className="profile-info">
          <h2 onClick={handleSetNickname} style={{ cursor: 'pointer' }}>
            {stats.nickname} ✏️
          </h2>
          <p className="profile-email">{email}</p>
          {userId && <p className="profile-id">ID: {userId.slice(0, 8)}...</p>}
        </div>
      </div>

      <div className="profile-stats-grid">
        <div className="stat-card">
          <div className="stat-value">{stats.totalScore}</div>
          <div className="stat-label">🏆 Всего очков</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.bestScores.hexagon}</div>
          <div className="stat-label">🥞 Блинопёк</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.bestScores.memory}</div>
          <div className="stat-label">🎴 Мемония</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.bestScores.flappy}</div>
          <div className="stat-label">🐦 Flappy Bird</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.bestScores.towers}</div>
          <div className="stat-label">🗼 Башенки</div>
        </div>
      </div>

      {stats.achievements.length > 0 && (
        <div className="profile-achievements">
          <h3>🏅 Достижения</h3>
          <div className="achievements-list">
            {stats.achievements.map((ach, i) => (
              <span key={i} className="achievement-badge">{ach}</span>
            ))}
          </div>
        </div>
      )}

      <div className="profile-balance">
        <span>💡 Лампочки: {balance.lamps}</span>
        <span>🎫 Билетики: {balance.tickets}</span>
      </div>

      <div className="profile-best">
        <h3>📊 Рекорды по играм</h3>
        <div className="best-row">
          <span>🥞 Никуся-Блинопёк</span>
          <span className="best-score">{stats.bestScores.hexagon} 🥞</span>
        </div>
        <div className="best-row">
          <span>🎴 Мемония</span>
          <span className="best-score">{stats.bestScores.memory} 🎴</span>
        </div>
        <div className="best-row">
          <span>🐦 Flappy Bird</span>
          <span className="best-score">{stats.bestScores.flappy} 🐦</span>
        </div>
        <div className="best-row">
          <span>🗼 Башенки</span>
          <span className="best-score">{stats.bestScores.towers} 🗼</span>
        </div>
      </div>

      <button onClick={handleResetStats} className="reset-stats-btn">
        🗑️ Сбросить статистику
      </button>
    </div>
  );
}
