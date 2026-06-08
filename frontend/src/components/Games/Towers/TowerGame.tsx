// frontend/src/components/Games/Towers/TowerGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTowerStore } from '../../../store/towerStore';
import { Balance } from '../../Billing/Balance';
import './TowerGame.css';

export function TowerGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
  
  const {
    towerBlocks,
    currentBlockX,
    blockWidth,
    score,
    level,
    combo,
    gameOver,
    GAME_WIDTH,
    GAME_HEIGHT,
    startGame,
    dropBlock,
    submitScore,
  } = useTowerStore();
  
  // Проверка авторизации
  useEffect(() => {
    if (!token) {
      navigate('/login');
    } else {
      startGame();
    }
  }, [token, navigate, startGame]);
  
  // Обработка кликов и пробела
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        if (!gameOver) {
          dropBlock();
        }
      }
    };
    
    const handleCanvasClick = () => {
      if (!gameOver) {
        dropBlock();
      }
    };
    
    window.addEventListener('keydown', handleKeyPress);
    
    const canvas = canvasRef.current;
    if (canvas) {
      canvas.addEventListener('click', handleCanvasClick);
    }
    
    return () => {
      window.removeEventListener('keydown', handleKeyPress);
      if (canvas) {
        canvas.removeEventListener('click', handleCanvasClick);
      }
    };
  }, [gameOver, dropBlock]);
  
  // Ручное сохранение рекорда
  const handleManualSave = async () => {
    await submitScore();
    setSaveMessage({ type: 'success', text: '✅ Рекорд сохранён!' });
    setTimeout(() => setSaveMessage(null), 3000);
  };
  
  const handleResetGame = () => {
    startGame();
  };
  
  const handleBack = () => {
    navigate('/');
  };
  
  // Получение цвета блока в зависимости от уровня
  const getBlockColor = (blockLevel: number) => {
    const colors = [
      '#FF6B6B', // 1-2 🔴 Красный
      '#FFA500', // 3-4 🟠 Оранжевый
      '#FFD700', // 5-6 🟡 Жёлтый
      '#4ADE80', // 7-8 🟢 Зелёный
      '#60A5FA', // 9-10 🔵 Голубой
      '#818CF8', // 11-12 🟣 Синий
      '#C084FC', // 13+ 🟣 Фиолетовый
    ];
    const index = Math.min(Math.floor((blockLevel - 1) / 2), colors.length - 1);
    return colors[index];
  };
  
  // Отрисовка игры
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Очищаем canvas
    ctx.clearRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
    
    // Фон
    ctx.fillStyle = '#1a1a2e';
    ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
    
    // Рисуем башню
    const blockHeight = 25;
    const startY = GAME_HEIGHT - 50;
    
    for (let i = 0; i < towerBlocks.length; i++) {
      const blockW = towerBlocks[i];
      const blockX = (GAME_WIDTH - blockW) / 2;
      const blockY = startY - (i * blockHeight);
      
      // Градиент для объёма
      const gradient = ctx.createLinearGradient(blockX, blockY, blockX + blockW, blockY);
      gradient.addColorStop(0, getBlockColor(i + 1));
      gradient.addColorStop(1, getBlockColor(i + 1) + 'aa');
      
      ctx.fillStyle = gradient;
      ctx.fillRect(blockX, blockY, blockW, blockHeight - 2);
      
      // Обводка
      ctx.strokeStyle = '#ffffffaa';
      ctx.strokeRect(blockX, blockY, blockW, blockHeight - 2);
      
      // Текстура (линии)
      ctx.fillStyle = '#ffffff33';
      for (let j = 0; j < 3; j++) {
        ctx.fillRect(blockX + 5 + j * (blockW - 10) / 3, blockY + 5, 2, blockHeight - 12);
      }
    }
    
    // Рисуем текущий движущийся блок
    const currentY = startY - (towerBlocks.length * blockHeight);
    const gradientCurrent = ctx.createLinearGradient(currentBlockX, currentY, currentBlockX + blockWidth, currentY);
    const currentColor = getBlockColor(towerBlocks.length + 1);
    gradientCurrent.addColorStop(0, currentColor);
    gradientCurrent.addColorStop(1, currentColor + 'aa');
    
    ctx.fillStyle = gradientCurrent;
    ctx.fillRect(currentBlockX, currentY, blockWidth, blockHeight - 2);
    
    ctx.strokeStyle = '#ffffff';
    ctx.strokeRect(currentBlockX, currentY, blockWidth, blockHeight - 2);
    
    // Добавляем тень для эффекта парящего блока
    ctx.shadowBlur = 10;
    ctx.shadowColor = 'rgba(0,0,0,0.5)';
    ctx.fillRect(currentBlockX, currentY, blockWidth, blockHeight - 2);
    ctx.shadowBlur = 0;
    
    // Рисуем стрелки направления
    ctx.font = '24px monospace';
    ctx.fillStyle = '#ffffffaa';
    if (useTowerStore.getState().direction === 1) {
      ctx.fillText('→', GAME_WIDTH - 30, currentY + 20);
    } else {
      ctx.fillText('←', 10, currentY + 20);
    }
    
    // Game Over экран
    if (gameOver) {
      ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
      ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
      
      ctx.font = 'bold 28px monospace';
      ctx.fillStyle = '#FF6B6B';
      ctx.fillText('GAME OVER', GAME_WIDTH / 2 - 90, GAME_HEIGHT / 2 - 40);
      
      ctx.font = '20px monospace';
      ctx.fillStyle = '#FFD700';
      ctx.fillText(`Счёт: ${score}`, GAME_WIDTH / 2 - 50, GAME_HEIGHT / 2 + 20);
      
      ctx.font = '14px monospace';
      ctx.fillStyle = '#ffffff';
      ctx.fillText('Нажмите "Новая игра"', GAME_WIDTH / 2 - 80, GAME_HEIGHT / 2 + 70);
    }
  }, [towerBlocks, currentBlockX, blockWidth, score, gameOver, GAME_WIDTH, GAME_HEIGHT]);
  
  // Получение множителя для отображения
  const getMultiplierDisplay = () => {
    if (combo >= 5) return `x${combo - 2}`;
    if (combo >= 4) return 'x3';
    if (combo >= 3) return 'x2';
    return 'x1';
  };
  
  return (
    <div className="tower-container">
      {saveMessage && (
        <div className={`tower-toast tower-toast--${saveMessage.type}`}>
          {saveMessage.text}
        </div>
      )}
      
      <div className="tower-header">
        <Balance />
        <div className="tower-stats">
          <div className="tower-debug" style={{ fontSize: '12px', opacity: 0.7, textAlign: 'center', marginTop: '8px' }}>
            Комбо: {combo} | Очки за блок: {useTowerStore.getState().calculateBlockScore()}
          </div>
          <div className="tower-stat">
            <span className="stat-label">🏆 Счёт</span>
            <span className="stat-value">{score}</span>
          </div>
          <div className="tower-stat">
            <span className="stat-label">📊 Уровень</span>
            <span className="stat-value">{level}</span>
          </div>
          <div className="tower-stat tower-stat--combo">
            <span className="stat-label">⚡ Комбо</span>
            <span className="stat-value">{getMultiplierDisplay()}</span>
          </div>
          <div className="tower-stat">
            <span className="stat-label">🏗️ Высота</span>
            <span className="stat-value">{towerBlocks.length}</span>
          </div>
        </div>
        
        <div className="tower-buttons">
          <button onClick={handleResetGame} className="tower-btn tower-btn--new">
            🔄 Новая игра
          </button>
          <button onClick={handleManualSave} className="tower-btn tower-btn--save">
            💾 Сохранить рекорд
          </button>
          <button onClick={handleBack} className="tower-btn tower-btn--back">
            ← На главную
          </button>
        </div>
      </div>
      
      <div className="tower-canvas-wrapper">
        <canvas
          ref={canvasRef}
          width={GAME_WIDTH}
          height={GAME_HEIGHT}
          className="tower-canvas"
        />
      </div>
      
      <div className="tower-rules">
        <details>
          <summary>📖 Как играть?</summary>
          <p>🏗️ Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы положить блок на башню</p>
          <p>🎯 Чем точнее попадание, тем шире будет следующий блок</p>
          <p>⚡ 3 блока подряд = x2, 4 = x3, 5+ = x{combo >= 5 ? combo - 2 : 'N'} множитель очков</p>
          <p>🏆 Очки: 10 × уровень × множитель</p>
          <p>💡 Башня сужается при неточном попадании!</p>
        </details>
      </div>
    </div>
  );
}