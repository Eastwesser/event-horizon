import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { HexGrid } from './HexGrid';
import { Tray } from './Tray';
import { Balance } from '../../Billing/Balance';
import { Leaderboard } from '../../Leaderboard/Leaderboard';
import { useGameStore } from '../../../store/gameStore';

export function HexagonGame() {
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

  const handleDrop = (item: any, coord: any) => {
    addPancakeToHex(item.id, coord);
  };

  const handleEndGame = () => {
    console.log('End game button clicked, current score:', score);
    if (confirm('Завершить игру? Ваш прогресс будет сохранён.')) {
      setGameOver(score);
    }
  };

  const handleBack = () => {
    navigate('/');
  };

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="game-container">

        <div className="game-header">

          <div className="score">🥞 Счёт: {score}</div>
          <div className="level">🍴 Уровень: {level}</div>
          
          <div className="header-buttons">
            <button onClick={handleEndGame} className="endgame-btn">
              ⏹️ Завершить
            </button>
            <button onClick={handleBack} className="back-btn">
              ← На главную
            </button>
          </div>

          <Balance />
          <Leaderboard />

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
              <button onClick={handleBack}>🏠 На главную</button>  {/* 👈 вместо handleLogout */}
            </div>
          </div>
        </div>
      )}
    </DndProvider>
  );
}
