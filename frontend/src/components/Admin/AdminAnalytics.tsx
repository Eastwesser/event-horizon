import { useEffect, useState } from 'react';
import {
  analyticsApi,
  type DayCount,
  type RetentionPoint,
} from '../../services/analyticsApi';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { Spinner } from '../ui/Spinner';
import { StatCard } from '../ui/StatCard';

const DAY_OPTIONS = [7, 14, 30, 60, 90] as const;

function BarRow({
  label,
  value,
  max,
  display,
  barClass,
}: {
  label: string;
  value: number;
  max: number;
  display: string;
  barClass: string;
}) {
  const pct = max > 0 ? Math.min(100, (value / max) * 100) : 0;
  return (
    <li className="flex items-center gap-3 border-b border-white/5 py-2 last:border-0">
      <span className="w-24 shrink-0 truncate text-sm text-text-secondary" title={label}>
        {label}
      </span>
      <div className="h-4 min-w-0 flex-1 overflow-hidden rounded-sm bg-void">
        <div className={`h-full rounded-sm ${barClass}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="min-w-12 text-right font-hud text-sm tabular-nums text-text-primary">
        {display}
      </span>
    </li>
  );
}

export function AdminAnalytics() {
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
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError('');
    void Promise.all([
      analyticsApi.dau(days),
      analyticsApi.mau(days),
      analyticsApi.retention(7, 7),
    ])
      .then(([dauData, mauData, retentionData]) => {
        if (cancelled) return;
        setDauDays(dauData.days ?? []);
        setMau(mauData);
        setRetention({
          ...retentionData,
          points: retentionData.points ?? [],
        });
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const status = (err as { response?: { status?: number } })?.response?.status;
        setError(status === 403 ? 'Доступ запрещён' : 'Не удалось загрузить аналитику');
        setDauDays([]);
        setMau(null);
        setRetention(null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [days, reloadKey]);

  const latestDau = dauDays.length > 0 ? dauDays[dauDays.length - 1]!.count : 0;
  const avgDau =
    dauDays.length > 0
      ? Math.round(dauDays.reduce((s, d) => s + d.count, 0) / dauDays.length)
      : 0;
  const maxDau = Math.max(1, ...dauDays.map((d) => d.count));
  const maxRetention = Math.max(0.01, ...(retention?.points.map((p) => p.rate) ?? [1]));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-medium text-text-primary">Аналитика</h2>
        <div className="flex flex-wrap items-center gap-2">
          <label className="flex items-center gap-2 text-sm text-text-secondary">
            Период
            <select
              value={days}
              onChange={(e) => setDays(Number(e.target.value))}
              className="rounded-sm border border-white/10 bg-nebula px-2.5 py-1.5 text-sm text-text-primary"
            >
              {DAY_OPTIONS.map((n) => (
                <option key={n} value={n}>
                  {n} дн.
                </option>
              ))}
            </select>
          </label>
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
      ) : (
        <>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <StatCard value={mau?.mau ?? 0} label="MAU" sub={`окно ${mau?.window_days ?? days} дн.`} tone="indigo" />
            <StatCard value={latestDau} label="DAU (последний день)" tone="cyan" />
            <StatCard value={avgDau} label="DAU (среднее)" sub={`${days} дн.`} tone="gold" />
          </div>

          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Card className="flex flex-col gap-2 p-4">
              <h3 className="text-sm font-medium text-text-primary">DAU по дням</h3>
              {dauDays.length === 0 ? (
                <p className="py-6 text-center text-sm text-text-muted">Нет данных</p>
              ) : (
                <ul className="max-h-80 overflow-y-auto">
                  {[...dauDays].reverse().map((row) => (
                    <BarRow
                      key={row.day}
                      label={row.day}
                      value={row.count}
                      max={maxDau}
                      display={String(row.count)}
                      barClass="bg-gradient-to-r from-photon-cyan to-indigo"
                    />
                  ))}
                </ul>
              )}
            </Card>

            <Card className="flex flex-col gap-2 p-4">
              <div className="flex flex-wrap items-baseline justify-between gap-2">
                <h3 className="text-sm font-medium text-text-primary">Retention</h3>
                {retention && (
                  <span className="text-xs text-text-muted">
                    когорта {retention.cohort_day} ·{' '}
                    <Badge tone="neutral">{retention.cohort_size}</Badge>
                  </span>
                )}
              </div>
              {!retention || retention.points.length === 0 ? (
                <p className="py-6 text-center text-sm text-text-muted">Нет данных</p>
              ) : (
                <ul>
                  {retention.points.map((pt) => (
                    <BarRow
                      key={pt.day_n}
                      label={`D${pt.day_n}`}
                      value={pt.rate}
                      max={maxRetention}
                      display={`${(pt.rate * 100).toFixed(1)}%`}
                      barClass="bg-gradient-to-r from-success to-horizon-gold"
                    />
                  ))}
                </ul>
              )}
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
