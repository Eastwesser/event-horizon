// frontend/src/components/Profile/Profile.tsx
import { useState, useEffect } from "react";
import { useNavigate } from 'react-router-dom';

export function Profile() {
  const navigate = useNavigate();
  const email = localStorage.getItem('userEmail') || 'unknown@example.com';
  const userId = localStorage.getItem('userId') || '';
  
  const [stats, setStats] = useState({
    nickname: localStorage.getItem('nickname') || email.split('@')[0],
    totalScore: 0,
    bestScores: { hexagon: 0, memory: 0, flappy: 0, towers: 0 },
    gamesPlayed: { hexagon: 0, memory: 0, flappy: 0, towers: 0 },
    achievements: [] as string[]
  });

  useEffect(() => {
    // Загружаем статистику из localStorage
    const savedScores = JSON.parse(localStorage.getItem('gameScores') || '{}');
    
    const hexagonBest = savedScores.hexagon || 0;
    const memoryBest = savedScores.memory || 0;
    const flappyBest = savedScores.flappy || 0;
    const towersBest = savedScores.towers || 0;
    
    const hexagonPlayed = parseInt(localStorage.getItem('hexagonGamesPlayed') || '0');
    const memoryPlayed = parseInt(localStorage.getItem('memoryGamesPlayed') || '0');
    const flappyPlayed = parseInt(localStorage.getItem('flappyGamesPlayed') || '0');
    const towersPlayed = parseInt(localStorage.getItem('towersGamesPlayed') || '0');
    
    const totalScore = parseInt(localStorage.getItem('totalScore') || '0');

    // Простые достижения
    const achievements: string[] = [];
    if (hexagonBest >= 100) achievements.push('🥞 100 блинов');
    if (hexagonBest >= 500) achievements.push('👑 Мастер-блинопёк');
    if (memoryBest >= 500) achievements.push('🎴 Мастер памяти');
    if (flappyBest >= 100) achievements.push('🐦 Мастер полёта');
    if (hexagonPlayed + memoryPlayed + flappyPlayed >= 10) achievements.push('🎮 Заядлый игрок');
    if (hexagonPlayed + memoryPlayed + flappyPlayed >= 50) achievements.push('🔥 Одержимый');
    if (towersBest >= 100) achievements.push('🗼 Мастер башен');

    setStats(prev => ({
      ...prev,
      totalScore: totalScore,
      bestScores: { hexagon: hexagonBest, memory: memoryBest, flappy: flappyBest, towers: towersBest },
      gamesPlayed: { hexagon: hexagonPlayed, memory: memoryPlayed, flappy: flappyPlayed, towers: towersPlayed },
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
      localStorage.removeItem('memoryGamesPlayed');
      localStorage.removeItem('flappyGamesPlayed');
      localStorage.removeItem('totalScore');
      localStorage.removeItem('nickname');
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
