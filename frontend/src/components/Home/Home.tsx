// frontend/src/components/Home/Home.tsx
import { useEffect, useRef, useState, type CSSProperties, type PointerEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Icon, IconLabel, type IconName } from '../ui/Icon';
import { AppFooter, AppNavbar, shellInner } from '../Layout/AppNavbar';
import { gameIcon } from '../../lib/gameIcons';
import { cn } from '../../lib/cn';
import { VOID_PARTICLES } from './voidParticles';
import {
  getVoidHideSet,
  isVoidDebug,
  prefersReducedMotion,
  syncMotionForceClass,
  syncVoidDebugMarks,
} from '../../lib/motion';

const games: {
  id: string;
  name: string;
  description: string;
  icon: IconName;
  path: string;
  available: boolean;
}[] = [
  {
    id: 'hexagon',
    name: 'Pancaker',
    description: 'Гексагональный пазл с блинчиками',
    icon: gameIcon('hexagon'),
    path: '/game/hexagon',
    available: true,
  },
  {
    id: 'flappy',
    name: 'Flappy Bird',
    description: 'Лети и не врезайся в трубы',
    icon: gameIcon('flappy'),
    path: '/game/flappy',
    available: true,
  },
  {
    id: 'towers',
    name: 'Builder',
    description: 'Строй башню из падающих блоков',
    icon: gameIcon('towers'),
    path: '/game/towers',
    available: true,
  },
  {
    id: 'hanoi',
    name: 'Hanoi',
    description: 'Классическая головоломка с кольцами',
    icon: gameIcon('hanoi'),
    path: '/game/hanoi',
    available: true,
  },
  {
    id: 'memory',
    name: 'Memonia',
    description: 'Найди пары фруктов',
    icon: gameIcon('memory'),
    path: '/game/memory',
    available: true,
  },
  {
    id: 'twenty48',
    name: 'Горизонт 2048',
    description: 'Сдвинь плитки — собери 2048',
    icon: gameIcon('twenty48'),
    path: '/game/twenty48',
    available: true,
  },
  {
    id: 'gears',
    name: 'Орбиты',
    description: 'Сливай шестерёнки до восьмой',
    icon: gameIcon('gears'),
    path: '/game/gears',
    available: true,
  },
  {
    id: 'companion',
    name: 'Компаньон',
    description: 'Мягкий тамагочи без FOMO-смерти',
    icon: gameIcon('companion'),
    path: '/game/companion',
    available: true,
  },
];

