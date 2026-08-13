// frontend/src/components/Analytics/AnalyticsDashboard.tsx
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { analyticsApi, type DayCount, type RetentionPoint } from '../../services/analyticsApi';
import { useUserRole } from '../../hooks/useUserRole';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import './AnalyticsDashboard.css';

export function AnalyticsDashboard() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { isAdmin, loading: roleLoading } = useUserRole();

  const [dauDays, setDauDays] = useState<DayCount[]>([]);
  const [mau, setMau] = useState<{ mau: number; window_days: number } | null>(null);
  const [retention, setRetention] = useState<{
    cohort_day: string;
    cohort_size: number;
    points: RetentionPoint[];
  } | null>(null);

  const [days, setDays] = useState(30);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  const loadAnalytics = async () => {
    if (!isAdmin) return;
    setLoading(true);
    setError('');
    try {
      const [dauData, mauData, retentionData] = await Promise.all([
        analyticsApi.dau(days),
        analyticsApi.mau(days),
        analyticsApi.retention(7, 7),
      ]);
      setDauDays(dauData.days || []);
      setMau(mauData);
      setRetention(retentionData);
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 403) {
        setError('Доступ запрещён — только для администраторов');
      } else {
        setError('Не удалось загрузить аналитику');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!roleLoading && isAdmin) {
      loadAnalytics();
    } else if (!roleLoading && !isAdmin) {
      setLoading(false);
    }
  }, [roleLoading, isAdmin, days]);

  const maxDau = useMemo(
    () => Math.max(1, ...dauDays.map((d) => d.count)),
    [dauDays]
  );

  const maxRetention = useMemo(
    () => Math.max(0.01, ...(retention?.points.map((p) => p.rate) ?? [1])),
    [retention]
  );

  if (!token) return null;

  if (!roleLoading && !isAdmin) {
    return (
      <div className="analytics-container">
        <button className="back-btn" onClick={() => navigate('/')}>
          ← Назад
        </button>
        <div className="analytics-forbidden">
          <h1>🔒 Доступ запрещён</h1>
          <p>Аналитика доступна только администраторам</p>
        </div>
      </div>
    );
  }

  return (
    <div className="analytics-container">
      <button className="back-btn" onClick={() => navigate('/')}>
        ← Назад
      </button>

      <div className="analytics-header">
        <h1>📊 Аналитика</h1>
        <p>DAU, MAU и удержание пользователей (admin)</p>
      </div>

      <div className="analytics-controls">
        <label>
          Период (дней):
          <select value={days} onChange={(e) => setDays(Number(e.target.value))}>
            {[7, 14, 30, 60, 90].map((n) => (
              <option key={n} value={n}>{n}</option>
            ))}
          </select>
        </label>
        <button onClick={loadAnalytics} disabled={loading}>
          🔄 Обновить
        </button>
      </div>

      {error && <div className="analytics-error">{error}</div>}

      {loading || roleLoading ? (
        <div className="analytics-loading">
          <LoadingSpinner />
        </div>
      ) : (
        <>
          <div className="analytics-section">
            <h2>MAU — месячные активные пользователи</h2>
            <div className="mau-value">{mau?.mau ?? 0}</div>
            <div className="mau-sub">Окно: {mau?.window_days ?? days} дней</div>
          </div>

          <div className="analytics-section">
            <h2>DAU — ежедневная активность</h2>
            {dauDays.length === 0 ? (
              <p style={{ color: '#8a94a8' }}>Нет данных</p>
            ) : (
              <table className="analytics-table">
                <thead>
                  <tr>
                    <th>День</th>
                    <th>Пользователи</th>
                  </tr>
                </thead>
                <tbody>
                  {dauDays.map((row) => (
                    <tr key={row.day}>
                      <td>{row.day}</td>
                      <td>
                        <div className="bar-cell">
                          <div className="bar-track">
                            <div
                              className="bar-fill"
                              style={{ width: `${(row.count / maxDau) * 100}%` }}
                            />
                          </div>
                          <span className="bar-value">{row.count}</span>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div className="analytics-section">
            <h2>Retention — удержание</h2>
            {retention ? (
              <>
                <div className="retention-meta">
                  Когорта: {retention.cohort_day} · Размер: {retention.cohort_size}
                </div>
                <table className="analytics-table">
                  <thead>
                    <tr>
                      <th>День N</th>
                      <th>Rate</th>
                    </tr>
                  </thead>
                  <tbody>
                    {retention.points.map((pt) => (
                      <tr key={pt.day_n}>
                        <td>D{pt.day_n}</td>
                        <td>
                          <div className="bar-cell">
                            <div className="bar-track">
                              <div
                                className="bar-fill bar-fill--retention"
                                style={{ width: `${(pt.rate / maxRetention) * 100}%` }}
                              />
                            </div>
                            <span className="bar-value">{(pt.rate * 100).toFixed(1)}%</span>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            ) : (
              <p style={{ color: '#8a94a8' }}>Нет данных по удержанию</p>
            )}
          </div>
        </>
      )}
    </div>
  );
}
