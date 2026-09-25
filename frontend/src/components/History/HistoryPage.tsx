import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { historyApi, type HistoryEvent } from '../../services/historyApi';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';

const PAGE_SIZE = 50;

const TYPE_FILTERS = [
  { value: '', label: 'Все' },
  { value: 'user.registered', label: 'Регистрация' },
  { value: 'score.updated', label: 'Рекорды' },
  { value: 'shop.purchased', label: 'Покупки' },
  { value: 'payment.completed', label: 'Оплата' },
  { value: 'author.upserted', label: 'Авторы' },
];

function formatTime(unix: number): string {
  if (!unix) return '—';
  return new Date(unix * 1000).toLocaleString('ru-RU');
}

function prettyPayload(raw: string): string {
  if (!raw) return '';
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

export function HistoryPage() {
  const navigate = useNavigate();
  const [events, setEvents] = useState<HistoryEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [eventType, setEventType] = useState('');
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!localStorage.getItem('accessToken')) {
      navigate('/login');
    }
  }, [navigate]);

  useEffect(() => {
    if (!localStorage.getItem('accessToken')) return;

    let cancelled = false;

    const run = async () => {
      try {
        setLoading(true);
        setError('');
        const data = await historyApi.list(PAGE_SIZE, 0, eventType);
        if (cancelled) return;
        setEvents(data.events ?? []);
        setTotal(data.total ?? 0);
      } catch {
        if (!cancelled) setError('Не удалось загрузить историю событий');
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    void run();
    return () => {
      cancelled = true;
    };
  }, [eventType]);

  const loadMore = async () => {
    try {
      setLoadingMore(true);
      const data = await historyApi.list(PAGE_SIZE, events.length, eventType);
      setEvents((prev) => [...prev, ...(data.events ?? [])]);
      setTotal(data.total ?? 0);
    } catch {
      setError('Не удалось загрузить историю событий');
    } finally {
      setLoadingMore(false);
    }
  };

  if (!localStorage.getItem('accessToken')) return null;

  return (
    <PageShell width="wide">
      <PageHeader
        title="📜 История"
        subtitle="События вашего аккаунта (окно хранения ~30 дней)"
        onBack={() => navigate('/')}
        backLabel="На главную"
      />

      <div className="mb-6 flex flex-wrap gap-2">
        {TYPE_FILTERS.map((f) => (
          <FilterChip
            key={f.value || 'all'}
            active={eventType === f.value}
            onClick={() => setEventType(f.value)}
          >
            {f.label}
          </FilterChip>
        ))}
      </div>

      {error && (
        <div role="alert" className="mb-4 rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-16">
          <Spinner size={48} />
        </div>
      ) : events.length === 0 ? (
        <div className="rounded-md border border-white/10 bg-nebula py-12 text-center text-text-secondary">
          Пока нет событий
        </div>
      ) : (
        <>
          <p className="mb-3 text-sm text-text-secondary">Всего: {total}</p>
          <ul className="flex flex-col gap-3">
            {events.map((ev) => (
              <Card key={ev.id} as="li">
                <div className="flex items-baseline justify-between gap-4">
                  <span className="font-hud text-sm font-semibold text-indigo-soft">{ev.event_type || 'event'}</span>
                  <span className="shrink-0 text-xs text-text-muted">{formatTime(ev.created_at_unix)}</span>
                </div>
                {ev.payload_json && (
                  <pre className="mt-3 overflow-x-auto whitespace-pre-wrap break-words rounded-sm bg-void p-3 font-hud text-xs text-text-secondary">
                    {prettyPayload(ev.payload_json)}
                  </pre>
                )}
              </Card>
            ))}
          </ul>
          {events.length < total && (
            <div className="mt-4 flex justify-center">
              <Button variant="ghost" disabled={loadingMore} onClick={() => void loadMore()}>
                {loadingMore ? 'Загрузка…' : 'Ещё'}
              </Button>
            </div>
          )}
        </>
      )}
    </PageShell>
  );
}
