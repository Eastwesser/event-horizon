import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useUserRole } from '../../hooks/useUserRole';
import { authorsApi, type Author, type AuthorSaleRow } from '../../services/authorsApi';
import { inventoryApi, type InventoryItem } from '../../services/inventoryApi';
import { InventoryItemCard } from '../Inventory/InventoryItemCard';
import { InventoryCreateModal } from '../Inventory/InventoryCreateModal';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';
import { Badge } from '../ui/Badge';
import { formatTicketPrice } from '../../lib/formatPrice';

type Tab = 'cards' | 'sales' | 'profile';

const TYPES = ['', 'карточка', 'мерч', 'брелок', 'картина', 'фенечка'] as const;

export function AuthorDashboard() {
  const navigate = useNavigate();
  const { isAuthor, isAdmin, loading: roleLoading } = useUserRole();
  const selfId = localStorage.getItem('userId') || '';

  const [tab, setTab] = useState<Tab>('cards');
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [itemsError, setItemsError] = useState('');
  const [itemsLoading, setItemsLoading] = useState(true);
  const [typeFilter, setTypeFilter] = useState('');
  const [query, setQuery] = useState('');
  const [showDeletedOnly, setShowDeletedOnly] = useState(false);
  const [showCreate, setShowCreate] = useState(false);

  const [sales, setSales] = useState<AuthorSaleRow[]>([]);
  const [salesCount, setSalesCount] = useState(0);
  const [ticketsEarned, setTicketsEarned] = useState(0);
  const [salesLoading, setSalesLoading] = useState(false);
  const [salesError, setSalesError] = useState('');

  const [profile, setProfile] = useState<Author | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileError, setProfileError] = useState('');
  const [profileSaving, setProfileSaving] = useState(false);
  const [displayName, setDisplayName] = useState('');
  const [bio, setBio] = useState('');
  const [portfolio, setPortfolio] = useState('');
  const [profileFlash, setProfileFlash] = useState('');

  useEffect(() => {
    if (!roleLoading && !isAuthor) {
      navigate('/register-author', { replace: true });
    }
  }, [roleLoading, isAuthor, navigate]);

  const loadItems = useCallback(() => {
    if (!selfId && !isAdmin) return;
    setItemsLoading(true);
    setItemsError('');
    void inventoryApi
      .searchAllItems({
        author_id: selfId || undefined,
        include_deleted: 1,
        type: typeFilter || undefined,
        query: query.trim() || undefined,
      })
      .then((res) => setItems(res.items))
      .catch(() => setItemsError('Не удалось загрузить товары'))
      .finally(() => setItemsLoading(false));
  }, [selfId, isAdmin, typeFilter, query]);

  useEffect(() => {
    if (!isAuthor || tab !== 'cards') return;
    loadItems();
  }, [isAuthor, tab, loadItems]);

  const loadSales = useCallback(() => {
    setSalesLoading(true);
    setSalesError('');
    void authorsApi
      .getSales({ limit: 50 })
      .then((res) => {
        setSales(res.purchases);
        setSalesCount(res.sales_count);
        setTicketsEarned(res.tickets_earned);
      })
      .catch(() => setSalesError('Не удалось загрузить продажи'))
      .finally(() => setSalesLoading(false));
  }, []);

  useEffect(() => {
    if (!isAuthor || tab !== 'sales') return;
    loadSales();
  }, [isAuthor, tab, loadSales]);

  const loadProfile = useCallback(() => {
    setProfileLoading(true);
    setProfileError('');
    void authorsApi
      .getMe()
      .then((a) => {
        setProfile(a);
        setDisplayName(a.display_name || '');
        setBio(a.bio || '');
        setPortfolio(a.portfolio || '');
      })
      .catch(() => setProfileError('Профиль ещё не создан — сохраните имя ниже'))
      .finally(() => setProfileLoading(false));
  }, []);

  useEffect(() => {
    if (!isAuthor || tab !== 'profile') return;
    loadProfile();
  }, [isAuthor, tab, loadProfile]);

  const visibleItems = useMemo(() => {
    if (!showDeletedOnly) return items;
    return items.filter((i) => i.deleted || i.attributes?._deleted);
  }, [items, showDeletedOnly]);

  const saveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) return;
    setProfileSaving(true);
    setProfileFlash('');
    try {
      const a = await authorsApi.upsertMe({
        display_name: displayName.trim(),
        bio: bio.trim(),
        portfolio: portfolio.trim(),
        avatar_url: profile?.avatar_url || '',
      });
      setProfile(a);
      setProfileFlash('Сохранено');
    } catch {
      setProfileError('Не удалось сохранить профиль');
    } finally {
      setProfileSaving(false);
    }
  };

  if (roleLoading || !isAuthor) {
    return (
      <PageShell width="wide">
        <div className="flex justify-center py-20">
          <Spinner />
        </div>
      </PageShell>
    );
  }

  return (
    <PageShell width="wide">
      <PageHeader
        title="Кабинет автора"
        subtitle="Ваши товары, продажи и профиль."
        onBack={() => navigate('/')}
        backLabel="На главную"
      />

      <div className="mb-6 flex flex-wrap gap-2">
        {(
          [
            ['cards', 'Мои товары'],
            ['sales', 'Продажи'],
            ['profile', 'Профиль автора'],
          ] as const
        ).map(([id, label]) => (
          <Button
            key={id}
            type="button"
            size="sm"
            variant={tab === id ? 'secondary' : 'ghost'}
            onClick={() => setTab(id)}
          >
            {label}
          </Button>
        ))}
      </div>

      {tab === 'cards' && (
        <div className="flex flex-col gap-4">
          <div className="flex flex-wrap items-end gap-3">
            <label className="flex flex-col gap-1 text-sm text-text-secondary">
              Тип
              <select
                className="rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary"
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
              >
                {TYPES.map((t) => (
                  <option key={t || 'all'} value={t}>
                    {t || 'Все'}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex min-w-[12rem] flex-1 flex-col gap-1 text-sm text-text-secondary">
              Поиск
              <input
                className="rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Название…"
              />
            </label>
            <label className="flex items-center gap-2 pb-2 text-sm text-text-secondary">
              <input
                type="checkbox"
                checked={showDeletedOnly}
                onChange={(e) => setShowDeletedOnly(e.target.checked)}
              />
              Только удалённые
            </label>
            <Button type="button" size="sm" variant="secondary" onClick={() => setShowCreate(true)}>
              Создать
            </Button>
          </div>

          {itemsError && (
            <Card className="border-error/30 bg-error/10 text-sm text-error">{itemsError}</Card>
          )}
          {itemsLoading ? (
            <div className="flex justify-center py-16">
              <Spinner />
            </div>
          ) : visibleItems.length === 0 ? (
            <p className="py-12 text-center text-text-secondary">Пока нет товаров</p>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {visibleItems.map((item) => (
                <InventoryItemCard
                  key={item.id}
                  item={item}
                  softManage
                  onChanged={loadItems}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {tab === 'sales' && (
        <div className="flex flex-col gap-4">
          <div className="flex flex-wrap gap-4">
            <Card className="min-w-[10rem] px-4 py-3">
              <div className="text-xs uppercase tracking-wide text-text-muted">Продажи</div>
              <div className="font-hud text-2xl text-text-primary">{salesCount}</div>
            </Card>
            <Card className="min-w-[10rem] px-4 py-3">
              <div className="text-xs uppercase tracking-wide text-text-muted">Билеты</div>
              <div className="font-hud text-2xl text-horizon-gold">{ticketsEarned}</div>
            </Card>
            <Button type="button" size="sm" variant="ghost" onClick={loadSales}>
              Обновить
            </Button>
          </div>
          {salesError && (
            <Card className="border-error/30 bg-error/10 text-sm text-error">{salesError}</Card>
          )}
          {salesLoading ? (
            <div className="flex justify-center py-16">
              <Spinner />
            </div>
          ) : sales.length === 0 ? (
            <p className="py-12 text-center text-text-secondary">Продаж пока нет</p>
          ) : (
            <div className="overflow-x-auto rounded-md border border-white/10">
              <table className="w-full min-w-[40rem] border-collapse text-left text-sm">
                <thead className="bg-white/5 text-xs uppercase tracking-wide text-text-muted">
                  <tr>
                    <th className="px-3 py-2.5 font-medium">Товар</th>
                    <th className="px-3 py-2.5 font-medium">Покупатель</th>
                    <th className="px-3 py-2.5 font-medium">Цена</th>
                    <th className="px-3 py-2.5 font-medium">Когда</th>
                    <th className="px-3 py-2.5 font-medium">Статус</th>
                  </tr>
                </thead>
                <tbody>
                  {sales.map((row) => (
                    <tr key={row.id} className="border-t border-white/10">
                      <td className="px-3 py-2.5 text-text-primary">
                        {row.item_name || row.item_id}
                      </td>
                      <td className="px-3 py-2.5 text-text-secondary">{row.buyer_email}</td>
                      <td className="px-3 py-2.5 font-hud text-horizon-gold">
                        {formatTicketPrice(row.price)}
                      </td>
                      <td className="px-3 py-2.5 text-text-muted">
                        {row.purchased_at
                          ? new Date(row.purchased_at).toLocaleString('ru-RU')
                          : '—'}
                      </td>
                      <td className="px-3 py-2.5">
                        {row.refunded_at ? (
                          <Badge tone="error">refunded</Badge>
                        ) : (
                          <Badge tone="indigo">{row.status || 'COMPLETED'}</Badge>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {tab === 'profile' && (
        <div className="mx-auto max-w-xl">
          {profileLoading ? (
            <div className="flex justify-center py-16">
              <Spinner />
            </div>
          ) : (
            <form className="flex flex-col gap-4" onSubmit={(e) => void saveProfile(e)}>
              {profileError && (
                <Card className="border-warning/30 bg-warning/10 text-sm text-warning">
                  {profileError}
                </Card>
              )}
              {profileFlash && (
                <Card className="border-photon-cyan/30 bg-photon-cyan/10 text-sm">{profileFlash}</Card>
              )}
              <label className="flex flex-col gap-1 text-sm text-text-secondary">
                Display name
                <input
                  required
                  className="rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                />
              </label>
              <label className="flex flex-col gap-1 text-sm text-text-secondary">
                Portfolio URL
                <input
                  className="rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary"
                  value={portfolio}
                  onChange={(e) => setPortfolio(e.target.value)}
                  placeholder="https://…"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm text-text-secondary">
                Bio
                <textarea
                  className="min-h-[6rem] rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary"
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                />
              </label>
              <div className="flex gap-2">
                <Button type="submit" variant="secondary" disabled={profileSaving}>
                  {profileSaving ? '…' : 'Сохранить'}
                </Button>
                {profile?.user_id && (
                  <Link
                    to={`/authors`}
                    className="inline-flex items-center text-sm text-photon-cyan hover:underline"
                  >
                    К каталогу авторов
                  </Link>
                )}
              </div>
            </form>
          )}
        </div>
      )}

      {showCreate && (
        <InventoryCreateModal
          onClose={() => {
            setShowCreate(false);
            loadItems();
          }}
        />
      )}
    </PageShell>
  );
}
