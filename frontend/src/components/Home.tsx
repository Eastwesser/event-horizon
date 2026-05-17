import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { HexGrid } from './Game/HexGrid';
import { Tray } from './Game/Tray';
import { Balance } from './Billing/Balance';
import { Leaderboard } from './Leaderboard/Leaderboard';
import { useGameStore } from '../store/gameStore';

export function Home() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { 
    score, 
    level, 
    tiles, 
    tray, 
    initGame, 
    addPancakeToHex, 
    isGameOver, 
    finalScore,
    setGameOver
  } = useGameStore();

  useEffect(() => {
    if (!token) {
      navigate('/login');
    } else {
      initGame();
    }
  }, [token, navigate, initGame]);

  const handleLogout = () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('userId');
    window.dispatchEvent(new Event('storage'));
    navigate('/login');
  };

  const handleDrop = (item: any, coord: any) => {
    addPancakeToHex(item.id, coord);
  };

  const handleEndGame = () => {
    console.log('End game button clicked');  // 👈 
    if (confirm('Завершить игру? Ваш прогресс будет сохранён.')) {
      console.log('Calling setGameOver with score:', score);  // 👈 
      setGameOver(score);
    }
  };

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="game-container">
        <div className="game-header">
          <div className="score">🥞 Счёт: {score}</div>
          <div className="level">🍴 Уровень: {level}</div>
          <Balance />
          <Leaderboard />
          <button onClick={handleEndGame} className="endgame-btn">
            ⏹️ Завершить
          </button>
          <button onClick={handleLogout} className="logout-btn">
            🚪 Выйти
          </button>
        </div>
        
        <HexGrid tiles={tiles} onDrop={handleDrop} />
        <Tray stacks={tray} />
      </div>

      {isGameOver && (
        <div className="game-over-overlay">
          <div className="game-over-modal">
            <h2>🥞 Игра окончена! 🥞</h2>
            <p>Вы испекли {finalScore} блинов!</p>
            <p>Никуся-Блинопёк счастлива! 🎉</p>
            <div className="game-over-buttons">
              <button onClick={() => initGame()}>🔄 Новая игра</button>
              <button onClick={handleLogout}>🚪 Выйти</button>
            </div>
          </div>
        </div>
      )}
    </DndProvider>
  );
}
