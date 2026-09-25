// frontend/src/components/Home/Home.tsx
import { useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUserRole } from '../../hooks/useUserRole';
import { clearAuth, getAccessToken } from '../../lib/auth';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';

const games = [
  {
    id: 'hexagon',
    name: 'Pancaker',
    description: 'Гексагональный пазл с блинчиками',
    icon: '🥞',
    path: '/game/hexagon',
    available: true,
  },
  {
    id: 'flappy',
    name: 'Flappy Bird',
    description: 'Лети и не врезайся в трубы',
    icon: '🐦',
    path: '/game/flappy',
    available: true,
  },
  {
    id: 'towers',
    name: 'Builder',
    description: 'Строй башню из падающих блоков',
    icon: '🗼',
    path: '/game/towers',
    available: true,
  },
  {
    id: 'hanoi',
    name: 'Hanoi',
    description: 'Классическая головоломка с кольцами',
    icon: '🪈',
    path: '/game/hanoi',
    available: true,
  },
  {
    id: 'memory',
    name: 'Memonia',
    description: 'Найди пары фруктов',
    icon: '🎴',
    path: '/game/memory',
    available: true,
  },
];

/** Always visible — kids-safe primary destinations. */
const primaryNav = [
  { label: '👤 Профиль', path: '/profile' },
  { label: '🏆 Лидерборд', path: '/leaderboard' },
  { label: '🛒 Магазин', path: '/shop' },
];

/** Secondary — burger menu (fewer labels on the bar). */
const moreNav = [
  { label: '📦 Инвентарь', path: '/inventory' },
  { label: '📜 История', path: '/history' },
  { label: '✍️ Авторы', path: '/authors' },
  { label: '💳 Подписка', path: '/subscription' },
];

/** Inline classes (not a variable) so Tailwind always scans px-6 / sm:px-8. */
const shellInner =
  'mx-auto box-border w-full max-w-6xl px-6 sm:px-8';

