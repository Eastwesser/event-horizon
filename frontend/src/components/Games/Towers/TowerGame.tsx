// frontend/src/components/Games/Towers/TowerGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTowerStore } from '../../../store/towerStore';
import { useSkins } from '../../../hooks/useSkins';
import { Balance } from '../../Billing/Balance';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Spinner } from '../../ui/Spinner';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';
import { cn } from '../../../lib/cn';
import {
  BLOCK_HEIGHT,
  blockColor,
  drawBackground,
  drawBlock,
  drawDirectionChevron,
  drawMovingBlock,
  stackStartY,
} from './towerDraw';
import './TowerGame.css';

function IconBlocks({ rainbow }: { rainbow: boolean }) {
  return (
    <svg className="eh-tower-skin-icon" viewBox="0 0 16 16" aria-hidden="true">
      <rect x="3" y="9" width="10" height="4" rx="0.5" fill={rainbow ? '#60A5FA' : '#C0392B'} />
      <rect x="4" y="5" width="8" height="4" rx="0.5" fill={rainbow ? '#FFD700' : '#E74C3C'} />
      <rect x="5" y="1" width="6" height="4" rx="0.5" fill={rainbow ? '#FF6B6B' : '#A93226'} />
    </svg>
  );
}

export function TowerGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const prevGameOver = useRef(false);
  const { skins, loading: skinsLoading } = useSkins();
  const [useRainbowBlocks, setUseRainbowBlocks] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(
    null,
  );
  const [dropPulse, setDropPulse] = useState(false);
  const [shake, setShake] = useState(false);

  const {
    towerBlocks,
    currentBlockX,
    blockWidth,
    score,
    level,
    combo,
    gameOver,
    direction,
    GAME_WIDTH,
    GAME_HEIGHT,
    startGame,
    dropBlock,
    submitScore,
  } = useTowerStore();

  useEffect(() => {
    const saved = localStorage.getItem('towers_rainbow_blocks');
    if (saved !== null) setUseRainbowBlocks(saved === 'true');
  }, []);

  const toggleRainbowBlocks = () => {
    const newVal = !useRainbowBlocks;
    setUseRainbowBlocks(newVal);
    localStorage.setItem('towers_rainbow_blocks', String(newVal));
  };

  useEffect(() => {
    if (!token) {
      navigate('/login');
    } else {
      startGame();
    }
  }, [token, navigate, startGame]);

  const pulseDrop = () => {
    setDropPulse(true);
    window.setTimeout(() => setDropPulse(false), 90);
  };

  const tryDrop = () => {
    if (gameOver) return;
    const beforeLen = useTowerStore.getState().towerBlocks.length;
    const beforeOver = useTowerStore.getState().gameOver;
    dropBlock();
    const after = useTowerStore.getState();
    if (!beforeOver && after.towerBlocks.length > beforeLen) {
      pulseDrop();
    }
  };

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        if (gameOver) return;
        const beforeLen = useTowerStore.getState().towerBlocks.length;
        dropBlock();
        const after = useTowerStore.getState();
        if (after.towerBlocks.length > beforeLen) {
          pulseDrop();
        }
      }
    };
    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [gameOver, dropBlock]);

  useEffect(() => {
    if (gameOver && !prevGameOver.current) {
      setShake(true);
      const t = window.setTimeout(() => setShake(false), 420);
      prevGameOver.current = true;
      return () => window.clearTimeout(t);
    }
    if (!gameOver) prevGameOver.current = false;
  }, [gameOver]);

  const handleManualSave = async () => {
    await submitScore();
    setSaveMessage({ type: 'success', text: 'Рекорд сохранён!' });
  };

  const handleResetGame = () => {
    startGame();
  };

  const handleBack = () => {
    navigate('/');
  };

  const rainbow = useRainbowBlocks && skins.towers.hasRainbowBlocks;

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
    drawBackground(ctx, GAME_WIDTH, GAME_HEIGHT);

    const startY = stackStartY(GAME_HEIGHT);
    const h = BLOCK_HEIGHT - 2;

    for (let i = 0; i < towerBlocks.length; i++) {
      const blockW = towerBlocks[i];
      const blockX = (GAME_WIDTH - blockW) / 2;
      const blockY = startY - i * BLOCK_HEIGHT;
      drawBlock(ctx, blockX, blockY, blockW, h, blockColor(i + 1, rainbow));
    }

    if (!gameOver) {
      const currentY = startY - towerBlocks.length * BLOCK_HEIGHT;
      const color = blockColor(towerBlocks.length + 1, rainbow);
      drawMovingBlock(ctx, currentBlockX, currentY, blockWidth, h, color);
      drawDirectionChevron(ctx, direction, GAME_WIDTH, currentY, h);
    }
  }, [
    towerBlocks,
    currentBlockX,
    blockWidth,
    gameOver,
    GAME_WIDTH,
    GAME_HEIGHT,
    rainbow,
    direction,
  ]);

  const getMultiplierDisplay = () => {
    if (combo >= 5) return `x${combo - 2}`;
    if (combo >= 4) return 'x3';
    if (combo >= 3) return 'x2';
    return 'x1';
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
      title="Builder"
      onBack={handleBack}
      actions={<Balance />}
      width="narrow"
      stats={
        <>
          <ScoreChip label="Счёт" value={score} className="[&_span:last-child]:text-horizon-gold" />
          <ScoreChip label="Уровень" value={level} />
          <ScoreChip
            label="Комбо"
            value={getMultiplierDisplay()}
            className="border-photon-cyan/30 [&_span:last-child]:text-photon-cyan"
          />
          <ScoreChip label="Высота" value={towerBlocks.length} />
          {skins.towers.hasRainbowBlocks && (
            <button
              type="button"
              onClick={toggleRainbowBlocks}
              title="Радужные блоки"
              className={cn(
                'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                useRainbowBlocks
                  ? 'border-photon-cyan/50 bg-photon-cyan/15 text-photon-cyan'
                  : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
              )}
            >
              <IconBlocks rainbow={useRainbowBlocks} />
              Радужные блоки
            </button>
          )}
        </>
      }
      controls={
        <>
          <Button variant="primary" size="sm" onClick={handleResetGame}>
            Новая игра
          </Button>
          <Button variant="secondary" size="sm" onClick={handleManualSave}>
            Сохранить рекорд
          </Button>
        </>
      }
      help={
        <>
          <p>Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы положить блок на башню</p>
          <p>Чем точнее попадание, тем шире будет следующий блок</p>
          <p>
            3 блока подряд = x2, 4 = x3, 5+ = x{combo >= 5 ? combo - 2 : 'N'} множитель очков
          </p>
          <p>Очки: 10 × уровень × множитель</p>
          <p>Башня сужается при неточном попадании</p>
        </>
      }
    >
      {saveMessage && (
        <Notification
          type={saveMessage.type}
          message={saveMessage.text}
          onClose={() => setSaveMessage(null)}
        />
      )}
      <div
        className={cn(
          'eh-tower-stage',
          dropPulse && 'eh-tower-stage--drop',
          shake && 'eh-tower-stage--shake',
        )}
      >
        <canvas
          ref={canvasRef}
          width={GAME_WIDTH}
          height={GAME_HEIGHT}
          onClick={tryDrop}
        />
      </div>

      <Modal open={gameOver} onClose={() => {}} title="Game Over">
        <p className="text-lg text-text-primary">
          Счёт: <span className="font-display font-semibold text-horizon-gold">{score}</span>
        </p>
        <p className="mt-1 text-sm text-text-secondary">Уровень {level}</p>
        <p className="mt-1 text-sm text-text-secondary">Высота: {towerBlocks.length}</p>
        <p className="mt-1 text-sm text-photon-cyan">Комбо: {getMultiplierDisplay()}</p>
        <div className="mt-6 flex flex-wrap gap-3">
          <Button variant="primary" size="sm" onClick={handleResetGame}>
            Новая игра
          </Button>
          <Button variant="secondary" size="sm" onClick={handleManualSave}>
            Сохранить рекорд
          </Button>
          <Button variant="ghost" size="sm" onClick={handleBack}>
            На главную
          </Button>
        </div>
      </Modal>
    </GameShell>
  );
}
