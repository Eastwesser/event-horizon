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
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Modal } from '../../ui/Modal';
import { Spinner } from '../../ui/Spinner';
import { cn } from '../../../lib/cn';

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
    setGameOver,
  } = useGameStore();

  useEffect(() => {
    const saved = localStorage.getItem('hexagon_space_pancakes');
    if (saved !== null) setUseSpacePancakes(saved === 'true');
  }, []);

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

  const handleDrop = (item: { id: number }, coord: { q: number; r: number }) => {
    addPancakeToHex(item.id, coord);
  };

  const handleEndGame = () => {
    if (confirm('Завершить игру? Ваш прогресс будет сохранён.')) {
      setGameOver(score);
    }
  };

  const handleBack = () => navigate('/');

  const spaceActive = useSpacePancakes && skins.hexagon.hasSpacePancakes;

  if (skinsLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-void">
        <Spinner size={56} />
      </div>
    );
  }

  return (
    <DndProvider backend={HTML5Backend}>
      <GameShell
        title="Pancaker"
        onBack={handleBack}
        actions={<Balance />}
        width="wide"
        stats={
          <>
            <ScoreChip
              label="Счёт"
              value={score}
              className="[&_span:last-child]:text-horizon-gold"
            />
            <ScoreChip
              label="Уровень"
              value={level}
              className="[&_span:last-child]:text-photon-cyan"
            />
            {skins.hexagon.hasSpacePancakes && (
              <button
                type="button"
                onClick={toggleSpacePancakes}
                title="Космические блины"
                className={cn(
                  'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                  spaceActive
                    ? 'border-photon-cyan/50 bg-photon-cyan/15 text-photon-cyan'
                    : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
                )}
              >
                Космические блины
              </button>
            )}
          </>
        }
        controls={
          <>
            <Leaderboard gameId="hexagon" />
            <Button variant="danger" size="sm" onClick={handleEndGame}>
              Завершить
            </Button>
          </>
        }
      >
        <div className="flex min-h-0 w-full flex-1 flex-col items-center justify-center gap-1 overflow-hidden">
          <HexGrid
            tiles={tiles}
            onDrop={handleDrop}
            skinMode={spaceActive ? 'space' : 'default'}
          />
          <Tray stacks={tray} skinMode={spaceActive ? 'space' : 'default'} />
        </div>
      </GameShell>

      <Modal open={isGameOver} onClose={() => {}} title="Игра окончена">
        <p className="text-text-secondary">Вы испекли {finalScore} блинов!</p>
        <p className="mt-1 text-text-secondary">Стопка блинов пополнилась.</p>

        <div className="mt-5 text-center">
          <p className="text-xs font-medium uppercase tracking-wide text-text-muted">Очки</p>
          <p className="font-hud text-3xl font-bold text-horizon-gold">{finalScore}</p>
        </div>

        <div className="mt-6 flex flex-wrap gap-3">
          <Button variant="primary" size="sm" onClick={() => initGame()}>
            Новая игра
          </Button>
          <Button variant="ghost" size="sm" onClick={handleBack}>
            На главную
          </Button>
        </div>
      </Modal>
    </DndProvider>
  );
}