export function Home() {
  const navigate = useNavigate();
  const token = getAccessToken();
  const { isAdmin } = useUserRole();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuId = useId();
  const menuRef = useRef<HTMLDivElement>(null);

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  useEffect(() => {
    if (!menuOpen) return;
    const onDoc = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuOpen(false);
    };
    document.addEventListener('mousedown', onDoc);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDoc);
      document.removeEventListener('keydown', onKey);
    };
  }, [menuOpen]);

  const go = (path: string) => {
    setMenuOpen(false);
    navigate(path);
  };

  return (
    <div className="min-h-screen bg-void font-body text-text-primary">
      <header className="sticky top-0 z-40 border-b border-white/10 bg-void/80 backdrop-blur-md">
        <div
          className={`${shellInner} grid grid-cols-[auto_1fr_auto] items-center gap-3 py-3 sm:gap-4 sm:py-4`}
        >
          <button
            type="button"
            onClick={() => navigate('/')}
            className="flex shrink-0 items-center gap-2 font-display text-lg font-bold tracking-tight text-horizon-gold transition-colors hover:text-horizon-gold-hot sm:gap-2.5 sm:text-xl"
          >
            <img
              src="/images/brand/logo-icon.png"
              alt=""
              width={36}
              height={36}
              className="h-8 w-8 rounded-md object-cover sm:h-9 sm:w-9"
            />
            <span className="whitespace-nowrap">Event Horizon</span>
          </button>

          <nav className="hidden min-w-0 items-center justify-center gap-x-5 text-sm text-text-secondary sm:flex">
            {primaryNav.map((link) => (
              <button
                key={link.path}
                type="button"
                onClick={() => navigate(link.path)}
                className="shrink-0 whitespace-nowrap transition-colors hover:text-indigo-soft"
              >
                {link.label}
              </button>
            ))}
          </nav>

          <div className="flex shrink-0 items-center justify-end gap-2 sm:gap-3">
            {/* Primary links move into burger on very small screens */}
            <div className="relative" ref={menuRef}>
              <button
                type="button"
                aria-expanded={menuOpen}
                aria-controls={menuId}
                aria-label={menuOpen ? 'Закрыть меню' : 'Ещё разделы'}
                onClick={() => setMenuOpen((o) => !o)}
                className="flex h-9 w-9 items-center justify-center rounded-sm border border-white/10 text-text-secondary transition-colors hover:border-indigo/40 hover:text-indigo-soft"
              >
                <span className="flex flex-col gap-[3px]" aria-hidden="true">
                  <span className="block h-0.5 w-4 rounded-full bg-current" />
                  <span className="block h-0.5 w-4 rounded-full bg-current" />
                  <span className="block h-0.5 w-4 rounded-full bg-current" />
                </span>
              </button>

              {menuOpen && (
                <div
                  id={menuId}
                  role="menu"
                  className="absolute right-0 top-full z-50 mt-2 min-w-[12rem] rounded-md border border-white/10 bg-nebula-elevated py-1 shadow-elevated"
                >
                  <div className="border-b border-white/10 py-1 sm:hidden">
                    {primaryNav.map((link) => (
                      <button
                        key={link.path}
                        type="button"
                        role="menuitem"
                        onClick={() => go(link.path)}
                        className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                      >
                        {link.label}
                      </button>
                    ))}
                  </div>
                  {moreNav.map((link) => (
                    <button
                      key={link.path}
                      type="button"
                      role="menuitem"
                      onClick={() => go(link.path)}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      {link.label}
                    </button>
                  ))}
                  {isAdmin && (
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => go('/admin')}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      🛠 Админ-панель
                    </button>
                  )}
                  {isAdmin && (
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => go('/analytics')}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      📊 Аналитика
                    </button>
                  )}
                </div>
              )}
            </div>

            {token && (
              <Button variant="ghost" size="sm" onClick={handleLogout} className="shrink-0">
                Выйти
              </Button>
            )}
          </div>
        </div>
      </header>

      <main className={shellInner}>
        <section className="relative py-12 sm:py-16">
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
                  onClick={() => document.getElementById('games')?.scrollIntoView({ behavior: 'smooth' })}
                >
                  Все игры
                </Button>
                <Button variant="ghost" size="md" onClick={() => navigate('/leaderboard')}>
                  🏆 Лидерборд
                </Button>
              </div>
            </div>

            {/* Decorative accretion disk — brand mark in the core (no flagship game). */}
            <div className="eh-disk mx-auto" aria-hidden="true">
              <div className="eh-disk-rings">
                <div className="eh-disk-ring eh-disk-ring--lensed" />
                <div className="eh-disk-ring eh-disk-ring--mid" />
                <div className="eh-disk-ring eh-disk-ring--main" />
              </div>
              <div className="eh-disk-core" />
              <div className="absolute left-1/2 top-1/2 flex w-[42%] -translate-x-1/2 -translate-y-1/2 flex-col items-center justify-center">
                <img
                  src="/images/brand/logo-minimal.png"
                  alt=""
                  className="h-auto w-full rounded-full object-cover opacity-95"
                />
              </div>
            </div>
          </div>
        </section>

        <section id="games" className="scroll-mt-24 pb-16 pt-4">
          <h2 className="mb-6 font-display text-xl font-semibold text-text-primary">Игры</h2>
          <div className="grid grid-cols-1 items-stretch gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
            {games.map((game, i) => (
              <Card
                key={game.id}
                interactive={game.available}
                className="eh-stagger-in flex h-full min-h-[220px] flex-col items-center gap-2 text-center"
                style={{ animationDelay: `${i * 70}ms` }}
                onClick={() => game.available && navigate(game.path)}
              >
                <div className="text-5xl">{game.icon}</div>
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

      <footer className="border-t border-white/10">
        <div
          className={`${shellInner} flex flex-col items-center gap-3 py-8 text-center text-sm text-text-muted sm:flex-row sm:justify-between sm:text-left`}
        >
          <p>© 2026 Event Horizon. Игры без FOMO и скрытых обнулений.</p>
          <a
            href="https://boosty.to/eastwesser"
            target="_blank"
            rel="noreferrer"
            className="font-medium text-indigo-soft transition-colors hover:text-horizon-gold hover:underline"
          >
            Поддержать проект
          </a>
        </div>
      </footer>
    </div>
  );
}
