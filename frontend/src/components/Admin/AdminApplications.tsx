import { useEffect, useState } from 'react';
import { adminApi, type AuthorApplication, type ApplicationStatusFilter } from '../../services/adminApi';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';

const PAGE_SIZE = 50;
const FILTERS: { id: ApplicationStatusFilter; label: string }[] = [
  { id: 'pending', label: 'Pending' },
  { id: 'approved', label: 'Approved' },
  { id: 'rejected', label: 'Rejected' },
];

function statusTone(status: string): 'gold' | 'indigo' | 'neutral' | 'error' {
  if (status === 'approved') return 'indigo';
  if (status === 'rejected') return 'error';
  if (status === 'pending') return 'gold';
  return 'neutral';
}

function formatUnix(sec?: number): string {
  if (!sec) return '—';
  return new Date(sec * 1000).toLocaleString('ru-RU');
}

export function AdminApplications() {
  const [status, setStatus] = useState<ApplicationStatusFilter>('pending');
  const [rows, setRows] = useState<AuthorApplication[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [busyId, setBusyId] = useState<string | null>(null);
  const [flash, setFlash] = useState('');

  const load = () => {
    setLoading(true);
    setError('');
    return adminApi
      .listApplications({ status, limit: PAGE_SIZE, offset })
      .then((res) => {
        setRows(res.applications);
        setTotal(res.total);
      })
      .catch((err: unknown) => {
        const code = (err as { response?: { status?: number } })?.response?.status;
        setError(code === 403 ? 'Доступ запрещён' : 'Не удалось загрузить заявки');
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError('');
    void adminApi
      .listApplications({ status, limit: PAGE_SIZE, offset })
      .then((res) => {
        if (cancelled) return;
        setRows(res.applications);
        setTotal(res.total);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const code = (err as { response?: { status?: number } })?.response?.status;
        setError(code === 403 ? 'Доступ запрещён' : 'Не удалось загрузить заявки');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [status, offset]);

  const approve = async (app: AuthorApplication) => {
    const ok = window.confirm(
      `Одобрить заявку ${app.display_name} (${app.contact_email})?\n\nРоль станет author. Пользователю может понадобиться re-login.`,
    );
    if (!ok) return;
    setBusyId(app.id);
    setFlash('');
    try {
      await adminApi.approveApplication(app.id);
      setFlash(`Одобрено: ${app.display_name}`);
      await load();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string }; status?: number } })?.response?.data
          ?.error || 'Не удалось одобрить';
      window.alert(msg);
    } finally {
      setBusyId(null);
    }
  };

  const reject = async (app: AuthorApplication) => {
    const note = window.prompt('Причина отклонения (необязательно):', '') ?? null;
    if (note === null) return;
    setBusyId(app.id);
    setFlash('');
    try {
      await adminApi.rejectApplication(app.id, note.trim() || undefined);
      setFlash(`Отклонено: ${app.display_name}`);
      await load();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Не удалось отклонить';
      window.alert(msg);
    } finally {
      setBusyId(null);
    }
  };

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        {FILTERS.map((f) => (
          <Button
            key={f.id}
            type="button"
            size="sm"
            variant={status === f.id ? 'secondary' : 'ghost'}
            onClick={() => {
              setOffset(0);
              setStatus(f.id);
            }}
          >
            {f.label}
          </Button>
        ))}
      </div>

      {flash && (
        <Card className="border-photon-cyan/30 bg-photon-cyan/10 text-sm text-text-primary">{flash}</Card>
      )}
      {error && <Card className="border-error/30 bg-error/10 text-sm text-error">{error}</Card>}

      {loading ? (
        <div className="flex justify-center py-16">
          <Spinner />
        </div>
      ) : rows.length === 0 ? (
        <p className="py-12 text-center text-text-secondary">Заявок нет</p>
      ) : (
        <div className="overflow-x-auto rounded-md border border-white/10">
          <table className="w-full min-w-[56rem] border-collapse text-left text-sm">
            <thead className="bg-white/5 text-xs uppercase tracking-wide text-text-muted">
              <tr>
                <th className="px-3 py-2.5 font-medium">Заявитель</th>
                <th className="px-3 py-2.5 font-medium">Портфолио</th>
                <th className="px-3 py-2.5 font-medium">Мотивация</th>
                <th className="px-3 py-2.5 font-medium">Подана</th>
                <th className="px-3 py-2.5 font-medium">Статус</th>
                <th className="px-3 py-2.5 font-medium">Действие</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((app) => (
                <tr key={app.id} className="border-t border-white/10 align-top">
                  <td className="px-3 py-2.5">
                    <div className="font-medium text-text-primary">{app.display_name}</div>
                    <div className="text-xs text-text-secondary">{app.contact_email}</div>
                    <div className="font-hud text-xs text-text-muted">{app.user_id}</div>
                  </td>
                  <td className="max-w-[12rem] px-3 py-2.5 text-text-secondary">
                    {app.portfolio ? (
                      <a
                        href={app.portfolio.startsWith('http') ? app.portfolio : undefined}
                        className="break-all text-photon-cyan hover:underline"
                        target="_blank"
                        rel="noreferrer"
                      >
                        {app.portfolio}
                      </a>
                    ) : (
                      '—'
                    )}
                  </td>
                  <td className="max-w-[16rem] px-3 py-2.5 text-text-secondary">
                    <span className="line-clamp-3">{app.motivation || '—'}</span>
                    {app.reviewer_note ? (
                      <div className="mt-1 text-xs text-text-muted">Заметка: {app.reviewer_note}</div>
                    ) : null}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-text-muted">
                    {formatUnix(app.created_at_unix)}
                  </td>
                  <td className="px-3 py-2.5">
                    <Badge tone={statusTone(app.status)}>{app.status}</Badge>
                  </td>
                  <td className="px-3 py-2.5">
                    {app.status === 'pending' ? (
                      <div className="flex flex-wrap gap-2">
                        <Button
                          type="button"
                          size="sm"
                          variant="secondary"
                          disabled={busyId === app.id}
                          onClick={() => void approve(app)}
                        >
                          Approve
                        </Button>
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          disabled={busyId === app.id}
                          onClick={() => void reject(app)}
                        >
                          Reject
                        </Button>
                      </div>
                    ) : (
                      <span className="text-xs text-text-muted">
                        {app.reviewed_at_unix ? formatUnix(app.reviewed_at_unix) : '—'}
                      </span>
                    )}
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
  );
}
