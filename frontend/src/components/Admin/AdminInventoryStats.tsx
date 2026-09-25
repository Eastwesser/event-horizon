import { useEffect, useState, type ReactNode } from 'react';
import { adminApi, type InventoryStats, type TopExpensiveItem } from '../../services/adminApi';
import { formatRubPrice } from '../../lib/formatPrice';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { Spinner } from '../ui/Spinner';
import { StatCard } from '../ui/StatCard';

function truncateId(id: string): string {
  if (id.length <= 12) return id;
  return `${id.slice(0, 4)}…${id.slice(-4)}`;
}

function sortedEntries(map: Record<string, number>): [string, number][] {
  return Object.entries(map).toSorted((a, b) => {
    if (b[1] !== a[1]) return b[1] - a[1];
    return a[0].localeCompare(b[0]);
  });
}

function CountList({
  title,
  entries,
  renderKey,
}: {
  title: string;
  entries: [string, number][];
  renderKey: (key: string) => ReactNode;
}) {
  return (
    <Card className="flex flex-col gap-3 p-4">
      <h3 className="text-sm font-medium text-text-primary">{title}</h3>
      {entries.length === 0 ? (
        <p className="py-6 text-center text-sm text-text-muted">Нет данных</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {entries.map(([key, count]) => (
            <li
              key={key}
              className="flex items-center justify-between gap-3 border-b border-white/5 pb-2 last:border-0 last:pb-0"
            >
              <span className="min-w-0 truncate text-sm text-text-secondary">{renderKey(key)}</span>
              <Badge tone="indigo">{count}</Badge>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}

function TopExpensiveList({ items }: { items: TopExpensiveItem[] }) {
  return (
    <Card className="flex flex-col gap-3 p-4">
      <h3 className="text-sm font-medium text-text-primary">Топ-5 по цене</h3>
      {items.length === 0 ? (
        <p className="py-6 text-center text-sm text-text-muted">Нет данных</p>
      ) : (
        <ol className="flex flex-col gap-2">
          {items.map((item, i) => (
            <li
              key={item.id || `${item.name}-${i}`}
              className="flex items-center justify-between gap-3 border-b border-white/5 pb-2 last:border-0 last:pb-0"
            >
              <span className="min-w-0 truncate text-sm text-text-secondary">
                <span className="mr-2 font-hud text-text-muted">{i + 1}.</span>
                {item.name || '—'}
              </span>
              <Badge tone="gold">{formatRubPrice(item.price)}</Badge>
            </li>
          ))}
        </ol>
      )}
    </Card>
  );
}

export function AdminInventoryStats() {
  const [stats, setStats] = useState<InventoryStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError('');
    void adminApi
      .getInventoryStats()
      .then((res) => {
        if (!cancelled) setStats(res);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const status = (err as { response?: { status?: number } })?.response?.status;
        setError(status === 403 ? 'Доступ запрещён' : 'Не удалось загрузить статистику');
        setStats(null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  const byType = stats ? sortedEntries(stats.by_type) : [];
  const byAuthor = stats ? sortedEntries(stats.by_author) : [];
  const isEmpty =
    !!stats &&
    stats.total_items === 0 &&
    byType.length === 0 &&
    byAuthor.length === 0;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-medium text-text-primary">Статистика инвентаря</h2>
        <Button
          type="button"
          size="sm"
          variant="ghost"
          disabled={loading}
          onClick={() => setReloadKey((k) => k + 1)}
        >
          Обновить
        </Button>
      </div>

      {error && (
        <Card className="flex flex-wrap items-center justify-between gap-3 border-error/30 bg-error/10 text-sm text-error">
          <span>{error}</span>
          <Button type="button" size="sm" variant="ghost" onClick={() => setReloadKey((k) => k + 1)}>
            Повторить
          </Button>
        </Card>
      )}

      {loading ? (
        <div className="flex justify-center py-16">
          <Spinner />
        </div>
      ) : isEmpty ? (
        <p className="py-12 text-center text-text-secondary">В инвентаре пока нет товаров</p>
      ) : stats ? (
        <>
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <StatCard value={stats.total_items} label="Всего товаров" tone="indigo" />
            <StatCard value={stats.total_stock} label="Всего stock" tone="gold" />
            <StatCard
              value={Object.keys(stats.by_type).length}
              label="Типов товаров"
              tone="indigo"
            />
            <StatCard
              value={Object.keys(stats.by_author).length}
              label="Авторов"
              tone="cyan"
            />
          </div>

          <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
            <CountList title="По типу" entries={byType} renderKey={(key) => key} />
            <CountList
              title="По автору"
              entries={byAuthor}
              renderKey={(key) => {
                const email = stats.author_emails[key];
                if (email) {
                  return (
                    <span title={key}>
                      {email}
                    </span>
                  );
                }
                return (
                  <span className="font-hud" title={key}>
                    {truncateId(key)}
                  </span>
                );
              }}
            />
            <TopExpensiveList items={stats.top_expensive} />
          </div>
        </>
      ) : null}
    </div>
  );
}
