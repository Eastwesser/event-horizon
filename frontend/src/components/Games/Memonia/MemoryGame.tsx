// frontend/src/components/Games/Memonia/MemoryGame.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMemoryStore } from './memoryStore';
import { useSkins } from '../../../hooks/useSkins';
import { useGameBoost } from '../../../hooks/useGameBoost';
import { MemoryBoard } from './MemoryBoard';
import { MemoryCard } from './MemoryCard';
import { Modal } from '../../ui/Modal';
import { Icon } from '../../ui/Icon';
import Notification from '../../Common/Notification/Notification';
import { Balance } from '../../Billing/Balance';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';
import { GameOverActions } from '../GameOverActions';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Spinner } from '../../ui/Spinner';
import { GameHelp } from '../../../lib/GameHelp';
import { cn } from '../../../lib/cn';
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
      <div className="flex min-h-screen items-center justify-center bg-void">
        <Spinner size={56} />
      </div>
    );
  }

  return (
    <GameShell
      title="Memonia"
      onBack={handleBack}
      gameId="memory"
      width="wide"
      actions={<Balance />}
      stats={
        <>
          <ScoreChip label="Пары" value={`${matchedPairs}/${totalPairs}`} />
          <ScoreChip label="Ходы" value={moves} />
          <ScoreChip label="Очки" value={score} className="[&_span:last-child]:text-horizon-gold" />
        </>
      }
      controls={
        <>
          <Button size="sm" variant="primary" onClick={handleNewGame} disabled={boostBusy}>
            {boostBusy ? 'Старт…' : runReady ? 'Новая игра' : 'Старт'}
          </Button>
          {skins.memory.hasAnimalCards && (
            <button
              type="button"
              className={cn(
                'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                useAnimalCards
                  ? 'border-photon-cyan/50 bg-photon-cyan/15 text-photon-cyan'
                  : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
              )}
              onClick={toggleAnimalCards}
            >
              {useAnimalCards ? 'Карточки со зверями' : 'Карточки с фруктами'}
            </button>
          )}
          <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} gameId={''} />
        </>
      }
      help={<GameHelp gameId="memory" />}
    >
      {saveMessage && (
        <Notification
          type={saveMessage.type}
          message={saveMessage.text}
          onClose={() => setSaveMessage(null)}
        />
      )}

      <div className="memory-board-wrapper">
        {runReady ? (
          <MemoryBoard skin={useAnimalCards && skins.memory.hasAnimalCards ? 'animals' : 'default'} />
        ) : (
          <div className="memory-board" aria-hidden>
            {Array.from({ length: 30 }, (_, i) => (
              <MemoryCard
                key={`preview-${i}`}
                emoji=""
                flipped={false}
                matched={false}
                onClick={() => {}}
                disabled
              />
            ))}
          </div>
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
    </GameShell>
  );
}
