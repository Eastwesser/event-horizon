// frontend/src/components/Games/Flappy/FlappyGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFlappyStore } from '../../../store/flappyStore';
import { useSkins } from '../../../hooks/useSkins';
import api from '../../../services/api';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Spinner } from '../../ui/Spinner';
import { Modal } from '../../ui/Modal';
import Notification from '../../Common/Notification/Notification';
import { Balance, invalidateBalanceCache } from '../../Billing/Balance';
import { cn } from '../../../lib/cn';
import {
  FLAPPY_H,
  FLAPPY_W,
  drawBird,
  drawClouds,
  drawPipe,
  drawScore,
  drawSky,
  drawStars,
  drawStartHint,
} from './flappyDraw';
import './FlappyGame.css';

const BOOST_COST = 10;
const BIRD_SIZE = 30;

function IconBird({ golden }: { golden: boolean }) {
  return (
    <svg className="eh-flappy-skin-icon" viewBox="0 0 16 16" aria-hidden="true">
      <ellipse cx="7" cy="8" rx="5" ry="4.5" fill={golden ? '#FFD700' : '#4A90D9'} />
      <circle cx="10" cy="6.5" r="1.2" fill="#111" />
      <path d="M11.5 8 L15 8.5 L11.5 9.2 Z" fill="#E53935" />
    </svg>
  );
}

function IconPipes({ rainbow }: { rainbow: boolean }) {
  return (
    <svg className="eh-flappy-skin-icon" viewBox="0 0 16 16" aria-hidden="true">
      <defs>
        <linearGradient id="eh-pipe-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={rainbow ? '#FF6B6B' : '#2E7D32'} />
          <stop offset="100%" stopColor={rainbow ? '#818CF8' : '#1B5E20'} />
        </linearGradient>
      </defs>
      <rect x="4" y="1" width="8" height="6" rx="1" fill="url(#eh-pipe-grad)" />
      <rect x="3" y="6" width="10" height="2" fill="url(#eh-pipe-grad)" />
      <rect x="3" y="9" width="10" height="2" fill="url(#eh-pipe-grad)" />
      <rect x="4" y="11" width="8" height="4" rx="1" fill="url(#eh-pipe-grad)" />
    </svg>
  );
}

