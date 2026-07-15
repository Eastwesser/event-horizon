// frontend/src/components/Games/Hexagon/HexagonGame.tsx
import { useEffect, useState } from 'react';
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
  const [useSpacePancakes, setUseSpacePancakes] = useState(false);
  
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

  // Загружаем настройки скинов из localStorage
  useEffect(() => {
    const saved = localStorage.getItem('hexagon_space_pancakes');
    if (saved !== null) setUseSpacePancakes(saved === 'true');
  }, []);

  // Сохраняем настройки скинов
  const toggleSpacePancakes = () => {
    const newVal = !useSpacePancakes;
    setUseSpacePancakes(newVal);
    localStorage.setItem('hexagon_space_pancakes', String(newVal));
  };

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
            {useSpacePancakes && skins.hexagon.hasSpacePancakes ? '🌌' : '🥞'} Счёт: {score}
          </div>
          <div className="level">🍴 Уровень: {level}</div>
          
          {/* Кнопка переключения скина */}
          {skins.hexagon.hasSpacePancakes && (
            <button 
              className={`skin-toggle-btn ${useSpacePancakes ? 'active' : ''}`}
              onClick={toggleSpacePancakes}
              title="Космические блины"
            >
              {useSpacePancakes ? '🌌' : '🥞'} Космические блины
            </button>
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
          skinMode={useSpacePancakes && skins.hexagon.hasSpacePancakes ? 'space' : 'default'}
        />
        <Tray stacks={tray} skinMode={useSpacePancakes && skins.hexagon.hasSpacePancakes ? 'space' : 'default'} />
      </div>

      {isGameOver && (
        <div className="game-over-overlay">
          <div className="game-over-modal">
            <h2>
              {useSpacePancakes && skins.hexagon.hasSpacePancakes ? '🌌' : '🥞'} 
              Игра окончена! 
              {useSpacePancakes && skins.hexagon.hasSpacePancakes ? '🌌' : '🥞'}
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
