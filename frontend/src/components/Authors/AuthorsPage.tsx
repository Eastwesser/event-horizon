// frontend/src/components/Authors/AuthorsPage.tsx
import { useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authorsApi, type Author } from '../../services/authorsApi';
import { inventoryApi } from '../../services/inventoryApi';
import { collectCardArtists, type CardArtist } from '../../lib/cardArtists';
import { pluralCards } from '../../lib/pluralize';
import { useUserRole } from '../../hooks/useUserRole';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';
import { Icon, IconLabel } from '../ui/Icon';

const PAGE_SIZE = 20;
const TOP_N = 24;

export function AuthorsPage() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { isAuthor, loading: roleLoading } = useUserRole();

  const [authors, setAuthors] = useState<Author[]>([]);
  const [total, setTotal] = useState(0);
  const [cardArtists, setCardArtists] = useState<CardArtist[]>([]);
  const [loading, setLoading] = useState(true);
  const [cardsLoading, setCardsLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [artistQuery, setArtistQuery] = useState('');
  const [showAllArtists, setShowAllArtists] = useState(false);

  const [displayName, setDisplayName] = useState('');
  const [bio, setBio] = useState('');
  const [avatarUrl, setAvatarUrl] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!token) navigate('/login');
  }, [token, navigate]);

  const loadAuthors = async (offset = 0, append = false) => {
    try {
      if (append) setLoadingMore(true);
      else setLoading(true);
      const data = await authorsApi.list(PAGE_SIZE, offset);
      const list = data.authors ?? [];
      setAuthors((prev) => (append ? [...prev, ...list] : list));
      setTotal(data.total ?? 0);
    } catch {
      setError('Не удалось загрузить список авторов');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  useEffect(() => {
    loadAuthors();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setCardsLoading(true);
      try {
        const res = await inventoryApi.searchAllItems({ type: 'карточка' });
        if (!cancelled) setCardArtists(collectCardArtists(res.items ?? []));
      } catch {
        if (!cancelled) setCardArtists([]);
      } finally {
        if (!cancelled) setCardsLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  const filteredArtists = useMemo(() => {
    const q = artistQuery.trim().toLowerCase();
    if (!q) return cardArtists;
    return cardArtists.filter((a) => a.display_name.toLowerCase().includes(q));
  }, [cardArtists, artistQuery]);

  const visibleArtists = showAllArtists
    ? filteredArtists
    : filteredArtists.slice(0, TOP_N);

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) return;

    setSaving(true);
    setError('');
    setSuccess('');
    try {
      await authorsApi.upsertMe({
        display_name: displayName.trim(),
        bio: bio.trim() || undefined,
        avatar_url: avatarUrl.trim() || undefined,
      });
      setSuccess('✅ Профиль автора сохранён');
      loadAuthors();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string }; status?: number } })?.response?.data
          ?.error || 'Ошибка сохранения профиля';
      setError(msg);
    } finally {
      setSaving(false);
    }
  };

  if (!token) return null;

  const inputClasses =
    'mt-1.5 block w-full rounded-sm border border-white/10 bg-void px-3.5 py-2.5 text-base text-text-primary placeholder:text-text-muted focus:border-photon-cyan/60 focus-visible:outline-none';

  return (
    <PageShell width="wide">
      <PageHeader
        title={<IconLabel name="pen" iconClassName="h-7 w-7 text-horizon-gold">Авторы</IconLabel>}
        subtitle="Художники карт и авторы сообщества Event Horizon"
        onBack={() => navigate('/')}
      />

      {!roleLoading && !isAuthor && (
        <div className="mb-6 flex flex-wrap items-center justify-between gap-3 rounded-md border border-horizon-gold/25 bg-horizon-gold/5 px-4 py-3">
          <p className="text-sm text-text-secondary">
            Хотите публиковать карты? Отправьте заявку на роль автора.
          </p>
          <Button variant="secondary" size="sm" onClick={() => navigate('/register-author')}>
            Стать автором
          </Button>
        </div>
      )}

      {error && (
        <div
          role="alert"
          className="mb-4 rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error"
        >
          {error}
        </div>
      )}
      {success && (
        <div
          role="status"
          className="mb-4 rounded-sm border border-success/30 bg-success/10 px-4 py-3 text-sm text-success"
        >
          {success}
        </div>
      )}

      <section className="mb-10">
        <div className="mb-3 flex flex-wrap items-end justify-between gap-3">
          <h2 className="font-display text-lg font-semibold text-text-primary">
            Художники карт ({cardArtists.length})
          </h2>
          <input
            type="search"
            placeholder="Поиск по имени…"
            value={artistQuery}
            onChange={(e) => {
              setArtistQuery(e.target.value);
              setShowAllArtists(false);
            }}
            className="w-full max-w-xs rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-photon-cyan/50"
          />
        </div>
        <p className="mb-3 text-xs text-text-muted">
          Соавторы на одной карте: ссылка ведёт на первого artist_id, в названии —
          полное display-имя (временно, до multi-author страниц).
        </p>
        {cardsLoading ? (
          <div className="flex justify-center py-12">
            <Spinner size={40} />
          </div>
        ) : filteredArtists.length === 0 ? (
          <div className="rounded-md border border-white/10 bg-nebula py-10 text-center text-text-secondary">
            {artistQuery.trim() ? 'Никого не найдено' : 'Пока нет карт с художниками'}
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-2">
              {visibleArtists.map((a) => (
                <Link
                  key={a.artist_id}
                  to={`/authors/${encodeURIComponent(a.artist_id)}`}
                  className="block rounded-md border border-white/10 bg-nebula px-4 py-3 text-inherit no-underline transition hover:border-photon-cyan/40"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="font-display font-semibold text-text-primary">
                      {a.display_name}
                    </span>
                    <span className="font-hud text-sm tabular-nums text-text-muted">
                      {pluralCards(a.count)}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
            {!showAllArtists && filteredArtists.length > TOP_N ? (
              <div className="mt-4 flex justify-center">
                <Button variant="ghost" onClick={() => setShowAllArtists(true)}>
                  Показать всех ({filteredArtists.length})
                </Button>
              </div>
            ) : null}
          </>
        )}
      </section>

      {!roleLoading && isAuthor && (
        <Card glow className="mb-8">
          <h2 className="mb-4 font-display text-lg font-semibold text-text-primary">
            Мой профиль автора
          </h2>
          <form onSubmit={handleSaveProfile} className="flex flex-col gap-4">
            <label className="text-sm font-medium text-text-secondary">
              Отображаемое имя *
              <input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="Ваше имя"
                required
                className={inputClasses}
              />
            </label>
            <label className="text-sm font-medium text-text-secondary">
              О себе
              <textarea
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                placeholder="Короткая биография"
                className={`${inputClasses} min-h-[80px] resize-y`}
              />
            </label>
            <label className="text-sm font-medium text-text-secondary">
              URL аватара
              <input
                value={avatarUrl}
                onChange={(e) => setAvatarUrl(e.target.value)}
                placeholder="https://..."
                type="url"
                className={inputClasses}
              />
            </label>
            <Button type="submit" disabled={saving} className="self-start">
              {saving ? 'Сохранение…' : 'Сохранить профиль'}
            </Button>
          </form>
        </Card>
      )}

      <div>
        <h2 className="mb-4 font-display text-lg font-semibold text-text-primary">
          Авторы сообщества ({total})
        </h2>
        {loading ? (
          <div className="flex justify-center py-16">
            <Spinner size={48} />
          </div>
        ) : (authors?.length ?? 0) === 0 ? (
          <div className="rounded-md border border-white/10 bg-nebula py-12 text-center text-text-secondary">
            Пока нет зарегистрированных авторов
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-3">
              {authors.map((author) => (
                <Card key={author.id} className="flex gap-4 p-4">
                  <div className="flex h-14 w-14 shrink-0 items-center justify-center overflow-hidden rounded-full bg-indigo/15 text-2xl">
                    {author.avatar_url ? (
                      <img
                        src={author.avatar_url}
                        alt={author.display_name}
                        className="h-full w-full object-cover"
                        loading="lazy"
                        decoding="async"
                      />
                    ) : (
                      <Icon name="pen" className="h-6 w-6 text-indigo-soft" />
                    )}
                  </div>
                  <div className="min-w-0">
                    <h3 className="font-display text-base font-semibold text-text-primary">
                      {author.display_name}
                    </h3>
                    {author.bio && (
                      <p className="mt-1 text-sm leading-relaxed text-text-secondary">
                        {author.bio}
                      </p>
                    )}
                    <div className="mt-1.5 text-xs text-text-muted">
                      {author.active ? 'Активен' : 'Неактивен'}
                      {' · '}
                      Обновлён:{' '}
                      {new Date(author.updated_at_unix * 1000).toLocaleDateString('ru-RU')}
                    </div>
                  </div>
                </Card>
              ))}
            </div>
            {authors.length < total && (
              <div className="mt-4 flex justify-center">
                <Button
                  variant="ghost"
                  disabled={loadingMore}
                  onClick={() => loadAuthors(authors.length, true)}
                >
                  {loadingMore ? 'Загрузка…' : 'Загрузить ещё'}
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </PageShell>
  );
}