export function Home() {
  const navigate = useNavigate();
  const heroRef = useRef<HTMLElement>(null);
  const diskRef = useRef<HTMLDivElement>(null);
  const pullLayerRef = useRef<HTMLDivElement>(null);
  const hotTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [scrollHintFaded, setScrollHintFaded] = useState(false);
  const [diskNear, setDiskNear] = useState(false);
  const [diskPull, setDiskPull] = useState(false);
  const [diskHot, setDiskHot] = useState(false);
  /** Bumps on each enter so particle animations remount at frame 0. */
  const [pullCycle, setPullCycle] = useState(0);

  useEffect(() => {
    const el = heroRef.current;
    if (!el) return;

    const io = new IntersectionObserver(
      ([entry]) => {
        setScrollHintFaded(!entry.isIntersecting || entry.intersectionRatio < 0.45);
      },
      { threshold: [0, 0.45, 0.8] },
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  useEffect(() => {
    return () => {
      if (hotTimerRef.current) clearTimeout(hotTimerRef.current);
    };
  }, []);

  // VOID debug — build mode + reduced-motion once; animationstart while mounted.
  useEffect(() => {
    const motionForce = syncMotionForceClass();
    const voidDebug = isVoidDebug();
    const osReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    console.log('[VOID] mount', {
      mode: import.meta.env.MODE,
      dev: import.meta.env.DEV,
      href: window.location.href,
      osReducedMotion: osReduced,
      motionForce,
      voidDebug,
      hide: [...getVoidHideSet()],
      reducedMotion: prefersReducedMotion(),
    });
    // Ancestors may not be tagged if this ran before paint; refresh marks.
    if (voidDebug) syncVoidDebugMarks();

    const disk = diskRef.current;
    const pull = pullLayerRef.current;
    if (!disk || !pull) return;

    const onAnimStart = (e: AnimationEvent) => {
      console.log('[VOID] animation start', e.animationName, (e.target as Element)?.className);
    };
    const onPullTransition = () => {
      console.log('[VOID] pull transition run', getComputedStyle(pull).transform);
    };
    // Glow + particles bubble on disk; pull layer itself is mostly transitions.
    disk.addEventListener('animationstart', onAnimStart);
    pull.addEventListener('animationstart', onAnimStart);
    pull.addEventListener('transitionrun', onPullTransition);

    return () => {
      disk.removeEventListener('animationstart', onAnimStart);
      pull.removeEventListener('animationstart', onAnimStart);
      pull.removeEventListener('transitionrun', onPullTransition);
    };
  }, []);

  const logDiskClasses = (label: 'enter' | 'leave') => {
    // Wait for React commit so --pull/--near are on the node.
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        const el = diskRef.current;
        console.log(`[VOID] ${label}`, el?.classList?.value ?? el?.className ?? '(no disk)');
      });
    });
  };

  const scrollToChoose = () => {
    const el = document.getElementById('choose');
    if (!el) return;
    el.scrollIntoView({
      behavior: prefersReducedMotion() ? 'auto' : 'smooth',
      block: 'start',
    });
  };

  const clearHotTimer = () => {
    if (hotTimerRef.current) {
      clearTimeout(hotTimerRef.current);
      hotTimerRef.current = null;
    }
  };

  const onDiskHitMove = (e: PointerEvent<HTMLDivElement>) => {
    const disk = diskRef.current;
    if (!disk) return;
    const r = disk.getBoundingClientRect();
    const cx = r.left + r.width / 2;
    const cy = r.top + r.height / 2;
    const dist = Math.hypot(e.clientX - cx, e.clientY - cy);
    setDiskNear(dist <= r.width / 2 + 80);
  };

  const onDiskEnter = () => {
    setPullCycle((n) => n + 1);
    setDiskPull(true);
    setDiskNear(true);
    clearHotTimer();
    hotTimerRef.current = setTimeout(() => setDiskHot(true), 2000);
    logDiskClasses('enter');
  };

  const onDiskLeave = () => {
    setDiskPull(false);
    setDiskHot(false);
    clearHotTimer();
    logDiskClasses('leave');
  };

  const onDiskHitLeave = () => {
    setDiskNear(false);
    setDiskPull(false);
    setDiskHot(false);
    clearHotTimer();
    logDiskClasses('leave');
  };

  return (
    <div className="min-h-screen bg-void font-body text-text-primary">
      <AppNavbar />

      <section ref={heroRef} className="eh-hero" aria-label="Event Horizon">
        <div className="eh-hero-stars" aria-hidden="true" />
        <img
          className="eh-hero-planet"
          src="/images/brand/planet-hero.png"
          alt=""
          width={1238}
          height={600}
          decoding="async"
          fetchPriority="high"
        />
        <div className={`${shellInner} eh-hero-wordmark-shell`}>
          <h1 className="eh-hero-wordmark" aria-label="EVENT HORIZON">
            <span aria-hidden="true">E</span>
            <span aria-hidden="true">V</span>
            <span aria-hidden="true">E</span>
            <span aria-hidden="true">N</span>
            <span aria-hidden="true">T</span>
            <span aria-hidden="true" className="eh-hero-wordmark-gap" />
            <span aria-hidden="true">H</span>
            <span aria-hidden="true">O</span>
            <span aria-hidden="true">R</span>
            <span aria-hidden="true">I</span>
            <span aria-hidden="true">Z</span>
            <span aria-hidden="true">O</span>
            <span aria-hidden="true">N</span>
          </h1>
        </div>
        <div className="eh-hero-fade-top" aria-hidden="true" />
        <div className="eh-hero-fade-bottom" aria-hidden="true" />
        <button
          type="button"
          className="eh-hero-scroll"
          data-faded={scrollHintFaded ? 'true' : 'false'}
          onClick={scrollToChoose}
          aria-label="Прокрутить к играм"
        >
          <svg
            className="eh-hero-scroll__chevron"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.75"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M6 9l6 6 6-6" />
          </svg>
        </button>
      </section>

      <main className={shellInner}>
        <section id="choose" className="eh-choose relative py-12 sm:py-16">
          <div className="grid items-center gap-12 lg:grid-cols-2 lg:gap-16">
            <div className="text-center lg:text-left">
              <h1 className="font-display text-4xl font-bold leading-tight text-text-primary sm:text-5xl">
                Выбери игру и ставь рекорды
              </h1>
              <p className="mx-auto mt-4 max-w-lg text-lg text-text-secondary lg:mx-0">
                Играй в мини-игры, зарабатывай лампочки и билетики, становись лучшим в лидерборде.
              </p>
              <div className="mt-8 flex flex-wrap justify-center gap-4 lg:justify-start">
                <Button
                  size="md"
                  onClick={() => {
                    const el = document.getElementById('games');
                    if (!el) return;
                    el.scrollIntoView({
                      behavior: prefersReducedMotion() ? 'auto' : 'smooth',
                      block: 'start',
                    });
                  }}
                >
                  Все игры
                </Button>
                <Button variant="ghost" size="md" onClick={() => navigate('/leaderboard')}>
                  <IconLabel name="trophy" iconClassName="h-4 w-4">
                    Лидерборд
                  </IconLabel>
                </Button>
              </div>
            </div>

            {/* VOID v2 — accretion disk, no logo. Near / pull / hot via classes. */}
            <div
              className="eh-disk-hit mx-auto"
              onPointerMove={onDiskHitMove}
              onPointerLeave={onDiskHitLeave}
              aria-hidden="true"
            >
              <div
                ref={diskRef}
                className={cn(
                  'eh-disk',
                  diskNear && 'eh-disk--near',
                  diskPull && 'eh-disk--pull',
                  diskHot && 'eh-disk--hot',
                )}
                onPointerEnter={onDiskEnter}
                onPointerLeave={onDiskLeave}
              >
                <div className="eh-disk-glow" />
                <div ref={pullLayerRef} className="eh-disk-pull">
                  <div className="eh-disk-rings">
                    <div className="eh-disk-ring eh-disk-ring--r1" />
                    <div className="eh-disk-ring eh-disk-ring--r2" />
                    <div className="eh-disk-ring eh-disk-ring--r3" />
                    <div className="eh-disk-ring eh-disk-ring--r4" />
                    <div className="eh-disk-ring eh-disk-ring--r5" />
                  </div>
                </div>
                <div className="eh-disk-core" />
                <svg key={pullCycle} className="eh-disk-particles" viewBox="0 0 100 100">
                  <defs>
                    <symbol id="eh-star4" viewBox="0 0 10 10">
                      <path d="M5 0.4 L5.85 4.15 L9.6 5 L5.85 5.85 L5 9.6 L4.15 5.85 L0.4 5 L4.15 4.15 Z" />
                    </symbol>
                  </defs>
                  {VOID_PARTICLES.map((p, i) => {
                    const style = {
                      '--ox': `${p.cx}px`,
                      '--oy': `${p.cy}px`,
                      '--dur': `${p.dur}s`,
                      '--delay': `${p.delay}s`,
                      '--spin': p.spin,
                      '--rot': `${p.rot}deg`,
                    } as CSSProperties;
                    const cls = `eh-disk-particle eh-disk-particle--${p.tone}`;
                    if (p.shape === 'dot') {
                      return (
                        <circle
                          key={i}
                          className={cls}
                          cx={p.cx}
                          cy={p.cy}
                          r={p.r}
                          style={style}
                        />
                      );
                    }
                    const size = p.r * 3.2;
                    return (
                      <use
                        key={i}
                        className={`${cls} eh-disk-particle--star4`}
                        href="#eh-star4"
                        x={p.cx - size / 2}
                        y={p.cy - size / 2}
                        width={size}
                        height={size}
                        style={style}
                      />
                    );
                  })}
                </svg>
              </div>
            </div>
          </div>
        </section>

        <section id="games" className="scroll-mt-24 pb-16 pt-4">
          <h2 className="mb-6 font-display text-xl font-semibold text-text-primary">Игры</h2>
          <div className="grid grid-cols-1 items-stretch gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {games.map((game, i) => (
              <Card
                key={game.id}
                interactive={game.available}
                className="eh-stagger-in flex h-full min-h-[220px] flex-col items-center gap-2 text-center"
                style={{ animationDelay: `${i * 70}ms` }}
                onClick={() => game.available && navigate(game.path)}
              >
                <Icon name={game.icon} className="h-12 w-12 text-horizon-gold" />
                <h3 className="font-display text-lg font-semibold text-text-primary">{game.name}</h3>
                <p className="min-h-[2.5rem] flex-1 text-sm text-text-secondary">{game.description}</p>
                <Badge
                  tone={game.available ? 'indigo' : 'neutral'}
                  className="mt-auto px-4 py-1.5 text-sm"
                >
                  {game.available ? 'Играть' : 'Скоро'}
                </Badge>
              </Card>
            ))}
          </div>
        </section>
      </main>

      <AppFooter />
    </div>
  );
}
