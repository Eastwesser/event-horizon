// frontend/src/components/Games/Flappy/FlappyGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFlappyStore } from '../../../store/flappyStore';
import { useSkins } from '../../../hooks/useSkins';
import './FlappyGame.css';
import api from '../../../services/api';

export function FlappyGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const { skins, loading: skinsLoading } = useSkins();
  
  const {
    birdY,
    pipes,
    score,
    gameOver,
    started,
    jump,
    resetGame,
  } = useFlappyStore();
  
  const GAME_WIDTH = 800;
  const GAME_HEIGHT = 500;
  const BIRD_SIZE = 30;
  
  // Проверка авторизации
  useEffect(() => {
    if (!token) {
      navigate('/login');
    }
  }, [token, navigate]);
  
  // Обработка кликов и пробела для прыжка
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        jump();
      }
    };
    
    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [jump]);
  
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  const handleManualSave = async () => {
    const { score } = useFlappyStore.getState();
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const token = localStorage.getItem('accessToken');
    
    try {
      const response = await api.post('/game/submit', {
        user_id: userId,
        game_id: 'flappy',
        level: 1,
        score: score,
        user_email: userEmail,
        seed: `flappy_manual_${Date.now()}`,
        moves: [],
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      if (response.status === 200) {
        const userId = localStorage.getItem('userId');
        const storageKey = `gameScores_${userId}`;
        const totalScoreKey = `totalScore_${userId}`;
        
        const savedScores = JSON.parse(localStorage.getItem(storageKey) || '{}');
        const currentBest = savedScores.flappy || 0;
        
        if (score > currentBest) {
          savedScores.flappy = score;
          localStorage.setItem(storageKey, JSON.stringify(savedScores));
        }
        
        const played = parseInt(localStorage.getItem(`flappyGamesPlayed_${userId}`) || '0');
        localStorage.setItem(`flappyGamesPlayed_${userId}`, String(played + 1));
        
        const totalScore = parseInt(localStorage.getItem(totalScoreKey) || '0');
        localStorage.setItem(totalScoreKey, String(totalScore + score));
        
        setSaveMessage({ type: 'success', text: '✅ Рекорд сохранён!' });
        setTimeout(() => setSaveMessage(null), 3000);
      }
    } catch (err) {
      setSaveMessage({ type: 'error', text: '❌ Ошибка' });
      setTimeout(() => setSaveMessage(null), 3000);
    }
  };

  // Определяем цвета для скинов
  const getBirdColor = () => {
    if (skins.flappy.hasGoldenBird) {
      return '#FFD700'; // Золотой
    }
    return '#FFD700'; // Стандартный (тоже золотой, можно поменять)
  };

  const getPipeColor = () => {
    if (skins.flappy.hasRainbowPipes) {
      return 'rainbow';
    }
    return '#228B22'; // Стандартный зеленый
  };

  // Функция для рисования радужной трубы
  const drawRainbowPipe = (ctx: CanvasRenderingContext2D, x: number, y: number, height: number, width: number, isTop: boolean) => {
    const colors = ['#FF6B6B', '#FFA500', '#FFD700', '#4ADE80', '#60A5FA', '#818CF8', '#C084FC'];
    const segmentHeight = Math.min(20, height / 7);
    
    for (let i = 0; i < 7; i++) {
      const startY = isTop ? y + i * segmentHeight : y + i * segmentHeight;
      const color = colors[i % colors.length];
      
      ctx.fillStyle = color;
      ctx.fillRect(x, startY, width, segmentHeight);
    }
  };

  // Отрисовка игры
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Очищаем canvas
    ctx.clearRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
    
    // Фон (небо)
    const gradient = ctx.createLinearGradient(0, 0, 0, GAME_HEIGHT);
    gradient.addColorStop(0, '#87CEEB');
    gradient.addColorStop(1, '#E0F6FF');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
    
    // Рисуем облака (декор)
    ctx.fillStyle = 'rgba(255, 255, 255, 0.7)';
    ctx.beginPath();
    ctx.ellipse(150, 80, 40, 30, 0, 0, Math.PI * 2);
    ctx.ellipse(180, 70, 50, 35, 0, 0, Math.PI * 2);
    ctx.ellipse(120, 70, 35, 25, 0, 0, Math.PI * 2);
    ctx.fill();
    
    ctx.beginPath();
    ctx.ellipse(600, 120, 45, 35, 0, 0, Math.PI * 2);
    ctx.ellipse(640, 110, 55, 40, 0, 0, Math.PI * 2);
    ctx.ellipse(570, 110, 40, 30, 0, 0, Math.PI * 2);
    ctx.fill();
    
    // Рисуем трубы
    const pipeColor = getPipeColor();
    const isRainbow = pipeColor === 'rainbow';
    
    pipes.forEach(pipe => {
      if (isRainbow) {
        // Верхняя труба (радужная)
        drawRainbowPipe(ctx, pipe.x, 0, pipe.topHeight, 60, true);
        // Нижняя труба (радужная)
        drawRainbowPipe(ctx, pipe.x, pipe.bottomY, GAME_HEIGHT - pipe.bottomY, 60, false);
        
        // Обводка для радужных труб
        ctx.strokeStyle = 'rgba(255,255,255,0.3)';
        ctx.strokeRect(pipe.x, 0, 60, pipe.topHeight);
        ctx.strokeRect(pipe.x, pipe.bottomY, 60, GAME_HEIGHT - pipe.bottomY);
      } else {
        // Обычные зеленые трубы
        ctx.fillStyle = '#228B22';
        ctx.fillRect(pipe.x, 0, 60, pipe.topHeight);
        ctx.fillStyle = '#2E7D32';
        ctx.fillRect(pipe.x - 5, pipe.topHeight - 30, 70, 30);
        
        ctx.fillStyle = '#228B22';
        ctx.fillRect(pipe.x, pipe.bottomY, 60, GAME_HEIGHT - pipe.bottomY);
        ctx.fillStyle = '#2E7D32';
        ctx.fillRect(pipe.x - 5, pipe.bottomY, 70, 30);
        
        // Детали труб
        ctx.fillStyle = '#1B5E20';
        for (let i = 0; i < 3; i++) {
          ctx.fillRect(pipe.x + 10, pipe.topHeight - 20 + i * 10, 40, 5);
        }
        for (let i = 0; i < 3; i++) {
          ctx.fillRect(pipe.x + 10, pipe.bottomY + 10 + i * 10, 40, 5);
        }
      }
    });
    
    // Рисуем птичку
    const birdColor = getBirdColor();
    ctx.save();
    ctx.shadowBlur = 10;
    ctx.shadowColor = 'rgba(0,0,0,0.3)';
    
    // Тело
    ctx.fillStyle = birdColor;
    ctx.beginPath();
    ctx.ellipse(100, birdY + BIRD_SIZE/2, BIRD_SIZE/2, BIRD_SIZE/2, 0, 0, Math.PI * 2);
    ctx.fill();
    
    // Если золотая птичка - добавляем блеск
    if (skins.flappy.hasGoldenBird) {
      ctx.shadowBlur = 20;
      ctx.shadowColor = 'rgba(255, 215, 0, 0.5)';
      ctx.fillStyle = 'rgba(255, 255, 255, 0.3)';
      ctx.beginPath();
      ctx.ellipse(95, birdY + BIRD_SIZE/2 - 8, 8, 5, 0, 0, Math.PI * 2);
      ctx.fill();
    }
    
    // Глаз
    ctx.shadowBlur = 0;
    ctx.fillStyle = '#000';
    ctx.beginPath();
    ctx.arc(110, birdY + BIRD_SIZE/2 - 5, 4, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = '#FFF';
    ctx.beginPath();
    ctx.arc(108, birdY + BIRD_SIZE/2 - 6, 1.5, 0, Math.PI * 2);
    ctx.fill();
    
    // Клюв
    ctx.fillStyle = '#FF6347';
    ctx.beginPath();
    ctx.moveTo(115, birdY + BIRD_SIZE/2 - 3);
    ctx.lineTo(125, birdY + BIRD_SIZE/2);
    ctx.lineTo(115, birdY + BIRD_SIZE/2 + 3);
    ctx.fill();
    
    // Крыло
    ctx.fillStyle = skins.flappy.hasGoldenBird ? '#FFC000' : '#FFA500';
    ctx.beginPath();
    ctx.ellipse(90, birdY + BIRD_SIZE/2, 12, 8, -Math.PI / 4, 0, Math.PI * 2);
    ctx.fill();
    
    ctx.restore();
    
    // Счёт
    ctx.font = 'bold 36px "Press Start 2P", monospace';
    ctx.fillStyle = '#FFF';
    ctx.shadowBlur = 0;
    ctx.fillText(`${score}`, GAME_WIDTH / 2 - 20, 60);
    
    // Стартовый экран
    if (!started && !gameOver) {
      ctx.font = 'bold 24px "Press Start 2P", monospace';
      ctx.fillStyle = '#FFF';
      ctx.shadowColor = '#000';
      ctx.fillText('НАЖМИТЕ ПРОБЕЛ', GAME_WIDTH / 2 - 150, GAME_HEIGHT / 2);
      ctx.font = '16px monospace';
      ctx.fillText('или кликните мышкой', GAME_WIDTH / 2 - 110, GAME_HEIGHT / 2 + 50);
    }
    
    // Game Over экран
    if (gameOver) {
      ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
      ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
      
      ctx.font = 'bold 36px "Press Start 2P", monospace';
      ctx.fillStyle = '#FF6B6B';
      ctx.fillText('GAME OVER', GAME_WIDTH / 2 - 120, GAME_HEIGHT / 2 - 40);
      
      ctx.font = '24px monospace';
      ctx.fillStyle = '#FFF';
      ctx.fillText(`Счёт: ${score}`, GAME_WIDTH / 2 - 50, GAME_HEIGHT / 2 + 20);
      
      ctx.font = '16px monospace';
      ctx.fillStyle = '#FFD700';
      ctx.fillText('Нажмите "Новая игра"', GAME_WIDTH / 2 - 100, GAME_HEIGHT / 2 + 80);
    }
  }, [birdY, pipes, score, gameOver, started, GAME_WIDTH, GAME_HEIGHT, BIRD_SIZE, skins]);
  
  const handleCanvasClick = () => {
    jump();
  };
  
  const handleBack = () => {
    navigate('/');
  };
  
  if (skinsLoading) {
    return (
      <div className="flappy-container">
        <div className="flappy-header">
          <div className="flappy-stats">
            <div className="flappy-stat">
              <span className="stat-label">🐦 Загрузка...</span>
            </div>
          </div>
        </div>
      </div>
    );
  }
  
  return (
    <div className="flappy-container">
      {saveMessage && (
        <div className={`flappy-toast flappy-toast--${saveMessage.type}`}>
          {saveMessage.text}
        </div>
      )}
      <div className="flappy-header">
        <div className="flappy-stats">
          <div className="flappy-stat">
            <span className="stat-label">🐦 Счёт</span>
            <span className="stat-value">{score}</span>
          </div>
          {skins.flappy.hasGoldenBird && (
            <div className="flappy-stat" style={{ borderColor: '#FFD700' }}>
              <span className="stat-label">⭐ Скин</span>
              <span className="stat-value" style={{ color: '#FFD700' }}>Золотая птичка</span>
            </div>
          )}
          {skins.flappy.hasRainbowPipes && (
            <div className="flappy-stat" style={{ borderColor: '#FF6B6B' }}>
              <span className="stat-label">🌈 Скин</span>
              <span className="stat-value" style={{ color: '#FF6B6B' }}>Радужные трубы</span>
            </div>
          )}
        </div>
        
        <div className="flappy-buttons">
          <button onClick={resetGame} className="flappy-btn flappy-btn--new">
            🔄 Новая игра
          </button>
          <button onClick={handleManualSave} className="flappy-btn flappy-btn--save">
            💾 Сохранить рекорд
          </button>
          <button onClick={handleBack} className="flappy-btn flappy-btn--back">
            ← На главную
          </button>
        </div>
      </div>

      <div className="flappy-canvas-wrapper">
        <canvas
          ref={canvasRef}
          width={800}
          height={500}
          className="flappy-canvas"
          onClick={handleCanvasClick}
        />
      </div>
      
      <div className="flappy-rules">
        <details>
          <summary>📖 Как играть?</summary>
          <p>🐦 Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы птичка летела вверх</p>
          <p>🚫 Не врезайтесь в трубы и не падайте на землю</p>
          <p>⭐ Каждая пройденная труба = 10 очков</p>
          <p>🏆 Чем дальше, тем выше счёт!</p>
          <p>💡 Чем выше счёт, тем больше билетиков получите</p>
        </details>
      </div>
    </div>
  );
}
