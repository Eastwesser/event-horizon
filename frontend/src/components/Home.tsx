import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { HexGrid } from './Game/HexGrid';
import { Tray } from './Game/Tray';
import { useGameStore } from '../store/gameStore';

export function Home() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { score, level, tiles, tray, initGame, addPancakeToHex } = useGameStore();

  useEffect(() => {
    if (!token) {
      navigate('/login');
    } else {
      initGame();
    }
  }, [token, navigate, initGame]);

  const handleLogout = () => {
    localStorage.removeItem('accessToken');
    window.dispatchEvent(new Event('storage'));
    navigate('/login');
  };

  const handleDrop = (item: any, coord: any) => {
    // item.id — это id стопки из подноса
    addPancakeToHex(item.id, coord);
  };

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="game-container">
        <div className="game-header">
          <div className="score">🥞 Счёт: {score}</div>
          <div className="level">🍴 Уровень: {level}</div>
          <button onClick={handleLogout} className="logout-btn">
            🚪 Выйти
          </button>
        </div>
        
        <HexGrid tiles={tiles} onDrop={handleDrop} />
        <Tray stacks={tray} />
      </div>
    </DndProvider>
  );
}
