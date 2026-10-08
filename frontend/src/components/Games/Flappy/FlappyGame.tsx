// frontend/src/components/Games/Flappy/FlappyGame.tsx
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFlappyStore } from './flappyStore';
import { useSkins } from '../../../hooks/useSkins';
import { useGameBoost } from '../../../hooks/useGameBoost';
import api from '../../../services/api';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import { Spinner } from '../../ui/Spinner';
import { Modal } from '../../ui/Modal';
import { Icon } from '../../ui/Icon';
import Notification from '../../Common/Notification/Notification';
import { Balance } from '../../Billing/Balance';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';
import { GameOverActions } from '../GameOverActions';
import { boostHelpLines } from '../../../lib/gameBoostCopy';
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

function IconPipes({ cosmic }: { cosmic: boolean }) {
  return (
    <svg className="eh-flappy-skin-icon" viewBox="0 0 16 16" aria-hidden="true">
      <defs>
        <linearGradient id="eh-pipe-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={cosmic ? '#FF6B6B' : '#2E7D32'} />
          <stop offset="100%" stopColor={cosmic ? '#818CF8' : '#1B5E20'} />
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

  const [useCosmicPipes, setUseCosmicPipes] = useState(false);
  const [useGoldenBird, setUseGoldenBird] = useState(false);
  const [selectedLevel, setSelectedLevel] = useState(1);
  const [shake, setShake] = useState(false);
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('flappy');

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

  useEffect(() => {
    const savedPipes = localStorage.getItem('flappy_rainbow_pipes');
    const savedBird = localStorage.getItem('flappy_golden_bird');
    if (savedPipes !== null) setUseCosmicPipes(savedPipes === 'true');
    if (savedBird !== null) setUseGoldenBird(savedBird === 'true');
    const savedLevel = parseInt(localStorage.getItem('flappy_level') || '1', 10);
    const lv = Number.isFinite(savedLevel) ? Math.min(10, Math.max(1, savedLevel)) : 1;
    setSelectedLevel(lv);
    setLevel(lv);
  }, [setLevel]);

  const toggleCosmicPipes = () => {
    const newVal = !useCosmicPipes;
    setUseCosmicPipes(newVal);
    localStorage.setItem('flappy_rainbow_pipes', String(newVal));
  };

  const toggleGoldenBird = () => {
    const newVal = !useGoldenBird;
    setUseGoldenBird(newVal);
    localStorage.setItem('flappy_golden_bird', String(newVal));
  };

  useEffect(() => {
    if (!token) {
      navigate('/login');
    }
  }, [token, navigate]);

  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    if (boostError) {
      setSaveMessage({ type: 'error', text: boostError });
      setBoostError(null);
    }
  }, [boostError, setBoostError]);

  const beginRun = async () => {
    if (boostBusy || started) return;
    const lv = Math.min(10, Math.max(1, selectedLevel));
    localStorage.setItem('flappy_level', String(lv));
    setLevel(lv);
    try {
      const { boostId, boosted: isBoosted } = await armBoost();
      startGame({ boosted: isBoosted, boostId, level: lv });
    } catch {
      /* handled */
    }
  };

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
        text: boostUnrankedToast(),
      });
    }
  }, [lastSubmitMessage, lastSubmitRanked]);

  const handleManualSave = async () => {
    const state = useFlappyStore.getState();
    const { score, boostId, boosted: runBoosted, level: runLevel } = state;
    const userId = localStorage.getItem('userId');
    const userEmail = localStorage.getItem('userEmail');
    const accessToken = localStorage.getItem('accessToken');

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
        headers: { Authorization: `Bearer ${accessToken}` },
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
          setSaveMessage({ type: 'success', text: 'Счёт сохранён · рекорд в лидерборд' });
          void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
        } else {
          setSaveMessage({ type: 'success', text: boostUnrankedToast() });
        }
      }
    } catch {
      setSaveMessage({ type: 'error', text: 'Ошибка сохранения' });
    }
  };

  useEffect(() => {
    if (gameOver && !prevGameOver.current) {
      setShake(true);
      const t = window.setTimeout(() => setShake(false), 420);
      prevGameOver.current = true;
      return () => window.clearTimeout(t);
    }
    if (!gameOver) prevGameOver.current = false;
  }, [gameOver]);

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

    const cosmic = useCosmicPipes && skins.flappy.hasRainbowPipes;
    const golden = useGoldenBird && skins.flappy.hasGoldenBird;

    for (const pipe of pipes) {
      drawPipe(ctx, pipe, 60, FLAPPY_H, cosmic);
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
    useCosmicPipes,
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
  };

  const handleBack = () => navigate('/#games');

  const handleNewGame = () => {
    cloudScrollRef.current = 0;
    starScrollRef.current = 0;
    resetGame();
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
              Boost
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
              onClick={toggleCosmicPipes}
              title="Космические трубы"
              className={cn(
                'rounded-sm border px-3 py-1.5 text-sm transition-colors',
                useCosmicPipes
                  ? 'border-photon-cyan/50 bg-photon-cyan/15 text-photon-cyan'
                  : 'border-white/10 text-text-secondary hover:border-white/20 hover:text-text-primary',
              )}
            >
              <IconPipes cosmic={useCosmicPipes} />
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
                  <option key={lv} value={lv} className="bg-void text-text-primary">
                    {lv}
                  </option>
                ))}
              </select>
            </label>
          )}
          <Button variant="primary" size="sm" onClick={handleNewGame} disabled={boostBusy}>
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
          {!started && (
            <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy} />
          )}
        </>
      }
      help={
        <>
          <p>Нажимайте ПРОБЕЛ или кликайте мышкой, чтобы птичка летела вверх</p>
          <p>Не врезайтесь в трубы и не падайте на землю</p>
          <p>Уровни 1–10: выше уровень — уже и быстрее трубы; щель чуть варьируется от трубы к трубе</p>
          <p>Очки за трубу × уровень (L1 ≠ L10). Лидерборд отдельный на каждый уровень</p>
          {boostHelpLines('flappy').map((line) => (
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
      <div className={cn('eh-flappy-stage', shake && 'eh-flappy-stage--shake')}>
        <canvas
          ref={canvasRef}
          width={FLAPPY_W}
          height={FLAPPY_H}
          onClick={handleCanvasClick}
        />
      </div>

      <Modal open={gameOver} onClose={() => {}} title="Game Over">
        <p className="text-lg text-text-primary">
          Счёт: <span className="font-display font-semibold text-horizon-gold">{score}</span>
        </p>
        <p className="mt-1 text-sm text-text-secondary">Уровень {level}</p>
        {boosted && (
          <p className="mt-3 text-sm text-horizon-gold">{boostUnrankedToast()}</p>
        )}
        {!boosted && lastSubmitRanked === true && (
          <p className="mt-3 inline-flex items-center gap-2 text-sm text-success">
            <Icon name="check" className="h-4 w-4" aria-hidden />
            Счёт сохранён · рекорд в лидерборд
          </p>
        )}
        {!boosted && lastSubmitRanked === false && lastSubmitMessage && (
          <p className="mt-3 text-sm text-text-secondary">{lastSubmitMessage}</p>
        )}
        <GameOverActions
          onNewGame={handleNewGame}
          onHome={handleBack}
          onSave={() => void handleManualSave()}
          saveLabel="Сохранить рекорд"
          busy={boostBusy}
        />
      </Modal>
    </GameShell>
  );
}
