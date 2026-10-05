import { useCallback, useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getAccessToken } from '../../lib/auth';
import {
  notificationsApi,
  type AppNotification,
} from '../../services/notificationsApi';
import { Icon } from '../ui/Icon';

const POLL_MS = 45_000;

export function NotificationBell() {
  const navigate = useNavigate();
  const token = getAccessToken();
  const [open, setOpen] = useState(false);
  const [items, setItems] = useState<AppNotification[]>([]);
  const [unread, setUnread] = useState(0);
  const menuId = useId();
  const ref = useRef<HTMLDivElement>(null);

  const load = useCallback(async () => {
    if (!getAccessToken()) return;
    try {
      const data = await notificationsApi.list(20, 0);
      setItems(data.notifications);
      setUnread(data.unread_count);
    } catch {
      /* ignore transient errors */
    }
  }, []);

  useEffect(() => {
    if (!token) return;
    void load();
    const t = window.setInterval(() => void load(), POLL_MS);
    return () => window.clearInterval(t);
  }, [token, load]);

  useEffect(() => {
    if (!open) return;
    void load();
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onDoc);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDoc);
      document.removeEventListener('keydown', onKey);
    };
  }, [open, load]);

  if (!token) return null;

  const onItem = async (n: AppNotification) => {
    if (!n.read_at_unix) {
      try {
        await notificationsApi.markRead(n.id);
        setUnread((u) => Math.max(0, u - 1));
        setItems((prev) =>
          prev.map((x) =>
            x.id === n.id ? { ...x, read_at_unix: Math.floor(Date.now() / 1000) } : x,
          ),
        );
      } catch {
        /* ignore */
      }
    }
    setOpen(false);
    if (n.link) navigate(n.link);
  };

  const markAll = async () => {
    try {
      await notificationsApi.markRead('all');
      setUnread(0);
      setItems((prev) =>
        prev.map((x) => ({
          ...x,
          read_at_unix: x.read_at_unix || Math.floor(Date.now() / 1000),
        })),
      );
    } catch {
      /* ignore */
    }
  };

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        aria-expanded={open}
        aria-controls={menuId}
        aria-label={unread > 0 ? `Уведомления, непрочитанных: ${unread}` : 'Уведомления'}
        onClick={() => setOpen((o) => !o)}
        className="relative flex h-9 w-9 items-center justify-center rounded-sm border border-white/10 text-text-secondary transition-colors hover:border-indigo/40 hover:text-indigo-soft"
      >
        <Icon name="bell" className="h-5 w-5" />
        {unread > 0 && (
          <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-sm bg-horizon-gold px-1 text-[10px] font-semibold text-void">
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div
          id={menuId}
          role="menu"
          className="absolute right-0 top-full z-50 mt-2 w-[min(20rem,calc(100vw-2rem))] rounded-md border border-white/10 bg-nebula-elevated shadow-elevated"
        >
          <div className="flex items-center justify-between border-b border-white/10 px-3 py-2">
            <span className="text-sm font-medium text-text-primary">Уведомления</span>
            {unread > 0 && (
              <button
                type="button"
                className="text-xs text-indigo-soft hover:underline"
                onClick={() => void markAll()}
              >
                Прочитать все
              </button>
            )}
          </div>
          <div className="max-h-80 overflow-y-auto py-1">
            {items.length === 0 ? (
              <p className="px-3 py-4 text-sm text-text-muted">Пока пусто</p>
            ) : (
              items.map((n) => (
                <button
                  key={n.id}
                  type="button"
                  role="menuitem"
                  onClick={() => void onItem(n)}
                  className={`block w-full px-3 py-2.5 text-left transition-colors hover:bg-white/5 ${
                    n.read_at_unix ? 'opacity-70' : ''
                  }`}
                >
                  <span className="block text-sm font-medium text-text-primary">{n.title}</span>
                  <span className="mt-0.5 block text-xs text-text-secondary">{n.body}</span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
