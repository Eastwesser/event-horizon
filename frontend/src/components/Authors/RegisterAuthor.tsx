import { useEffect, useState } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { authorsApi, type AuthorApplication } from '../../services/authorsApi';
import { useUserRole } from '../../hooks/useUserRole';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';

const inputClass =
  'rounded-md border border-white/10 bg-void px-4 py-3 text-base text-text-primary placeholder:text-text-muted focus:border-photon-cyan/60 focus-visible:outline-none';

export function RegisterAuthor() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { isAuthor, isAdmin, loading: roleLoading } = useUserRole();

  const [displayName, setDisplayName] = useState('');
  const [portfolio, setPortfolio] = useState('');
  const [motivation, setMotivation] = useState('');
  const [contactEmail, setContactEmail] = useState(
    () => localStorage.getItem('userEmail') || '',
  );
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [checking, setChecking] = useState(true);
  const [application, setApplication] = useState<AuthorApplication | null>(null);

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setChecking(true);
      try {
        const app = await authorsApi.getMyApplication();
        if (!cancelled) setApplication(app);
      } catch {
        if (!cancelled) setApplication(null);
      } finally {
        if (!cancelled) setChecking(false);
      }
    };
    if (token) void load();
    return () => {
      cancelled = true;
    };
  }, [token]);

  if (!token) return null;
  if (roleLoading || checking) {
    return (
      <PageShell>
        <div className="flex min-h-[40vh] items-center justify-center">
          <Spinner size={48} />
        </div>
      </PageShell>
    );
  }
  if (isAuthor || isAdmin) {
    return <Navigate to="/authors" replace />;
  }

  const pending = application?.status === 'pending';
  const submitted = pending || application?.status === 'approved';

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const app = await authorsApi.apply({
        display_name: displayName.trim(),
        portfolio: portfolio.trim(),
        motivation: motivation.trim(),
        contact_email: contactEmail.trim() || undefined,
      });
      setApplication(app);
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string; message?: string } } };
      const msg = ax.response?.data?.error || ax.response?.data?.message || 'Не удалось отправить заявку';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageShell>
      <PageHeader title="Стать автором" onBack={() => navigate('/authors')} backLabel="К авторам" />

      {submitted ? (
        <div className="mx-auto mt-8 max-w-lg rounded-lg border border-white/10 bg-nebula/80 p-8 text-center">
          <h2 className="font-display text-2xl font-semibold text-text-primary">Заявка отправлена</h2>
          <p className="mt-3 text-text-secondary">
            Ожидайте решения администрации. Мы свяжемся с вами после рассмотрения.
          </p>
          {application && (
            <p className="mt-4 text-sm text-text-muted">
              Статус: <span className="text-horizon-gold">{application.status}</span>
              {application.display_name ? ` · ${application.display_name}` : ''}
            </p>
          )}
          <div className="mt-8 flex flex-wrap justify-center gap-3">
            <Button variant="secondary" size="sm" onClick={() => navigate('/authors')}>
              К списку авторов
            </Button>
            <Button variant="ghost" size="sm" onClick={() => navigate('/')}>
              На главную
            </Button>
          </div>
        </div>
      ) : (
        <form
          onSubmit={handleSubmit}
          className="mx-auto mt-8 flex max-w-lg flex-col gap-5 rounded-lg border border-white/10 bg-nebula/80 p-8"
          noValidate
        >
          <p className="text-sm text-text-secondary">
            Заполните заявку. После одобрения администратором вы получите роль автора.
          </p>

          <label className="flex flex-col gap-2 text-sm font-medium text-text-secondary">
            Отображаемое имя
            <input
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              required
              maxLength={80}
              className={inputClass}
              placeholder="Как вас будут видеть в каталоге"
            />
          </label>

          <label className="flex flex-col gap-2 text-sm font-medium text-text-secondary">
            Портфолио (URL или текст)
            <textarea
              value={portfolio}
              onChange={(e) => setPortfolio(e.target.value)}
              rows={3}
              maxLength={2000}
              className={inputClass}
              placeholder="Ссылка на работы или краткое описание опыта"
            />
          </label>

          <label className="flex flex-col gap-2 text-sm font-medium text-text-secondary">
            Мотивация
            <textarea
              value={motivation}
              onChange={(e) => setMotivation(e.target.value)}
              required
              rows={4}
              maxLength={4000}
              className={inputClass}
              placeholder="Почему хотите публиковать карты в Event Horizon"
            />
          </label>

          <label className="flex flex-col gap-2 text-sm font-medium text-text-secondary">
            Контактный email
            <input
              type="email"
              value={contactEmail}
              onChange={(e) => setContactEmail(e.target.value)}
              required
              className={inputClass}
            />
          </label>

          {error && (
            <div role="alert" className="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
              {error}
            </div>
          )}

          <Button type="submit" disabled={loading} className="w-full">
            {loading ? 'Отправка…' : 'Отправить заявку'}
          </Button>

          <p className="text-center text-sm text-text-muted">
            Уже автор?{' '}
            <Link to="/authors" className="text-indigo-soft hover:underline">
              Перейти к профилю автора
            </Link>
          </p>
        </form>
      )}
    </PageShell>
  );
}
