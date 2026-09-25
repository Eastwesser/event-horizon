// frontend/src/components/Games/Towers/TowerGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTowerStore } from '../../../store/towerStore';
import { useSkins } from '../../../hooks/useSkins';
import { Balance } from '../../Billing/Balance';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Spinner } from '../../ui/Spinner';
import Notification from '../../Common/Notification/Notification';
import { cn } from '../../../lib/cn';

export function TowerGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const { skins, loading: skinsLoading } = useSkins();
  const [useRainbowBlocks, setUseRainbowBlocks] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

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

  // Загружаем настройки скинов из localStorage
  useEffect(() => {
    const saved = localStorage.getItem('towers_rainbow_blocks');
    if (saved !== null) setUseRainbowBlocks(saved === 'true');
  }, []);

  // Сохраняем настройки скинов
  const toggleRainbowBlocks = () => {
    const newVal = !useRainbowBlocks;
    setUseRainbowBlocks(newVal);
    localStorage.setItem('towers_rainbow_blocks', String(newVal));
  };

  // Проверка авторизации
  useEffect(() => {
    if (!token) {
      navigate('/login');
    } else {
      startGame();
    }
  }, [token, navigate, startGame]);

  // SPACE / ArrowUp — window-level (works even before canvas mounts)
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        if (!gameOver) dropBlock();
      }
    };
    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [gameOver, dropBlock]);

  const handleCanvasClick = () => {
    if (!gameOver) dropBlock();
  };

  // Ручное сохранение рекорда
  const handleManualSave = async () => {
    await submitScore();
    setSaveMessage({ type: 'success', text: '✅ Рекорд сохранён!' });
  };

  const handleResetGame = () => {
    startGame();
  };

  const handleBack = () => {
    navigate('/');
  };

  // Получение цвета блока с учетом скина
  const getBlockColor = (blockLevel: number) => {
    if (useRainbowBlocks && skins.towers.hasRainbowBlocks) {
      const rainbowColors = [
        '#FF6B6B', // Красный
        '#FF9F43', // Оранжевый
        '#FFD700', // Жёлтый
        '#4ADE80', // Зелёный
        '#60A5FA', // Голубой
        '#818CF8', // Синий
        '#C084FC', // Фиолетовый
      ];
      return rainbowColors[(blockLevel - 1) % rainbowColors.length];
    }

    // Стандартные цвета
    const colors = ['#E74C3C', '#C0392B', '#A93226', '#922B21', '#7B241C', '#641E16', '#4A1A0A'];

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
      const blockY = startY - i * blockHeight;

      const color = getBlockColor(i + 1);

      // Градиент для объёма
      const gradient = ctx.createLinearGradient(blockX, blockY, blockX + blockW, blockY);
      gradient.addColorStop(0, color);
      gradient.addColorStop(1, color + 'aa');

      ctx.fillStyle = gradient;
      ctx.fillRect(blockX, blockY, blockW, blockHeight - 2);

      // Обводка
      ctx.strokeStyle = '#ffffffaa';
      ctx.strokeRect(blockX, blockY, blockW, blockHeight - 2);

      // Текстура (линии)
      ctx.fillStyle = '#ffffff33';
      for (let j = 0; j < 3; j++) {
        ctx.fillRect(blockX + 5 + (j * (blockW - 10)) / 3, blockY + 5, 2, blockHeight - 12);
      }
    }

    // Рисуем текущий движущийся блок
    const currentY = startY - towerBlocks.length * blockHeight;
    const currentColor = getBlockColor(towerBlocks.length + 1);
    const gradientCurrent = ctx.createLinearGradient(currentBlockX, currentY, currentBlockX + blockWidth, currentY);
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

    // Game Over экран — warm/neutral game state, not an error red
    if (gameOver) {
      ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
      ctx.fillRect(0, 0, GAME_WIDTH, GAME_HEIGHT);

      ctx.font = 'bold 28px "Space Grotesk", system-ui, sans-serif';
      ctx.fillStyle = '#E8D5A3';
      ctx.textAlign = 'center';
      ctx.fillText('GAME OVER', GAME_WIDTH / 2, GAME_HEIGHT / 2 - 40);

      ctx.font = '20px "Space Grotesk", system-ui, sans-serif';
      ctx.fillStyle = '#FFD700';
      ctx.fillText(`Счёт: ${score}`, GAME_WIDTH / 2, GAME_HEIGHT / 2 + 20);

      ctx.font = '14px "Space Grotesk", system-ui, sans-serif';
      ctx.fillStyle = '#c8c4b8';
      ctx.fillText('Нажмите "Новая игра"', GAME_WIDTH / 2, GAME_HEIGHT / 2 + 70);
      ctx.textAlign = 'start';
    }
  }, [towerBlocks, currentBlockX, blockWidth, score, gameOver, GAME_WIDTH, GAME_HEIGHT, skins, useRainbowBlocks]);

  // Получение множителя для отображения
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
              {useRainbowBlocks ? '🌈' : '🧱'} Радужные блоки
            </button>
          )}
        </>
      }
      controls={
        <>
          <Button variant="primary" size="sm" onClick={handleResetGame}>
            🔄 Новая игра
          </Button>
          <Button variant="secondary" size="sm" onClick={handleManualSave}>
            💾 Сохранить рекорд
          </Button>
        </>
      }
      help={
        <>
          <p>🏗️ Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы положить блок на башню</p>
          <p>🎯 Чем точнее попадание, тем шире будет следующий блок</p>
          <p>⚡ 3 блока подряд = x2, 4 = x3, 5+ = x{combo >= 5 ? combo - 2 : 'N'} множитель очков</p>
          <p>🏆 Очки: 10 × уровень × множитель</p>
          <p>💡 Башня сужается при неточном попадании!</p>
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
      <canvas
        ref={canvasRef}
        width={GAME_WIDTH}
        height={GAME_HEIGHT}
        onClick={handleCanvasClick}
        className="mx-auto block h-auto w-full max-w-[400px] shrink-0 cursor-pointer rounded-md shadow-elevated"
      />
    </GameShell>
  );
}
