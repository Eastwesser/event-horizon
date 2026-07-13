// frontend/src/components/Games/Memonia/MemoryGame.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMemoryStore } from '../../../store/memoryStore';
import { useSkins } from '../../../hooks/useSkins';
import { MemoryBoard } from './MemoryBoard';
import api from '../../../services/api';
import './memory.css';

export function MemoryGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { skins, loading: skinsLoading } = useSkins();
  const [useAnimalCards, setUseAnimalCards] = useState(false);
  
  const {
    moves,
    matchedPairs,
    gameOver,
    score,
    multiplier,
    combo,
    initGame,
    resetGame,
  } = useMemoryStore();
  
  const totalPairs = 15;
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  // Загружаем настройки скинов из localStorage
  useEffect(() => {
    const saved = localStorage.getItem('memory_animal_cards');
    if (saved !== null) setUseAnimalCards(saved === 'true');
  }, []);

  // Сохраняем настройки скинов
  const toggleAnimalCards = () => {
    const newVal = !useAnimalCards;
    setUseAnimalCards(newVal);
    localStorage.setItem('memory_animal_cards', String(newVal));
    // Перезапускаем игру с новым скином
    resetGame();
  };
  
  useEffect(() => {
    if (!token) {
      navigate('/login');
    } else {
      initGame();
    }
  }, [token, navigate, initGame]);
  
  const handleNewGame = () => {
    resetGame();
  };
  
  const handleBack = () => {
    navigate('/');
  };
  
  const handleSubmitScore = async () => {
    try {
      const userId = localStorage.getItem('userId');
      const userEmail = localStorage.getItem('userEmail');
      
      const response = await api.post('/game/submit', {
        user_id: userId,
        game_id: 'memory',
        level: 1,
        score: score,
        user_email: userEmail,
        seed: `memory_seed_${Date.now()}`,
        moves: [],
      });
      
      if (response.data) {
        setSaveMessage({ type: 'success', text: '✅ Рекорд сохранён!' });
        setTimeout(() => setSaveMessage(null), 3000);
      }
    } catch (err) {
      setSaveMessage({ type: 'error', text: '❌ Ошибка при сохранении' });
      setTimeout(() => setSaveMessage(null), 3000);
    }
  };
  
  if (skinsLoading) {
    return (
      <div className="memory-game-container">
        <div className="memory-game-header">
          <div className="memory-stats">
            <div className="memory-stat">
              <span className="stat-label">🎴 Загрузка...</span>
            </div>
          </div>
        </div>
      </div>
    );
  }
  
  return (
    <div className="memory-game-container">
      {saveMessage && (
        <div className={`memory-toast memory-toast--${saveMessage.type}`}>
          {saveMessage.text}
        </div>
      )}
      
      <div className="memory-game-header">
        <div className="memory-stats">
          <div className="memory-stat">
            <span className="stat-label">🎴 Пары</span>
            <span className="stat-value">{matchedPairs}/{totalPairs}</span>
          </div>
          <div className="memory-stat">
            <span className="stat-label">🖱️ Ходы</span>
            <span className="stat-value">{moves}</span>
          </div>
          <div className="memory-stat memory-stat--combo">
            <span className="stat-label">⚡ Комбо</span>
            <span className="stat-value">
              {combo > 0 ? `x${multiplier} (${combo})` : '—'}
            </span>
          </div>
          <div className="memory-stat memory-stat--score">
            <span className="stat-label">🏆 Очки</span>
            <span className="stat-value">{score}</span>
          </div>
          
          {/* Кнопка переключения скина */}
          {skins.memory.hasAnimalCards && (
            <button 
              className={`memory-skin-btn ${useAnimalCards ? 'active' : ''}`}
              onClick={toggleAnimalCards}
              title="Карточки со зверями"
            >
              {useAnimalCards ? '🐾' : '🍎'} Карточки со зверями
            </button>
          )}
        </div>
        
        <div className="memory-buttons">
          <button onClick={handleNewGame} className="memory-btn memory-btn--new">
            🔄 Новая игра
          </button>
          <button onClick={handleBack} className="memory-btn memory-btn--back">
            ← На главную
          </button>
        </div>
      </div>
      
      <div className="memory-board-wrapper">
        <MemoryBoard skin={useAnimalCards && skins.memory.hasAnimalCards ? 'animals' : 'default'} />
      </div>
      
      {gameOver && (
        <div className="memory-game-over">
          <div className="memory-game-over__content">
            <h2>🎉 Победа! 🎉</h2>
            <p>Вы нашли все {totalPairs} пар за {moves} ходов</p>
            <p className="memory-game-over__score">Ваши очки: {score}</p>
            <div className="memory-game-over__formula">
              📖 Формула: 1000 - (лишние ходы × 20), минимум 100
            </div>
            <div className="memory-game-over__buttons">
              <button onClick={handleSubmitScore} className="memory-btn memory-btn--submit">
                📤 Сохранить рекорд
              </button>
              <button onClick={handleNewGame} className="memory-btn memory-btn--new">
                🔄 Сыграть ещё
              </button>
            </div>
          </div>
        </div>
      )}
      
      <div className="memory-rules">
        <details>
          <summary>📖 Как считаются очки?</summary>
          <p>✅ Идеально: 15 ходов → 1000 очков</p>
          <p>➖ Каждый лишний ход: -20 очков</p>
          <p>🛡️ Минимум: 100 очков</p>
          <p>⚡ Комбо: 2 пары подряд → x2, 4+ пар подряд → x3</p>
          <p>🎯 Совет: запоминайте, где лежат парные карты!</p>
        </details>
      </div>
    </div>
  );
}
