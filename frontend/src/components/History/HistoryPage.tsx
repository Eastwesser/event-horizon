import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { historyApi, type HistoryEvent } from '../../services/historyApi';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import './HistoryPage.css';

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
  const token = localStorage.getItem('accessToken');
  const [events, setEvents] = useState<HistoryEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [eventType, setEventType] = useState('');
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  const load = async (offset = 0, append = false, type = eventType) => {
    try {
      if (append) setLoadingMore(true);
      else setLoading(true);
      setError('');
      const data = await historyApi.list(PAGE_SIZE, offset, type);
      setEvents((prev) => (append ? [...prev, ...data.events] : data.events));
      setTotal(data.total);
    } catch {
      setError('Не удалось загрузить историю событий');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  useEffect(() => {
    if (!token) return;
    load(0, false, eventType);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [eventType, token]);

  if (!token) return null;

  return (
    <div className="history-container">
      <button className="back-btn" onClick={() => navigate('/')}>
        ← На главную
      </button>
      <header className="history-header">
        <h1>📜 История</h1>
        <p>События вашего аккаунта (окно хранения ~30 дней)</p>
      </header>

      <div className="history-filters">
        {TYPE_FILTERS.map((f) => (
          <button
            key={f.value || 'all'}
            className={`history-filter ${eventType === f.value ? 'active' : ''}`}
            onClick={() => setEventType(f.value)}
          >
            {f.label}
          </button>
        ))}
      </div>

      {error && <div className="history-error">{error}</div>}

      {loading ? (
        <LoadingSpinner />
      ) : events.length === 0 ? (
        <div className="history-empty">Пока нет событий</div>
      ) : (
        <>
          <p className="history-count">Всего: {total}</p>
          <ul className="history-list">
            {events.map((ev) => (
              <li key={ev.id} className="history-card">
                <div className="history-card-top">
                  <span className="history-type">{ev.event_type || 'event'}</span>
                  <span className="history-time">{formatTime(ev.created_at_unix)}</span>
                </div>
                {ev.payload_json && (
                  <pre className="history-payload">{prettyPayload(ev.payload_json)}</pre>
                )}
              </li>
            ))}
          </ul>
          {events.length < total && (
            <button
              className="history-load-more"
              disabled={loadingMore}
              onClick={() => load(events.length, true)}
            >
              {loadingMore ? 'Загрузка…' : 'Ещё'}
            </button>
          )}
        </>
      )}
    </div>
  );
}
