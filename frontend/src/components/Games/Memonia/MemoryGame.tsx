// frontend/src/components/Games/Memonia/MemoryGame.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMemoryStore } from '../../../store/memoryStore';
import { useSkins } from '../../../hooks/useSkins';
import { MemoryBoard } from './MemoryBoard';
import { Modal } from '../../ui/Modal';
import { Button } from '../../ui/Button';
import Notification from '../../Common/Notification/Notification';
import api from '../../../services/api';
import './memory.css';

function pluralMoves(n: number): string {
  const abs = Math.abs(n) % 100;
  const d = abs % 10;
  if (abs > 10 && abs < 20) return `${n} ходов`;
  if (d === 1) return `${n} ход`;
  if (d >= 2 && d <= 4) return `${n} хода`;
  return `${n} ходов`;
}

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
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [scoreSaved, setScoreSaved] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem('memory_animal_cards');
    if (saved !== null) setUseAnimalCards(saved === 'true');
  }, []);

  const toggleAnimalCards = () => {
    const newVal = !useAnimalCards;
    setUseAnimalCards(newVal);
    localStorage.setItem('memory_animal_cards', String(newVal));
    resetGame();
    setScoreSaved(false);
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
    setScoreSaved(false);
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
        setScoreSaved(true);
        setSaveMessage({ type: 'success', text: 'Рекорд сохранён' });
        setTimeout(() => setSaveMessage(null), 3000);
        void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
      }
    } catch {
      setSaveMessage({ type: 'error', text: 'Ошибка при сохранении' });
      setTimeout(() => setSaveMessage(null), 3000);
    }
  };

  if (skinsLoading) {
    return (
      <div className="memory-game-container">
        <div className="memory-game-header">
          <div className="memory-stats">
            <div className="memory-stat">
              <span className="stat-label">Загрузка...</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="memory-game-container">
      {saveMessage && (
        <Notification
          type={saveMessage.type}
          message={saveMessage.text}
          onClose={() => setSaveMessage(null)}
        />
      )}

      <div className="memory-game-header">
        <h1 className="memory-game-title">Memonia</h1>
        <div className="memory-stats">
          <div className="memory-stat">
            <span className="stat-label">Пары</span>
            <span className="stat-value">{matchedPairs}/{totalPairs}</span>
          </div>
          <div className="memory-stat">
            <span className="stat-label">Ходы</span>
            <span className="stat-value">{moves}</span>
          </div>
          <div className="memory-stat memory-stat--combo">
            <span className="stat-label">Комбо</span>
            <span className="stat-value">
              {combo > 0 ? `x${multiplier} (${combo})` : '—'}
            </span>
          </div>
          <div className="memory-stat memory-stat--score">
            <span className="stat-label">Очки</span>
            <span className="stat-value">{score}</span>
          </div>

          {skins.memory.hasAnimalCards && (
            <button
              type="button"
              className={`memory-skin-btn ${useAnimalCards ? 'active' : ''}`}
              onClick={toggleAnimalCards}
              title={useAnimalCards ? 'Вернуть фрукты' : 'Показать зверей'}
            >
              {useAnimalCards ? 'Карточки со зверями' : 'Карточки с фруктами'}
            </button>
          )}
        </div>

        <div className="memory-buttons">
          <button type="button" onClick={handleNewGame} className="memory-btn memory-btn--new">
            Новая игра
          </button>
          <button type="button" onClick={handleBack} className="memory-btn memory-btn--back">
            На главную
          </button>
        </div>
      </div>

      <div className="memory-board-wrapper">
        <MemoryBoard skin={useAnimalCards && skins.memory.hasAnimalCards ? 'animals' : 'default'} />
      </div>

      <Modal
        open={gameOver}
        onClose={() => {
          /* dismiss via actions below */
        }}
        title="Победа"
      >
        <p className="text-text-secondary">
          Вы нашли все {totalPairs} пар за {pluralMoves(moves)}
        </p>

        <div className="mt-5 text-center">
          <p className="text-xs font-medium uppercase tracking-wide text-text-muted">Очки</p>
          <p className="font-hud text-3xl font-bold text-horizon-gold">{score}</p>
        </div>

        <p className="mt-4 text-sm text-text-muted">
          Формула: 1000 − (лишние ходы × 20), минимум 100
        </p>

        <div className="mt-6 flex flex-wrap gap-3">
          <Button variant="primary" size="sm" onClick={handleSubmitScore} disabled={scoreSaved}>
            {scoreSaved ? 'Сохранено' : 'Сохранить рекорд'}
          </Button>
          <Button variant="secondary" size="sm" onClick={handleNewGame}>
            Сыграть ещё
          </Button>
          <Button variant="ghost" size="sm" onClick={handleBack}>
            На главную
          </Button>
        </div>
      </Modal>

      <div className="memory-rules">
        <details>
          <summary>Как считаются очки?</summary>
          <p>Идеально: 15 ходов → 1000 очков</p>
          <p>Каждый лишний ход: −20 очков</p>
          <p>Минимум: 100 очков</p>
          <p>Комбо: 2 пары подряд → x2, 4+ пар подряд → x3</p>
          <p>Совет: запоминайте, где лежат парные карты!</p>
        </details>
      </div>
    </div>
  );
}
