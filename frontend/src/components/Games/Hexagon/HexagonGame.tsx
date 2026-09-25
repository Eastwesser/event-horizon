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
    if (confirm('Завершить игру? Ваш прогресс будет сохранён.')) {
      setGameOver(score);
    }
  };

  const handleBack = () => navigate('/');

  const activeIcon = useSpacePancakes && skins.hexagon.hasSpacePancakes ? '🌌' : '🥞';

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
            <ScoreChip label="Счёт" value={`${activeIcon} ${score}`} className="[&_span:last-child]:text-horizon-gold" />
            <ScoreChip label="Уровень" value={level} className="[&_span:last-child]:text-photon-cyan" />
            {skins.hexagon.hasSpacePancakes && (
              <button
                onClick={toggleSpacePancakes}
                title="Космические блины"
                className={cn(
                  'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                  useSpacePancakes
                    ? 'border-photon-cyan/50 bg-photon-cyan/15 text-photon-cyan'
                    : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
                )}
              >
                {useSpacePancakes ? '🌌' : '🥞'} Космические блины
              </button>
            )}
          </>
        }
        controls={
          <>
            <Leaderboard gameId="hexagon" />
            <Button variant="danger" size="sm" onClick={handleEndGame}>
              ⏹️ Завершить
            </Button>
          </>
        }
      >
        <div className="flex min-h-0 w-full flex-1 flex-col items-center justify-center gap-1 overflow-hidden">
          <HexGrid
            tiles={tiles}
            onDrop={handleDrop}
            skinMode={useSpacePancakes && skins.hexagon.hasSpacePancakes ? 'space' : 'default'}
          />
          <Tray
            stacks={tray}
            skinMode={useSpacePancakes && skins.hexagon.hasSpacePancakes ? 'space' : 'default'}
          />
        </div>
      </GameShell>

      <Modal open={isGameOver} onClose={() => {}} title={`${activeIcon} Игра окончена! ${activeIcon}`}>
        <p className="text-text-secondary">Вы испекли {finalScore} блинов!</p>
        <p className="mt-1 text-text-secondary">Стопка блинов пополнилась! 🎉</p>
        <div className="mt-6 flex flex-wrap gap-3">
          <Button variant="primary" onClick={() => initGame()}>
            🔄 Новая игра
          </Button>
          <Button variant="ghost" onClick={handleBack}>
            🏠 На главную
          </Button>
        </div>
      </Modal>
    </DndProvider>
  );
}
