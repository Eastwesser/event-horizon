// frontend/src/components/Profile/Profile.tsx
import { useState, useEffect } from "react";

export function Profile() {
  const email = localStorage.getItem('userEmail') || 'unknown@example.com';
  const userId = localStorage.getItem('userId') || '';
  
  const [stats, setStats] = useState({
    nickname: localStorage.getItem('nickname') || email.split('@')[0],
    totalScore: 0,
    bestScores: { hexagon: 0 },
    gamesPlayed: { hexagon: 0 },
    achievements: [] as string[]
  });

  useEffect(() => {
    // Загружаем статистику из localStorage
    const savedScores = JSON.parse(localStorage.getItem('gameScores') || '{}');
    const hexagonBest = savedScores.hexagon || 0;
    const hexagonPlayed = parseInt(localStorage.getItem('hexagonGamesPlayed') || '0');
    const totalScore = parseInt(localStorage.getItem('totalScore') || '0');
    
    // Простые достижения
    const achievements: string[] = [];
    if (hexagonBest >= 100) achievements.push('🥞 100 блинов');
    if (hexagonBest >= 500) achievements.push('👑 Мастер-блинопёк');
    if (hexagonPlayed >= 10) achievements.push('🎮 Заядлый игрок');
    if (hexagonPlayed >= 50) achievements.push('🔥 Одержимый');
    
    setStats(prev => ({
      ...prev,
      totalScore: totalScore,
      bestScores: { hexagon: hexagonBest },
      gamesPlayed: { hexagon: hexagonPlayed },
      achievements
    }));
  }, []);

  const handleSetNickname = () => {
    const newNick = prompt('Введите ваш ник:', stats.nickname);
    if (newNick && newNick.trim()) {
      localStorage.setItem('nickname', newNick.trim());
      setStats(prev => ({ ...prev, nickname: newNick.trim() }));
    }
  };

  const handleResetStats = () => {
    if (confirm('Сбросить всю статистику? Это действие необратимо.')) {
      localStorage.removeItem('gameScores');
      localStorage.removeItem('hexagonGamesPlayed');
      localStorage.removeItem('totalScore');
      localStorage.removeItem('nickname');
      window.location.reload();
    }
  };

  return (
    <div className="profile-container">
      <div className="profile-header">
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
          <div className="stat-label">🥞 Лучший результат</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.gamesPlayed.hexagon}</div>
          <div className="stat-label">🎮 Сыграно партий</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">🥞</div>
          <div className="stat-label">Любимая игра</div>
          <div className="stat-sub">Никуся-Блинопёк</div>
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

      <div className="profile-best">
        <h3>📊 Рекорды по играм</h3>
        <div className="best-row">
          <span>Никуся-Блинопёк</span>
          <span className="best-score">{stats.bestScores.hexagon} 🥞</span>
        </div>
      </div>

      <button onClick={handleResetStats} className="reset-stats-btn">
        🗑️ Сбросить статистику
      </button>
    </div>
  );
}