export function FlappyGame() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const cloudScrollRef = useRef(0);
  const starScrollRef = useRef(0);
  const prevGameOver = useRef(false);
  const { skins, loading: skinsLoading } = useSkins();

  const [useRainbowPipes, setUseRainbowPipes] = useState(false);
  const [useGoldenBird, setUseGoldenBird] = useState(false);
  const [useBoost, setUseBoost] = useState(false);
  const [boostBusy, setBoostBusy] = useState(false);
  const [selectedLevel, setSelectedLevel] = useState(1);
  const [flapPulse, setFlapPulse] = useState(false);
  const [shake, setShake] = useState(false);

  const {
    birdY,
    birdVelocity,
    pipes,
    score,
    gameOver,
    started,
    jump,
    resetGame,
    startGame,
    setLevel,
    boosted,
    level,
    PIPE_SPEED,
    lastSubmitRanked,
    lastSubmitMessage,
  } = useFlappyStore();

  // Загружаем настройки скинов / уровня из localStorage
  useEffect(() => {
    const savedPipes = localStorage.getItem('flappy_rainbow_pipes');
    const savedBird = localStorage.getItem('flappy_golden_bird');
    if (savedPipes !== null) setUseRainbowPipes(savedPipes === 'true');
    if (savedBird !== null) setUseGoldenBird(savedBird === 'true');
    const savedLevel = parseInt(localStorage.getItem('flappy_level') || '1', 10);
    const lv = Number.isFinite(savedLevel) ? Math.min(10, Math.max(1, savedLevel)) : 1;
    setSelectedLevel(lv);
    setLevel(lv);
  }, [setLevel]);

  // Сохраняем настройки скинов
  const toggleRainbowPipes = () => {
    const newVal = !useRainbowPipes;
    setUseRainbowPipes(newVal);
    localStorage.setItem('flappy_rainbow_pipes', String(newVal));
  };

  const toggleGoldenBird = () => {
    const newVal = !useGoldenBird;
    setUseGoldenBird(newVal);
    localStorage.setItem('flappy_golden_bird', String(newVal));
  };

  // Проверка авторизации
  useEffect(() => {
    if (!token) {
      navigate('/login');
    }
  }, [token, navigate]);

  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const beginRun = async () => {
    if (boostBusy || started) return;
    const lv = Math.min(10, Math.max(1, selectedLevel));
    localStorage.setItem('flappy_level', String(lv));
    setLevel(lv);
    if (!useBoost) {
      startGame({ boosted: false, boostId: null, level: lv });
      return;
    }
    setBoostBusy(true);
    try {
      const response = await api.post('/game/boost/start', { game_id: 'flappy' });
      const boostId = response.data?.boost_id as string | undefined;
      if (!boostId) {
        throw new Error(response.data?.message || 'boost_id missing');
      }
      invalidateBalanceCache();
      startGame({ boosted: true, boostId, level: lv });
    } catch (e: any) {
      const msg =
        e?.response?.data?.error ||
        e?.response?.data?.message ||
        e?.message ||
        'Недостаточно лампочек для boost';
      setSaveMessage({ type: 'error', text: String(msg) });
    } finally {
      setBoostBusy(false);
    }
  };

  // Обработка кликов и пробела для прыжка / старта
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        const { started: isStarted, gameOver: isOver } = useFlappyStore.getState();
        if (!isStarted && !isOver) {
          void beginRun();
          return;
        }
        if (isOver) return;
        jump();
        pulseFlap();
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [jump, useBoost, boostBusy, started]);

  useEffect(() => {
    if (lastSubmitMessage == null || lastSubmitRanked == null) return;
    if (lastSubmitRanked === false) {
      setSaveMessage({
        type: 'success',
        text: 'Забег с boost — не попал в лидерборд',
      });
    }
  }, [lastSubmitMessage, lastSubmitRanked]);

  const handleManualSave = async () => {
    const state = useFlappyStore.getState();
    const { score, boostId, boosted: runBoosted, level: runLevel } = state;
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const token = localStorage.getItem('accessToken');

    try {
      const body: Record<string, unknown> = {
        user_id: userId,
        game_id: 'flappy',
        level: runLevel || selectedLevel || 1,
        score: score,
        user_email: userEmail,
        seed: `flappy_manual_${Date.now()}`,
        moves: [],
      };
      if (runBoosted && boostId) {
        body.boost_id = boostId;
      }
      const response = await api.post('/game/submit', body, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (response.status === 200) {
        const ranked =
          response.data?.ranked === true &&
          !runBoosted &&
          !String(response.data?.message || '').includes('not ranked');
        if (ranked) {
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
          setSaveMessage({ type: 'success', text: 'Рекорд сохранён!' });
          void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
        } else {
          setSaveMessage({ type: 'success', text: 'Забег с boost — не попал в лидерборд' });
        }
      }
    } catch (err) {
      setSaveMessage({ type: 'error', text: 'Ошибка сохранения' });
    }
  };

  const pulseFlap = () => {
    setFlapPulse(true);
    window.setTimeout(() => setFlapPulse(false), 90);
  };

  // Shake once when run ends
  useEffect(() => {
    if (gameOver && !prevGameOver.current) {
      setShake(true);
      const t = window.setTimeout(() => setShake(false), 420);
      prevGameOver.current = true;
      return () => window.clearTimeout(t);
    }
    if (!gameOver) prevGameOver.current = false;
  }, [gameOver]);

  // Draw loop (cosmetic) — physics stay in flappyStore
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    if (started && !gameOver) {
      cloudScrollRef.current += PIPE_SPEED * 0.3;
      starScrollRef.current += PIPE_SPEED * 0.08;
    }

    ctx.clearRect(0, 0, FLAPPY_W, FLAPPY_H);
    drawSky(ctx, FLAPPY_W, FLAPPY_H);
    drawStars(ctx, FLAPPY_W, FLAPPY_H, starScrollRef.current);
    drawClouds(ctx, FLAPPY_W, cloudScrollRef.current);

    const rainbow = useRainbowPipes && skins.flappy.hasRainbowPipes;
    const golden = useGoldenBird && skins.flappy.hasGoldenBird;

    for (const pipe of pipes) {
      drawPipe(ctx, pipe, 60, FLAPPY_H, rainbow);
    }

    drawBird(ctx, birdY, BIRD_SIZE, birdVelocity, golden);
    drawScore(ctx, score, FLAPPY_W);

    if (!started && !gameOver) {
      drawStartHint(ctx, FLAPPY_W, FLAPPY_H);
    }
  }, [
    birdY,
    birdVelocity,
    pipes,
    score,
    gameOver,
    started,
    skins,
    useRainbowPipes,
    useGoldenBird,
    PIPE_SPEED,
  ]);

  const handleCanvasClick = () => {
    const { started: isStarted, gameOver: isOver } = useFlappyStore.getState();
    if (!isStarted && !isOver) {
      void beginRun();
      return;
    }
    if (isOver) return;
    jump();
    pulseFlap();
  };

  const handleBack = () => {
    navigate('/');
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
      title="Flappy Bird"
      onBack={handleBack}
      width="wide"
      actions={<Balance />}
      stats={
        <>
          <ScoreChip label="Счёт" value={score} className="[&_span:last-child]:text-horizon-gold" />
          <ScoreChip label="Уровень" value={started || gameOver ? level : selectedLevel} />
          {boosted && (
            <span className="rounded-sm border border-horizon-gold/40 bg-horizon-gold/10 px-3 py-1.5 text-sm text-horizon-gold">
              Boost 5с
            </span>
          )}
          {skins.flappy.hasGoldenBird && (
            <button
              type="button"
              onClick={toggleGoldenBird}
              title="Золотая птичка"
              className={cn(
                'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                useGoldenBird
                  ? 'border-indigo/50 bg-indigo/15 text-indigo-soft'
                  : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
              )}
            >
              <IconBird golden={useGoldenBird} />
              Птичка
            </button>
          )}
          {skins.flappy.hasRainbowPipes && (
            <button
              type="button"
              onClick={toggleRainbowPipes}
              title="Радужные трубы"
              className={cn(
                'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                useRainbowPipes
                  ? 'border-photon-cyan/50 bg-photon-cyan/15 text-photon-cyan'
                  : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
              )}
            >
              <IconPipes rainbow={useRainbowPipes} />
              Трубы
            </button>
          )}
        </>
      }
      controls={
        <>
          {!started && (
            <label className="flex items-center gap-2 text-sm text-text-primary">
              Уровень
              <select
                className="rounded-sm border border-white/15 bg-void px-2 py-1 text-text-primary"
                value={selectedLevel}
                disabled={boostBusy}
                onChange={(e) => {
                  const lv = Math.min(10, Math.max(1, parseInt(e.target.value, 10) || 1));
                  setSelectedLevel(lv);
                  setLevel(lv);
                  localStorage.setItem('flappy_level', String(lv));
                }}
              >
                {Array.from({ length: 10 }, (_, i) => i + 1).map((lv) => (
                  <option key={lv} value={lv}>
                    {lv}
                  </option>
                ))}
              </select>
            </label>
          )}
          {!started && (
            <label className="flex max-w-md cursor-pointer flex-col gap-1 text-sm text-text-secondary">
              <span className="inline-flex items-center gap-2 text-text-primary">
                <input
                  type="checkbox"
                  checked={useBoost}
                  disabled={boostBusy}
                  onChange={(e) => setUseBoost(e.target.checked)}
                />
                Использовать boost ({BOOST_COST} лампочек)
              </span>
              {useBoost && (
                <span className="text-xs text-horizon-gold/90">
                  Этот забег не попадёт в лидерборд — boost считается нечестным преимуществом
                </span>
              )}
            </label>
          )}
          <Button
            variant="primary"
            size="sm"
            onClick={() => {
              cloudScrollRef.current = 0;
              starScrollRef.current = 0;
              resetGame();
            }}
            disabled={boostBusy}
          >
            Новая игра
          </Button>
          {!started && (
            <Button variant="secondary" size="sm" onClick={() => void beginRun()} disabled={boostBusy}>
              {boostBusy ? 'Старт…' : 'Старт'}
            </Button>
          )}
          <Button variant="secondary" size="sm" onClick={handleManualSave} disabled={!gameOver && !started}>
            Сохранить рекорд
          </Button>
        </>
      }
      help={
        <>
          <p>Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы птичка летела вверх</p>
          <p>Не врезайтесь в трубы и не падайте на землю</p>
          <p>Уровни 1–10: выше уровень — уже щель и быстрее трубы</p>
          <p>Boost (10 лампочек): 5 сек slow-mo в начале; забег не в лидерборд</p>
          <p>Лидерборд отдельный на каждый уровень</p>
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
          'eh-flappy-stage',
          flapPulse && 'eh-flappy-stage--flap',
          shake && 'eh-flappy-stage--shake',
        )}
      >
        <canvas
          ref={canvasRef}
          width={FLAPPY_W}
          height={FLAPPY_H}
          onClick={handleCanvasClick}
        />
      </div>

      <Modal
        open={gameOver}
        onClose={() => {
          /* keep run state; dismiss overlay only via buttons below if needed */
        }}
        title="Game Over"
      >
        <p className="text-lg text-text-primary">
          Счёт: <span className="font-display font-semibold text-horizon-gold">{score}</span>
        </p>
        <p className="mt-1 text-sm text-text-secondary">Уровень {level}</p>
        {boosted && (
          <p className="mt-3 text-sm text-horizon-gold">
            Забег с boost — не попал в лидерборд
          </p>
        )}
        {!boosted && lastSubmitRanked === true && (
          <p className="mt-3 text-sm text-photon-cyan">Рекорд отправлен в лидерборд</p>
        )}
        {!boosted && lastSubmitRanked === false && lastSubmitMessage && (
          <p className="mt-3 text-sm text-text-secondary">{lastSubmitMessage}</p>
        )}
        <div className="mt-6 flex flex-wrap gap-3">
          <Button
            variant="primary"
            size="sm"
            onClick={() => {
              cloudScrollRef.current = 0;
              starScrollRef.current = 0;
              resetGame();
            }}
          >
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
