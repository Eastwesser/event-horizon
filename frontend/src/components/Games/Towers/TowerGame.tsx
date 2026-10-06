// frontend/src/components/Games/Towers/TowerGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTowerStore } from '../../../store/towerStore';
import { useSkins } from '../../../hooks/useSkins';
import { useGameBoost } from '../../../hooks/useGameBoost';
import { Balance } from '../../Billing/Balance';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Spinner } from '../../ui/Spinner';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';
import { Icon } from '../../ui/Icon';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';
import { GameOverActions } from '../GameOverActions';
import { boostHelpLines } from '../../../lib/gameBoostCopy';
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
  const [selectedDifficulty, setSelectedDifficulty] = useState(1);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(
    null,
  );
  const [shake, setShake] = useState(false);
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('towers');

  const {
    towerBlocks,
    currentBlockX,
    blockWidth,
    score,
    floor,
    difficulty,
    combo,
    gameOver,
    started,
    boosted,
    direction,
    GAME_WIDTH,
    GAME_HEIGHT,
    startGame,
    dropBlock,
    submitScore,
    lastSubmitRanked,
    lastSubmitMessage,
  } = useTowerStore();

  useEffect(() => {
    const saved = localStorage.getItem('towers_rainbow_blocks');
    if (saved !== null) setUseRainbowBlocks(saved === 'true');
    const savedDiff = parseInt(localStorage.getItem('towers_difficulty') || '1', 10);
    if (Number.isFinite(savedDiff)) {
      setSelectedDifficulty(Math.min(10, Math.max(1, savedDiff)));
    }
  }, []);

  const toggleRainbowBlocks = () => {
    const newVal = !useRainbowBlocks;
    setUseRainbowBlocks(newVal);
    localStorage.setItem('towers_rainbow_blocks', String(newVal));
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

  useEffect(() => {
    if (lastSubmitMessage == null || lastSubmitRanked == null) return;
    if (lastSubmitRanked === false) {
      setSaveMessage({ type: 'success', text: boostUnrankedToast() });
    }
  }, [lastSubmitMessage, lastSubmitRanked]);

  const beginRun = async () => {
    if (boostBusy || (started && !gameOver)) return;
    const lv = Math.min(10, Math.max(1, selectedDifficulty));
    localStorage.setItem('towers_difficulty', String(lv));
    try {
      const { boostId, boosted: isBoosted } = await armBoost();
      startGame({ boosted: isBoosted, boostId, difficulty: lv });
    } catch {
      /* boostError handled via effect */
    }
  };

  const tryDrop = () => {
    if (gameOver || !started) return;
    dropBlock();
  };

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        if (!started || gameOver) return;
        dropBlock();
      }
    };
    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [gameOver, started, dropBlock]);

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
    const { lastSubmitRanked: ranked } = useTowerStore.getState();
    setSaveMessage({
      type: 'success',
      text: ranked === false ? boostUnrankedToast() : 'Счёт сохранён · рекорд в лидерборд',
    });
  };

  const handleBack = () => navigate('/#games');
  const midRun = started && !gameOver;

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

    if (started && !gameOver) {
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
    started,
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
          <ScoreChip
            label="Уровень"
            value={started || gameOver ? difficulty : selectedDifficulty}
          />
          <ScoreChip label="Этаж" value={floor} />
          <ScoreChip
            label="Комбо"
            value={getMultiplierDisplay()}
            className="border-photon-cyan/30 [&_span:last-child]:text-photon-cyan"
          />
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
          {!midRun && (
            <label className="flex items-center gap-2 text-sm text-text-primary">
              Уровень
              <select
                className="rounded-sm border border-white/15 bg-void px-2 py-1 text-text-primary"
                value={selectedDifficulty}
                disabled={boostBusy}
                onChange={(e) => {
                  const lv = Math.min(10, Math.max(1, parseInt(e.target.value, 10) || 1));
                  setSelectedDifficulty(lv);
                  localStorage.setItem('towers_difficulty', String(lv));
                }}
              >
                {Array.from({ length: 10 }, (_, i) => i + 1).map((lv) => (
                  <option key={lv} value={lv} className="bg-void text-text-primary">
                    {lv}
                  </option>
                ))}
              </select>
            </label>
          )}
          {midRun ? (
            <Button variant="primary" size="sm" disabled>
              В игре
            </Button>
          ) : (
            <Button variant="primary" size="sm" onClick={() => void beginRun()} disabled={boostBusy}>
              {boostBusy ? 'Старт…' : gameOver ? 'Новая игра' : 'Старт'}
            </Button>
          )}
          <Button variant="secondary" size="sm" onClick={handleManualSave} disabled={!started}>
            Сохранить рекорд
          </Button>
          {!midRun && (
            <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} />
          )}
        </>
      }
      help={
        <>
          <p>Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы положить блок на башню</p>
          <p>Чем точнее попадание, тем шире будет следующий блок. Комбо растёт за точные укладки.</p>
          <p>Уровни 1–10: выше уровень — быстрее блок и уже старт. Этаж — высота башни (не в лидерборд).</p>
          <p>Промах — игра окончена. На низких уровнях есть небольшой запас на ошибку.</p>
          {boostHelpLines('towers').map((line) => (
            <p key={line}>{line}</p>
          ))}
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
      <div className={cn('eh-tower-stage', shake && 'eh-tower-stage--shake')}>
        <canvas ref={canvasRef} width={GAME_WIDTH} height={GAME_HEIGHT} onClick={tryDrop} />
      </div>

      <Modal open={gameOver} onClose={() => {}} title="Game Over">
        <p className="text-lg text-text-primary">
          Счёт: <span className="font-display font-semibold text-horizon-gold">{score}</span>
        </p>
        <p className="mt-1 text-sm text-text-secondary">Уровень {difficulty}</p>
        <p className="mt-1 text-sm text-text-secondary">Этаж: {floor}</p>
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
          onNewGame={() => void beginRun()}
          onHome={handleBack}
          onSave={() => void handleManualSave()}
          saveLabel="Сохранить рекорд"
          busy={boostBusy}
        />
      </Modal>
    </GameShell>
  );
}
