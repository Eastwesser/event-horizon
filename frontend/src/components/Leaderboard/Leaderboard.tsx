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
  const wsRef = useRef<WebSocket | null>(null);

  // Функция запроса топа через HTTP
  const fetchLeaderboard = async () => {
    try {
      const response = await fetch('/api/leaderboard?game_id=hexagon&limit=10');
      if (response.ok) {
        const data = await response.json();
        setEntries(data.entries || []);
      }
    } catch (err) {
      console.error('Failed to fetch leaderboard:', err);
    }
  };

  // WebSocket соединение (постоянное)
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
          console.log('📊 New score event, fetching full leaderboard...');
          fetchLeaderboard();
        } else {
          console.log('⏳ Unknown data format, requesting leaderboard...');
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
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // При открытии модального окна
  const handleOpen = () => {
    setIsOpen(true);
    fetchLeaderboard();
  };

  return (
    <>
      <button onClick={handleOpen} className="leaderboard-btn">
        🏆 Топ-10
      </button>
      
      {isOpen && (
        <div className="leaderboard-overlay">
          <div className="leaderboard-modal">
            <h2>🏆 Лучшие блинопёки 🏆</h2>
            {entries.length === 0 ? (
              <p style={{ textAlign: 'center', padding: '2rem' }}>Нет данных</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>#</th>
                    <th>Игрок</th>
                    <th>Очки</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map((entry, idx) => (
                    <tr key={entry.userId || idx}>
                      <td>{entry.rank || idx + 1}</td>
                      <td>{entry.user_email?.split('@')[0] || entry.userId?.slice(0, 8) || 'Аноним'}</td>
                      <td>{entry.score}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
            <button onClick={() => setIsOpen(false)}>Закрыть</button>
          </div>
        </div>
      )}
    </>
  );
}