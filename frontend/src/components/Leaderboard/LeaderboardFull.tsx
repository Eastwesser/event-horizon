import { useEffect, useState } from 'react';

export function LeaderboardFull() {
  const [entries, setEntries] = useState([]);

  useEffect(() => {
    fetch('/api/leaderboard?game_id=hexagon&limit=50')
      .then(res => res.json())
      .then(data => setEntries(data.entries || []))
      .catch(console.error);
  }, []);

  return (
    <div className="leaderboard-full">
      <h2>🏆 Лучшие блинопёки 🏆</h2>
      <table>
        <thead>
          <tr><th>#</th><th>Игрок</th><th>Очки</th></tr>
        </thead>
        <tbody>
          {entries.map((entry, idx) => (
            <tr key={entry.userId}>
              <td>{idx + 1}</td>
              <td>{entry.user_email?.split('@')[0] || 'Аноним'}</td>
              <td>{entry.score}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
