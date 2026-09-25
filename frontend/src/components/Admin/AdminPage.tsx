import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminApi, type AdminRole, type AdminUser } from '../../services/adminApi';
import { useUserRole } from '../../hooks/useUserRole';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';
import { AdminInventoryStats } from './AdminInventoryStats';
import { AdminAnalytics } from './AdminAnalytics';

type Tab = 'users' | 'inventory' | 'analytics';

const PAGE_SIZE = 50;
const ROLES: AdminRole[] = ['user', 'author', 'admin'];

function roleTone(role: string): 'gold' | 'indigo' | 'neutral' {
  if (role === 'admin') return 'gold';
  if (role === 'author') return 'indigo';
  return 'neutral';
}

function subLabel(u: AdminUser): string {
  const s = u.subscription;
  if (!s) return 'нет';
  if (s.active) return s.plan || s.status || 'активна';
  return s.status || 'нет';
}

export function AdminPage() {
  const navigate = useNavigate();
  const { isAdmin, loading: roleLoading } = useUserRole();
  const selfId = localStorage.getItem('userId') || '';

  const [tab, setTab] = useState<Tab>('users');
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [query, setQuery] = useState('');
  const [searchDraft, setSearchDraft] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [busyId, setBusyId] = useState<string | null>(null);

  useEffect(() => {
    if (!roleLoading && !isAdmin) {
      navigate('/', { replace: true });
    }
  }, [roleLoading, isAdmin, navigate]);

  useEffect(() => {
    if (!isAdmin || tab !== 'users') return;
    let cancelled = false;
    setLoading(true);
    setError('');
    void adminApi
      .listUsers({ q: query || undefined, limit: PAGE_SIZE, offset })
      .then((res) => {
        if (cancelled) return;
        setUsers(res.users);
        setTotal(res.total);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const status = (err as { response?: { status?: number } })?.response?.status;
        setError(status === 403 ? 'Доступ запрещён' : 'Не удалось загрузить пользователей');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [isAdmin, tab, query, offset]);

  const changeRole = async (user: AdminUser, next: AdminRole) => {
    if (next === user.role) return;
    if (user.user_id === selfId) {
      window.alert('Нельзя изменить свою роль — риск потерять доступ.');
      return;
    }
    const ok = window.confirm(
      `Сменить роль пользователя ${user.email}?\n\nСейчас: ${user.role}\nНовая: ${next}`,
    );
    if (!ok) return;

    setBusyId(user.user_id);
    try {
      await adminApi.updateRole(user.user_id, next);
      setUsers((prev) =>
        prev.map((u) => (u.user_id === user.user_id ? { ...u, role: next } : u)),
      );
    } catch {
      window.alert('Не удалось сменить роль');
    } finally {
      setBusyId(null);
    }
  };

  if (roleLoading || !isAdmin) {
    return (
      <PageShell width="wide">
        <div className="flex justify-center py-20">
          <Spinner />
        </div>
      </PageShell>
    );
  }

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <PageShell width="wide">
      <PageHeader
        title="Админ-панель"
        subtitle="Внутренний инструмент. Только для admin."
        onBack={() => navigate('/')}
        backLabel="На главную"
      />

      <div className="mb-6 flex flex-wrap gap-2">
        {(
          [
            ['users', 'Пользователи'],
            ['inventory', 'Инвентарь'],
            ['analytics', 'Аналитика'],
          ] as const
        ).map(([id, label]) => (
          <Button
            key={id}
            type="button"
            size="sm"
            variant={tab === id ? 'secondary' : 'ghost'}
            onClick={() => setTab(id)}
          >
            {label}
          </Button>
        ))}
      </div>

      {tab === 'users' && (
        <div className="flex flex-col gap-4">
          <form
            className="flex flex-wrap gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              setOffset(0);
              setQuery(searchDraft.trim());
            }}
          >
            <input
              type="search"
              value={searchDraft}
              onChange={(e) => setSearchDraft(e.target.value)}
              placeholder="Поиск по email…"
              className="min-w-[14rem] flex-1 rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary focus:border-photon-cyan/50"
            />
            <Button type="submit" size="sm" variant="secondary">
              Найти
            </Button>
          </form>

          {error && (
            <Card className="border-error/30 bg-error/10 text-sm text-error">{error}</Card>
          )}

          {loading ? (
            <div className="flex justify-center py-16">
              <Spinner />
            </div>
          ) : users.length === 0 ? (
            <p className="py-12 text-center text-text-secondary">Пользователи не найдены</p>
          ) : (
            <div className="overflow-x-auto rounded-md border border-white/10">
              <table className="w-full min-w-[52rem] border-collapse text-left text-sm">
                <thead className="bg-white/5 text-xs uppercase tracking-wide text-text-muted">
                  <tr>
                    <th className="px-3 py-2.5 font-medium">Email</th>
                    <th className="px-3 py-2.5 font-medium">Роль</th>
                    <th className="px-3 py-2.5 font-medium">Лампы</th>
                    <th className="px-3 py-2.5 font-medium">Билеты</th>
                    <th className="px-3 py-2.5 font-medium">Подписка</th>
                    <th className="px-3 py-2.5 font-medium">Создан</th>
                    <th className="px-3 py-2.5 font-medium">Действие</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((u) => (
                    <tr key={u.user_id} className="border-t border-white/10">
                      <td className="px-3 py-2.5">
                        <div className="font-medium text-text-primary">{u.email}</div>
                        <div className="font-hud text-xs text-text-muted">{u.nickname}</div>
                      </td>
                      <td className="px-3 py-2.5">
                        <Badge tone={roleTone(u.role)}>{u.role}</Badge>
                      </td>
                      <td className="px-3 py-2.5 font-hud tabular-nums text-text-secondary">
                        {u.lamps ?? 0}
                      </td>
                      <td className="px-3 py-2.5 font-hud tabular-nums text-text-secondary">
                        {u.tickets ?? 0}
                      </td>
                      <td className="px-3 py-2.5 text-text-secondary">{subLabel(u)}</td>
                      <td className="px-3 py-2.5 text-text-muted">
                        {u.created_at
                          ? new Date(u.created_at).toLocaleDateString('ru-RU')
                          : '—'}
                      </td>
                      <td className="px-3 py-2.5">
                        <select
                          className="rounded-sm border border-white/10 bg-nebula px-2 py-1.5 text-sm text-text-primary disabled:opacity-40"
                          value={u.role}
                          disabled={busyId === u.user_id || u.user_id === selfId}
                          onChange={(e) => void changeRole(u, e.target.value as AdminRole)}
                          title={
                            u.user_id === selfId
                              ? 'Нельзя менять свою роль'
                              : 'Сменить роль'
                          }
                        >
                          {ROLES.map((r) => (
                            <option key={r} value={r}>
                              {r}
                            </option>
                          ))}
                        </select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          <div className="flex items-center justify-between gap-3 text-sm text-text-secondary">
            <span>
              Всего: {total} · стр. {page}/{pageCount}
            </span>
            <div className="flex gap-2">
              <Button
                type="button"
                size="sm"
                variant="ghost"
                disabled={offset <= 0 || loading}
                onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}
              >
                Назад
              </Button>
              <Button
                type="button"
                size="sm"
                variant="ghost"
                disabled={offset + PAGE_SIZE >= total || loading}
                onClick={() => setOffset((o) => o + PAGE_SIZE)}
              >
                Далее
              </Button>
            </div>
          </div>
        </div>
      )}

      {tab === 'inventory' && <AdminInventoryStats />}

      {tab === 'analytics' && <AdminAnalytics />}
    </PageShell>
  );
}
