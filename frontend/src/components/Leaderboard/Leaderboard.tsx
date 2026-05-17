import { useEffect, useState } from 'react';

interface LeaderboardEntry {
  rank: number;
  userId: string;
  userEmail: string;
  score: number;
}

export function Leaderboard() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws/leaderboard');
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setEntries(data.entries || []);
      } catch (e) {
        console.error('Failed to parse:', e);
      }
    };
    ws.onerror = (error) => console.error('WebSocket error:', error);
    return () => ws.close();
  }, []); // без зависимости от isOpen — постоянно слушаем

  return (
    <>
      <button onClick={() => setIsOpen(true)} className="leaderboard-btn">
        🏆 Топ-10
      </button>
      
      {isOpen && (
        <div className="leaderboard-overlay">
          <div className="leaderboard-modal">
            <h2>🏆 Лучшие блинопёки 🏆</h2>
            <table>
              <thead>
                <tr>
                  <th>#</th>
                  <th>Игрок</th>
                  <th>Очки</th>
                </tr>
              </thead>
              <tbody>
                {entries.map(entry => (
                  <tr key={entry.userId}>
                    <td>{entry.rank}</td>
                    <td>{entry.userEmail?.split('@')[0] || entry.userId.slice(0, 8)}</td>
                    <td>{entry.score}</td>
                  </tr>
                ))}
                {entries.length === 0 && (
                  <tr>
                    <td colSpan={3}>Нет данных</td>
                  </tr>
                )}
              </tbody>
            </table>
            <button onClick={() => setIsOpen(false)}>Закрыть</button>
          </div>
        </div>
      )}
    </>
  );
}
