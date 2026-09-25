// frontend/src/components/Analytics/AnalyticsDashboard.tsx
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { analyticsApi, type DayCount, type RetentionPoint } from '../../services/analyticsApi';
import { useUserRole } from '../../hooks/useUserRole';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roleLoading, isAdmin, days]);

  const maxDau = useMemo(() => Math.max(1, ...dauDays.map((d) => d.count)), [dauDays]);

  const maxRetention = useMemo(
    () => Math.max(0.01, ...(retention?.points.map((p) => p.rate) ?? [1])),
    [retention]
  );

  if (!token) return null;

  if (!roleLoading && !isAdmin) {
    return (
      <PageShell width="wide">
        <PageHeader title="📊 Аналитика" subtitle="DAU, MAU и удержание пользователей (admin)" onBack={() => navigate('/')} />
        <div className="rounded-md border border-error/30 bg-error/10 py-16 text-center">
          <h1 className="mb-2 font-display text-xl font-semibold text-error">🔒 Доступ запрещён</h1>
          <p className="text-sm text-text-secondary">Аналитика доступна только администраторам</p>
        </div>
      </PageShell>
    );
  }

  return (
    <PageShell width="wide">
      <PageHeader title="📊 Аналитика" subtitle="DAU, MAU и удержание пользователей (admin)" onBack={() => navigate('/')} />

      <div className="mb-6 flex flex-wrap items-center gap-3">
        <label className="flex items-center gap-2 text-sm text-text-secondary">
          Период (дней):
          <select
            value={days}
            onChange={(e) => setDays(Number(e.target.value))}
            className="rounded-sm border border-white/10 bg-nebula px-2.5 py-1.5 text-sm text-text-primary focus-visible:outline-none"
          >
            {[7, 14, 30, 60, 90].map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </label>
        <Button variant="secondary" size="sm" onClick={loadAnalytics} disabled={loading}>
          🔄 Обновить
        </Button>
      </div>

      {error && (
        <div role="alert" className="mb-4 rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error">
          {error}
        </div>
      )}

      {loading || roleLoading ? (
        <div className="flex justify-center py-16">
          <Spinner size={48} />
        </div>
      ) : (
        <div className="flex flex-col gap-6">
          <Card>
            <h2 className="mb-3 font-display text-lg font-semibold text-text-primary">
              MAU — месячные активные пользователи
            </h2>
            <div className="font-hud text-4xl font-bold tabular-nums text-success">{mau?.mau ?? 0}</div>
            <div className="mt-1 text-sm text-text-secondary">Окно: {mau?.window_days ?? days} дней</div>
          </Card>

          <Card>
            <h2 className="mb-3 font-display text-lg font-semibold text-text-primary">DAU — ежедневная активность</h2>
            {dauDays.length === 0 ? (
              <p className="text-sm text-text-secondary">Нет данных</p>
            ) : (
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <th className="border-b border-white/10 py-2 text-left text-xs font-medium text-text-secondary">День</th>
                    <th className="border-b border-white/10 py-2 text-left text-xs font-medium text-text-secondary">
                      Пользователи
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {dauDays.map((row) => (
                    <tr key={row.day}>
                      <td className="border-b border-white/10 py-2 text-sm text-text-primary">{row.day}</td>
                      <td className="border-b border-white/10 py-2">
                        <div className="flex items-center gap-3">
                          <div className="h-5 min-w-20 flex-1 overflow-hidden rounded-sm bg-void">
                            <div
                              className="h-full rounded-sm bg-gradient-to-r from-photon-cyan to-horizon-gold transition-[width] duration-300"
                              style={{ width: `${(row.count / maxDau) * 100}%` }}
                            />
                          </div>
                          <span className="min-w-10 text-right font-hud text-sm font-semibold tabular-nums text-text-primary">
                            {row.count}
                          </span>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </Card>

          <Card>
            <h2 className="mb-3 font-display text-lg font-semibold text-text-primary">Retention — удержание</h2>
            {retention ? (
              <>
                <div className="mb-4 text-sm text-text-secondary">
                  Когорта: {retention.cohort_day} · Размер: {retention.cohort_size}
                </div>
                <table className="w-full border-collapse">
                  <thead>
                    <tr>
                      <th className="border-b border-white/10 py-2 text-left text-xs font-medium text-text-secondary">
                        День N
                      </th>
                      <th className="border-b border-white/10 py-2 text-left text-xs font-medium text-text-secondary">
                        Rate
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {retention.points.map((pt) => (
                      <tr key={pt.day_n}>
                        <td className="border-b border-white/10 py-2 text-sm text-text-primary">D{pt.day_n}</td>
                        <td className="border-b border-white/10 py-2">
                          <div className="flex items-center gap-3">
                            <div className="h-5 min-w-20 flex-1 overflow-hidden rounded-sm bg-void">
                              <div
                                className="h-full rounded-sm bg-gradient-to-r from-success to-warning transition-[width] duration-300"
                                style={{ width: `${(pt.rate / maxRetention) * 100}%` }}
                              />
                            </div>
                            <span className="min-w-10 text-right font-hud text-sm font-semibold tabular-nums text-text-primary">
                              {(pt.rate * 100).toFixed(1)}%
                            </span>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            ) : (
              <p className="text-sm text-text-secondary">Нет данных по удержанию</p>
            )}
          </Card>
        </div>
      )}
    </PageShell>
  );
}
