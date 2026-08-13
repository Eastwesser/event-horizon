// frontend/src/components/Authors/AuthorsPage.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authorsApi, type Author } from '../../services/authorsApi';
import { useUserRole } from '../../hooks/useUserRole';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import './AuthorsPage.css';

const PAGE_SIZE = 20;

export function AuthorsPage() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');
  const { isAuthor, loading: roleLoading } = useUserRole();

  const [authors, setAuthors] = useState<Author[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

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
      setAuthors((prev) => (append ? [...prev, ...data.authors] : data.authors));
      setTotal(data.total);
    } catch {
      setError('Не удалось загрузить список авторов');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  useEffect(() => {
    loadAuthors();
  }, []);

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
        (err as { response?: { data?: { error?: string }; status?: number } })?.response?.data?.error ||
        'Ошибка сохранения профиля';
      setError(msg);
    } finally {
      setSaving(false);
    }
  };

  if (!token) return null;

  return (
    <div className="authors-container">
      <button className="back-btn" onClick={() => navigate('/')}>
        ← Назад
      </button>

      <div className="authors-header">
        <h1>✍️ Авторы</h1>
        <p>Авторы контента и сообщества EventHorizon</p>
      </div>

      {error && <div className="authors-error">{error}</div>}
      {success && <div className="authors-success">{success}</div>}

      {!roleLoading && isAuthor && (
        <form className="authors-form" onSubmit={handleSaveProfile}>
          <h2>Мой профиль автора</h2>
          <label>
            Отображаемое имя *
            <input
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Ваше имя"
              required
            />
          </label>
          <label>
            О себе
            <textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Короткая биография"
            />
          </label>
          <label>
            URL аватара
            <input
              value={avatarUrl}
              onChange={(e) => setAvatarUrl(e.target.value)}
              placeholder="https://..."
              type="url"
            />
          </label>
          <button type="submit" disabled={saving}>
            {saving ? 'Сохранение...' : '💾 Сохранить профиль'}
          </button>
        </form>
      )}

      <div className="authors-list">
        <h2>Авторы ({total})</h2>
        {loading ? (
          <LoadingSpinner />
        ) : authors.length === 0 ? (
          <div className="authors-empty">Пока нет зарегистрированных авторов</div>
        ) : (
          <>
            {authors.map((author) => (
              <div key={author.id} className="author-card">
                <div className="author-avatar">
                  {author.avatar_url ? (
                    <img src={author.avatar_url} alt={author.display_name} />
                  ) : (
                    '✍️'
                  )}
                </div>
                <div className="author-info">
                  <h3>{author.display_name}</h3>
                  {author.bio && <p>{author.bio}</p>}
                  <div className="author-meta">
                    {author.active ? '🟢 Активен' : '⚫ Неактивен'}
                    {' · '}
                    Обновлён: {new Date(author.updated_at_unix * 1000).toLocaleDateString('ru-RU')}
                  </div>
                </div>
              </div>
            ))}
            {authors.length < total && (
              <button
                className="authors-load-more"
                disabled={loadingMore}
                onClick={() => loadAuthors(authors.length, true)}
              >
                {loadingMore ? 'Загрузка...' : 'Загрузить ещё'}
              </button>
            )}
          </>
        )}
      </div>
    </div>
  );
}
