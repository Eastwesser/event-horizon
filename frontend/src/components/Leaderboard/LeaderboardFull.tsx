// frontend/src/components/Leaderboard/LeaderboardFull.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  score: number;
  game_id?: string;
}

export function LeaderboardFull() {
  const navigate = useNavigate();
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedGame, setSelectedGame] = useState<'hexagon' | 'memory' | 'flappy' | 'towers'>('hexagon');

  useEffect(() => {
    setLoading(true);
    fetch(`/api/leaderboard?game_id=${selectedGame}&limit=50`)
      .then(res => res.json())
      .then(data => {
        setEntries(data.entries || []);
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
      default: return '';
    }
  };

  const getRankClass = (rank: number) => {
    switch (rank) {
      case 1: return 'leaderboard-rank--gold';
      case 2: return 'leaderboard-rank--silver';
      case 3: return 'leaderboard-rank--bronze';
      default: return '';
    }
  };

  const handleBack = () => {
    navigate('/');
  };

  return (
    <div className="leaderboard-full">
      <div className="leaderboard-header">
        <div className="leaderboard-title">
          <button onClick={handleBack} className="back-btn-small" title="На главную">
            ←
          </button>
          <h2>🏆 Лидерборд</h2>
        </div>
        <div className="leaderboard-game-selector">
          <button
            className={`game-tab ${selectedGame === 'hexagon' ? 'active' : ''}`}
            onClick={() => setSelectedGame('hexagon')}
          >
            🥞 Блинопёк
          </button>
          <button
              className={`game-tab ${selectedGame === 'flappy' ? 'active' : ''}`}
              onClick={() => setSelectedGame('flappy')}
          >
              🐦 Flappy Bird
          </button>
          <button
            className={`game-tab ${selectedGame === 'memory' ? 'active' : ''}`}
            onClick={() => setSelectedGame('memory')}
          >
            🎴 Мемония
          </button>
          <button
            className={`game-tab ${selectedGame === 'towers' ? 'active' : ''}`}
            onClick={() => setSelectedGame('towers')}
          >
            🗼 Башенки
          </button>
        </div>
      </div>

      {loading ? (
        <div className="leaderboard-loading">
          <div className="spinner"></div>
          <p>Загрузка рекордов...</p>
        </div>
      ) : entries.length === 0 ? (
        <div className="leaderboard-empty">
          <p>😢 Пока нет рекордов в этой игре</p>
          <p>Стань первым!</p>
        </div>
      ) : (
        <div className="leaderboard-table-wrapper">
          <table className="leaderboard-table">
            <thead>
              <tr>
                <th>#</th>
                <th>Игрок</th>
                <th>Очки</th>
                <th>🏅</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry, idx) => {
                const rank = idx + 1;
                const medal = getMedal(rank);
                const rankClass = getRankClass(rank);
                
                return (
                  <tr key={entry.userId || `row-${idx}`} className={rankClass}>
                    <td className="leaderboard-rank">
                      {medal ? <span className="medal">{medal}</span> : rank}
                    </td>
                    <td className="leaderboard-player">
                      <div className="player-avatar">
                        {entry.user_email?.split('@')[0]?.charAt(0).toUpperCase() || '?'}
                      </div>
                      <span className="player-name">
                        {entry.user_email?.split('@')[0] || 'Аноним'}
                      </span>
                    </td>
                    <td className="leaderboard-score">
                      <span className="score-value">{entry.score.toLocaleString()}</span>
                      <span className="score-unit">
                        {selectedGame === 'hexagon' ? '🥞' : 
                        selectedGame === 'memory' ? '🎴' : 
                        selectedGame === 'flappy' ? '🐦' : '🗼'}
                      </span>
                    </td>
                    <td className="leaderboard-badge">
                      {medal && <span className="badge-icon">{medal}</span>}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
