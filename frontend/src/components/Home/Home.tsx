import { useNavigate } from 'react-router-dom';

const games = [
  {
    id: 'hexagon',
    name: 'Никуся — Блинопёк',
    description: 'Гексагональный пазл с блинами',
    icon: '🥞',
    path: '/game/hexagon',
    available: true,
  },
  {
    id: 'flappy',
    name: 'Flappy Bird',
    description: 'Лети и не врезайся в трубы',
    icon: '🐦',
    path: '/game/flappy',
    available: false,
  },
  {
    id: 'towers',
    name: 'Башенки',
    description: 'Строй башню из падающих блоков',
    icon: '🗼',
    path: '/game/towers',
    available: false,
  },
  {
    id: 'memory',
    name: 'Меморина',
    description: 'Найди пары карт',
    icon: '🎴',
    path: '/game/memory',
    available: false,
  },
];

export function Home() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');

  const handleLogout = () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('userId');
    window.dispatchEvent(new Event('storage'));
    navigate('/login');
  };

  return (
    <div className="landing">
      {/* Хедер с логотипом, меню и кнопкой выхода */}
      <header className="landing-header">
        <div className="logo">
          <h1>🎮 EventHorizon</h1>
        </div>
        
        <nav className="main-nav">
          <a href="#" onClick={(e) => { e.preventDefault(); navigate('/'); }}>Главная</a>
          <a href="#" onClick={(e) => { e.preventDefault(); navigate('/leaderboard'); }}>Лидерборд</a>
          <a href="#" onClick={(e) => { e.preventDefault(); navigate('/profile'); }}>Профиль</a>
          <a href="#" onClick={(e) => { e.preventDefault(); window.open('https://boosty.to/eastwesser', '_blank'); }}>Поддержать</a>
        </nav>

        {token && (
          <button onClick={handleLogout} className="logout-btn-header" title="Выйти">
            🚪 Выйти
          </button>
        )}
      </header>

      <main className="landing-main">
        <section className="hero">
          <h2>Выбери игру и ставь рекорды!</h2>
          <p>Играй в мини-игры, зарабатывай лампочки и билетики, становись лучшим в лидерборде!</p>
        </section>

        <section className="games-section">
          <h3>🎮 Игры</h3>
          <div className="game-cards">
            {games.map((game) => (
              <div
                key={game.id}
                className={`game-card ${game.available ? 'active' : 'disabled'}`}
                onClick={() => game.available && navigate(game.path)}
              >
                <div className="game-icon">{game.icon}</div>
                <h4>{game.name}</h4>
                <p>{game.description}</p>
                <span className="badge">
                  {game.available ? 'Играть' : 'Скоро'}
                </span>
              </div>
            ))}
          </div>
        </section>
      </main>

      <footer className="landing-footer">
        <p>© 2026 EventHorizon. Игры без FOMO и скрытых обнулений.</p>
      </footer>
    </div>
  );
}
