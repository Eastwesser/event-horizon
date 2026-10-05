import { useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUserRole } from '../../hooks/useUserRole';
import { clearAuth, getAccessToken } from '../../lib/auth';
import { Button } from '../ui/Button';
import { IconLabel, type IconName } from '../ui/Icon';

/** Always visible — kids-safe primary destinations. */
const primaryNav: { label: string; path: string; icon: IconName }[] = [
  { label: 'Профиль', path: '/profile', icon: 'user' },
  { label: 'Лидерборд', path: '/leaderboard', icon: 'trophy' },
  { label: 'Магазин', path: '/shop', icon: 'cart' },
];

/** Secondary — burger menu (fewer labels on the bar). */
const moreNav: { label: string; path: string; icon: IconName }[] = [
  { label: 'Инвентарь', path: '/inventory', icon: 'package' },
  { label: 'История', path: '/history', icon: 'scroll' },
  { label: 'Авторы', path: '/authors', icon: 'pen' },
  { label: 'Подписка', path: '/subscription', icon: 'credit-card' },
];

/** Shared with Home footer — keep padding in sync with PageShell. */
export const shellInner = 'mx-auto box-border w-full max-w-6xl px-6 sm:px-8';

export function AppNavbar() {
  const navigate = useNavigate();
  const token = getAccessToken();
  const { isAdmin, isAuthor } = useUserRole();
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
              <IconLabel name={link.icon}>{link.label}</IconLabel>
            </button>
          ))}
        </nav>

        <div className="flex shrink-0 items-center justify-end gap-2 sm:gap-3">
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
                      <IconLabel name={link.icon}>{link.label}</IconLabel>
                    </button>
                  ))}
                </div>
                <div className="py-1">
                  {moreNav.map((link) => (
                    <button
                      key={link.path}
                      type="button"
                      role="menuitem"
                      onClick={() => go(link.path)}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      <IconLabel name={link.icon}>{link.label}</IconLabel>
                    </button>
                  ))}
                </div>
                {isAuthor && (
                  <div className="border-t border-white/10 py-1">
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => go('/author/dashboard')}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      <IconLabel name="pen">Автор</IconLabel>
                    </button>
                  </div>
                )}
                {isAdmin && (
                  <div className="border-t border-white/10 py-1">
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => go('/admin')}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      <IconLabel name="wrench">Админ-панель</IconLabel>
                    </button>
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => go('/analytics')}
                      className="block w-full px-4 py-2.5 text-left text-sm text-text-secondary transition-colors hover:bg-white/5 hover:text-indigo-soft"
                    >
                      <IconLabel name="chart">Аналитика</IconLabel>
                    </button>
                  </div>
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
  );
}

export function AppFooter() {
  return (
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
  );
}
