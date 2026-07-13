// frontend/src/components/Games/Hexagon/HexagonGame.tsx
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { HexGrid } from './HexGrid';
import { Tray } from './Tray';
import { Balance } from '../../Billing/Balance';
import { Leaderboard } from '../../Leaderboard/Leaderboard';
import { useGameStore } from '../../../store/gameStore';
import { useSkins } from '../../../hooks/useSkins';

export function HexagonGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { skins, loading: skinsLoading } = useSkins();
  
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

  if (skinsLoading) {
    return (
      <div className="game-container">
        <div className="game-header">
          <div className="score">🥞 Загрузка...</div>
        </div>
      </div>
    );
  }

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="game-container">
        <div className="game-header">
          <div className="score">
            {skins.hexagon.hasSpacePancakes ? '🌌' : '🥞'} Счёт: {score}
          </div>
          <div className="level">🍴 Уровень: {level}</div>
          
          {skins.hexagon.hasSpacePancakes && (
            <div className="skin-badge" style={{ 
              background: 'linear-gradient(135deg, #6C5CE7, #A29BFE)',
              padding: '0.25rem 1rem',
              borderRadius: '20px',
              fontSize: '0.8rem',
              fontWeight: 'bold'
            }}>
              🌌 Космические блины
            </div>
          )}
          
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
      
        <HexGrid 
          tiles={tiles} 
          onDrop={handleDrop}
          skinMode={skins.hexagon.hasSpacePancakes ? 'space' : 'default'}
        />
        <Tray stacks={tray} />
      </div>

      {isGameOver && (
        <div className="game-over-overlay">
          <div className="game-over-modal">
            <h2>
              {skins.hexagon.hasSpacePancakes ? '🌌' : '🥞'} 
              Игра окончена! 
              {skins.hexagon.hasSpacePancakes ? '🌌' : '🥞'}
            </h2>
            <p>Вы испекли {finalScore} блинов!</p>
            <p>Никуся-Блинопёк счастлива! 🎉</p>
            <div className="game-over-buttons">
              <button onClick={() => initGame()}>🔄 Новая игра</button>
              <button onClick={handleBack}>🏠 На главную</button>
            </div>
          </div>
        </div>
      )}
    </DndProvider>
  );
}
