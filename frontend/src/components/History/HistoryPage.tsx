import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { historyApi, type HistoryEvent } from '../../services/historyApi';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';
import { IconLabel } from '../ui/Icon';
import { gameIcon } from '../../lib/gameIcons';
import { Badge } from '../ui/Badge';

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

// Безопасный парсинг JSON
function safeParse(raw: string): any {
  if (!raw) return {};
  try {
    return JSON.parse(raw);
  } catch {
    return {};
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
        title={<IconLabel name="scroll" iconClassName="h-7 w-7 text-horizon-gold">История</IconLabel>}
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
        <div className="rounded-md border border-white/10 bg-nebula px-6 py-12 text-center text-text-secondary">
          <p className="font-display text-base text-text-primary">Пока нет событий</p>
          <p className="mt-2 text-sm">
            Здесь появятся регистрация, рекорды, покупки и оплаты после действий в аккаунте
            (окно ~30 дней). Сыграйте игру со сохранением или купите товар в магазине.
          </p>
        </div>
      ) : (
        <>
          <p className="mb-3 text-sm text-text-secondary">Всего: {total}</p>
          <ul className="flex flex-col gap-3">
            {events.map((ev) => {
              // Парсим payload безопасно ОДИН раз для каждой карточки
              const payload = safeParse(ev.payload_json); 
              
              return (
                <Card key={ev.id} as="li">
                  {/* Заголовок события: тип и время */}
                  <div className="flex items-baseline justify-between gap-4">
                    <span className="font-hud text-sm font-semibold text-indigo-soft">
                      {ev.event_type || 'event'}
                    </span>
                    <span className="shrink-0 text-xs text-text-muted">
                      {formatTime(ev.created_at_unix)}
                    </span>
                  </div>

                  {/* Тело события: красивая верстка */}
                  <div className="mt-3 text-sm">
                    {ev.event_type === 'score.updated' && (
                      <div className="flex items-center gap-2 text-text-secondary">
                        <IconLabel name={gameIcon(payload.game_id || 'flappy')}>
                          Игра: {payload.game_id || 'Неизвестно'}
                        </IconLabel>
                        <span>•</span>
                        <span>Результат: {payload.score || 0} очков</span>
                        {payload.is_record && (
                          <Badge tone="gold">Рекорд!</Badge>
                        )}
                      </div>
                    )}

                    {ev.event_type === 'shop.purchased' && (
                      <div className="flex items-center gap-2 text-text-secondary">
                        <IconLabel name="cart">Покупка</IconLabel>
                        <span>{payload.item_name || 'Товар'}</span>
                        <span className="text-horizon-gold">-{payload.price || 0} билетиков</span>
                      </div>
                    )}

                    {ev.event_type === 'user.registered' && (
                      <div className="text-text-secondary">
                        Добро пожаловать в Event Horizon!
                      </div>
                    )}

                    {ev.event_type === 'payment.completed' && (
                      <div className="text-text-secondary">
                        Оплата подписки: {payload.plan || 'Неизвестно'}
                      </div>
                    )}

                    {ev.event_type === 'author.upserted' && (
                      <div className="text-text-secondary">
                        Обновлен профиль автора: {payload.display_name || 'Без имени'}
                      </div>
                    )}

                    {/* Fallback для тех событий, которые мы не распарсили выше */}
                    {!['score.updated', 'shop.purchased', 'user.registered', 'payment.completed', 'author.upserted'].includes(ev.event_type) && (
                      <div className="text-text-secondary">
                        Событие: {ev.event_type}
                      </div>
                    )}
                  </div>
                </Card>
              );
            })}
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
