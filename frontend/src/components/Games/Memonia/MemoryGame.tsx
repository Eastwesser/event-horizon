// frontend/src/components/Games/Memonia/MemoryGame.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMemoryStore } from '../../../store/memoryStore';
import { useSkins } from '../../../hooks/useSkins';
import { useGameBoost } from '../../../hooks/useGameBoost';
import { MemoryBoard } from './MemoryBoard';
import { Modal } from '../../ui/Modal';
import { Icon } from '../../ui/Icon';
import Notification from '../../Common/Notification/Notification';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';
import { GameOverActions } from '../GameOverActions';
import { boostHelpLines } from '../../../lib/gameBoostCopy';
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
  const [runReady, setRunReady] = useState(false);
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('memory');

  const {
    moves,
    matchedPairs,
    gameOver,
    score,
    boosted,
    lastSubmitRanked,
    initGame,
    submitScore,
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
  };

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  useEffect(() => {
    if (boostError) {
      setSaveMessage({ type: 'error', text: boostError });
      setBoostError(null);
    }
  }, [boostError, setBoostError]);

  const beginRun = async () => {
    if (boostBusy) return;
    try {
      const { boostId, boosted: isBoosted } = await armBoost();
      initGame({ boosted: isBoosted, boostId });
      setRunReady(true);
      setScoreSaved(false);
    } catch {
      /* handled */
    }
  };

  const handleNewGame = () => {
    void beginRun();
  };

  const handleBack = () => navigate('/#games');

  const handleSubmitScore = async () => {
    await submitScore();
    const ranked = useMemoryStore.getState().lastSubmitRanked;
    setScoreSaved(true);
    setSaveMessage({
      type: 'success',
      text: ranked === false ? boostUnrankedToast() : 'Счёт сохранён · рекорд в лидерборд',
    });
    setTimeout(() => setSaveMessage(null), 3000);
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
            <span className="stat-value">
              {matchedPairs}/{totalPairs}
            </span>
          </div>
          <div className="memory-stat">
            <span className="stat-label">Ходы</span>
            <span className="stat-value">{moves}</span>
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

        <div className="memory-buttons" style={{ flexDirection: 'column', alignItems: 'flex-start', gap: 8 }}>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <button
              type="button"
              onClick={handleNewGame}
              className="memory-btn memory-btn--new"
              disabled={boostBusy}
            >
              {boostBusy ? 'Старт…' : runReady ? 'Новая игра' : 'Старт'}
            </button>
            <button type="button" onClick={handleBack} className="memory-btn memory-btn--back">
              На главную
            </button>
          </div>
          <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} />
        </div>
      </div>

      <div className="memory-board-wrapper">
        {runReady ? (
          <MemoryBoard skin={useAnimalCards && skins.memory.hasAnimalCards ? 'animals' : 'default'} />
        ) : (
          <p className="text-text-secondary p-6">Нажмите Старт — boost по желанию (ниже / в «Как играть»)</p>
        )}
      </div>

      <Modal open={gameOver} onClose={() => {}} title="Победа">
        <p className="text-text-secondary">
          Вы нашли все {totalPairs} пар за {pluralMoves(moves)}
        </p>
        <div className="mt-5 text-center">
          <p className="text-xs font-medium uppercase tracking-wide text-text-muted">Очки</p>
          <p className="font-hud text-3xl font-bold text-horizon-gold">{score}</p>
        </div>
        {boosted && (
          <p className="mt-3 text-sm text-horizon-gold">{boostUnrankedToast()}</p>
        )}
        {!boosted && lastSubmitRanked === true && (
          <p className="mt-3 inline-flex items-center gap-2 text-sm text-success">
            <Icon name="check" className="h-4 w-4" aria-hidden />
            Счёт сохранён · рекорд в лидерборд
          </p>
        )}
        <GameOverActions
          onNewGame={handleNewGame}
          onHome={handleBack}
          onSave={() => {
            if (!scoreSaved) void handleSubmitScore();
          }}
          saveLabel={scoreSaved ? 'Сохранено' : 'Сохранить рекорд'}
          busy={boostBusy}
        />
      </Modal>

      <div className="memory-rules">
        <details>
          <summary>Как играть?</summary>
          <p>Найди все пары карточек за минимум ходов.</p>
          <p>Идеально: 15 ходов → 1000 очков</p>
          {boostHelpLines('memory').map((line) => (
            <p key={line}>{line}</p>
          ))}
        </details>
      </div>
    </div>
  );
}
