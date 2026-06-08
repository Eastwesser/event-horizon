// frontend/src/components/Leaderboard/Leaderboard.tsx
import { useEffect, useState, useRef } from 'react';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  user_email: string;
  score: number;
}

export function Leaderboard() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedGame, setSelectedGame] = useState<'hexagon' | 'memory' | 'flappy' | 'towers'>('hexagon');
  const wsRef = useRef<WebSocket | null>(null);

  const fetchLeaderboard = async () => {
    try {
      const response = await fetch(`/api/leaderboard?game_id=${selectedGame}&limit=10`);
      if (response.ok) {
        const data = await response.json();
        setEntries(data.entries || []);
      }
    } catch (err) {
      console.error('Failed to fetch leaderboard:', err);
    }
  };

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws/leaderboard');
    wsRef.current = ws;
    
    ws.onopen = () => {
      console.log('✅ WebSocket connected');
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log('📥 WebSocket raw data:', data);
        
        if (Array.isArray(data)) {
          setEntries(data);
        } else if (data.entries && Array.isArray(data.entries)) {
          setEntries(data.entries);
        } else if (data.user_id && data.score) {
          fetchLeaderboard();
        } else {
          fetchLeaderboard();
        }
      } catch (e) {
        console.error('Failed to parse:', e);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onclose = () => {
      console.log('WebSocket disconnected');
    };
    
    return () => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.close();
      }
    };
  }, [selectedGame]);

  const handleOpen = () => {
    setIsOpen(true);
    fetchLeaderboard();
  };

  const getMedal = (rank: number) => {
    switch (rank) {
      case 1: return '🥇';
      case 2: return '🥈';
      case 3: return '🥉';
      default: return '';
    }
  };

  return (
    <>
      <button onClick={handleOpen} className="leaderboard-btn">
        🏆 Топ-10
      </button>
      
      {isOpen && (
        <div className="leaderboard-overlay" onClick={() => setIsOpen(false)}>
          <div className="leaderboard-modal" onClick={(e) => e.stopPropagation()}>
            <div className="leaderboard-modal-header">
              <h2>🏆 Топ-10 игроков</h2>
              <button className="leaderboard-close" onClick={() => setIsOpen(false)}>✕</button>
            </div>
            
            <div className="leaderboard-game-selector">

              <button
                className={`game-tab ${selectedGame === 'hexagon' ? 'active' : ''}`}
                onClick={() => setSelectedGame('hexagon')}
              >
                🥞 Блинопёк
              </button>

              <button
                className={`game-tab ${selectedGame === 'memory' ? 'active' : ''}`}
                onClick={() => setSelectedGame('memory')}
              >
                🎴 Мемония
              </button>

              <button
                className={`game-tab ${selectedGame === 'flappy' ? 'active' : ''}`}
                onClick={() => setSelectedGame('flappy')}
              >
                🐦 Flappy Bird
              </button>

              <button
                className={`game-tab ${selectedGame === 'towers' ? 'active' : ''}`}
                onClick={() => setSelectedGame('towers')}
              >
                🗼 Башенки
              </button>
              
            </div>
            
            {entries.length === 0 ? (
              <p style={{ textAlign: 'center', padding: '2rem' }}>Нет данных</p>
            ) : (
              <div className="leaderboard-modal-list">
                {entries.map((entry, idx) => {
                  const rank = idx + 1;
                  const medal = getMedal(rank);
                  
                  return (
                    <div key={entry.userId || idx} className="leaderboard-item">
                      <div className="leaderboard-item-rank">
                        {medal || <span className="rank-number">{rank}</span>}
                      </div>
                      <div className="leaderboard-item-player">
                        <div className="player-avatar-small">
                          {entry.user_email?.split('@')[0]?.charAt(0).toUpperCase() || '?'}
                        </div>
                        <span className="player-name">
                          {entry.user_email?.split('@')[0] || 'Аноним'}
                        </span>
                      </div>
                      <div className="leaderboard-item-score">
                        <span className="score-value">{entry.score.toLocaleString()}</span>
                        <span className="score-unit">
                          {selectedGame === 'hexagon' ? '🥞' : 
                          selectedGame === 'memory' ? '🎴' : 
                          selectedGame === 'flappy' ? '🐦' : '🗼'}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
