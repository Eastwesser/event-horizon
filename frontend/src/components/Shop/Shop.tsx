// frontend/src/components/Shop/Shop.tsx
import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import ShopItemCard from './ShopItemCard';
import PurchaseModal from './PurchaseModal';
import CancelPurchaseModal from './CancelPurchaseModal';
import Notification from '../Common/Notification/Notification';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { paymentApi } from '../../services/paymentApi';
import { inventoryApi } from '../../services/inventoryApi';
import { useShopStore, type ShopItem, type PurchasedItem } from '../../store/shopStore';
import { inventoryToShopItem } from '../../lib/shopItemMap';
import { formatTicketAmount } from '../../lib/formatPrice';
import { Card } from '../ui/Card';
import {
  CATALOG_PAGE_SIZE,
  CATALOG_SORT_OPTIONS,
  FILTER_URL_KEYS,
  type CatalogSort,
  collectCatalogFacets,
  filterCatalogItems,
  paginateItems,
  parseCatalogFilters,
  parseCatalogPage,
  parseCatalogSort,
  sortCatalogItems,
} from '../../lib/catalogQuery';
import { saveCatalogNav } from '../../lib/catalogNav';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';
import { CatalogPager } from '../ui/CatalogPager';
import { CardImage } from '../ui/CardImage';
import { CatalogFiltersPanel } from './CatalogFiltersPanel';
import { Icon, IconLabel, type IconName } from '../ui/Icon';
import { itemFallbackIcon } from '../../lib/itemIcons';

function isMerchItem(item: ShopItem): boolean {
  const cat = (item.category || '').toLowerCase();
  const type = (item.type || '').toLowerCase();
  return (
    cat.includes('merch') ||
    type.includes('merch') ||
    type === 'карточка' ||
    type === 'брелок' ||
    type === 'картина' ||
    type === 'фенечка'
  );
}

const itemTypes: { value: string; label: string; icon?: IconName }[] = [
  { value: 'all', label: 'Все' },
  { value: 'карточка', label: 'Карточки', icon: 'cards' },
  { value: 'game_skin', label: 'Скины', icon: 'palette' },
  { value: 'profile_theme', label: 'Темы', icon: 'palette' },
  { value: 'merch', label: 'Мерч', icon: 'gift' },
  { value: 'брелок', label: 'Брелок', icon: 'key' },
  { value: 'картина', label: 'Картина', icon: 'frame' },
  { value: 'фенечка', label: 'Фенечка', icon: 'sparkle' },
];

export const Shop: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const {
    inventory,
    balance,
    loading,
    error,
    fetchItems,
    fetchInventory,
    fetchBalance,
    clearError,
    catalogPatches,
    cancelPurchase,
    cancelling,
  } = useShopStore();

  const [catalog, setCatalog] = useState<ShopItem[]>([]);
  const [catalogLoading, setCatalogLoading] = useState(true);
  const [selectedItem, setSelectedItem] = useState<ShopItem | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<PurchasedItem | null>(null);
  const [activeTab, setActiveTab] = useState<'shop' | 'inventory'>('shop');

  useEffect(() => {
    if (searchParams.get('tab') === 'inventory') {
      setActiveTab('inventory');
    }
  }, [searchParams]);
  const [notification, setNotification] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
    link?: { label: string; path: string };
  } | null>(null);
  const [merchAllowed, setMerchAllowed] = useState<boolean | null>(null);
  const [merchBlockReason, setMerchBlockReason] = useState('');

  const filterType = searchParams.get('type') || 'all';
  const sort = parseCatalogSort(searchParams.get('sort'));
  const page = parseCatalogPage(searchParams.get('page'));
  const catalogFilters = useMemo(
    () => parseCatalogFilters(searchParams),
    [searchParams]
  );
  const [qDraft, setQDraft] = useState(catalogFilters.query || '');

  useEffect(() => {
    setQDraft(catalogFilters.query || '');
  }, [catalogFilters.query]);

  const setQuery = (patch: Record<string, string | null>, replace = false) => {
    const next = new URLSearchParams(searchParams);
    for (const [k, v] of Object.entries(patch)) {
      if (
        v === null ||
        v === '' ||
        (k === 'page' && v === '1') ||
        (k === 'sort' && v === 'newest') ||
        (k === 'type' && v === 'all')
      ) {
        if (k === 'type' && v === 'all') next.delete('type');
        else if (k === 'sort' && v === 'newest') next.delete('sort');
        else if (k === 'page' && (v === '1' || v === null)) next.delete('page');
        else if (v === null || v === '') next.delete(k);
        else next.set(k, v);
      } else {
        next.set(k, v);
      }
    }
    setSearchParams(next, { replace });
  };

  useEffect(() => {
    const t = window.setTimeout(() => {
      const next = qDraft.trim();
      if (next === (catalogFilters.query || '').trim()) return;
      setQuery({ q: next || null, page: '1' }, true);
    }, 300);
    return () => window.clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [qDraft]);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token) return;
    fetchItems();
    fetchBalance();
    fetchInventory();
  }, [fetchItems, fetchBalance, fetchInventory]);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setCatalogLoading(true);
      try {
        const ownedIds = new Set(inventory.map((p) => p.item_id));
        const response = await inventoryApi.searchAllItems();
        if (cancelled) return;
        const patches = useShopStore.getState().catalogPatches;
        const mapped = (response.items ?? []).map((item) => {
          const base = inventoryToShopItem(item, ownedIds.has(item.id));
          const patch = patches[item.id];
          if (!patch) return base;
          const stock =
            patch.stock !== undefined ? patch.stock : base.stock;
          return {
            ...base,
            owned: patch.owned ?? base.owned,
            stock,
            available: (stock ?? 1) > 0,
          };
        });
        setCatalog(mapped);
      } catch (e) {
        console.error('Failed to load inventory catalog for shop', e);
        if (!cancelled) setCatalog([]);
      } finally {
        if (!cancelled) setCatalogLoading(false);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, [inventory]);

  const facets = useMemo(() => collectCatalogFacets(catalog), [catalog]);

  /** Enrich shop inventory rows with catalog images (inventory svc), shop.image_url is often empty. */
  const catalogById = useMemo(() => {
    const m = new Map<string, ShopItem>();
    for (const it of catalog) m.set(it.id, it);
    return m;
  }, [catalog]);

  const filteredSorted = useMemo(() => {
    const patched = catalog.map((item) => {
      const patch = catalogPatches[item.id];
      if (!patch) return item;
      const stock = patch.stock !== undefined ? patch.stock : item.stock;
      return {
        ...item,
        owned: patch.owned ?? item.owned,
        stock,
        available: (stock ?? 1) > 0,
      };
    });
    const byType = patched.filter(
      (item) => filterType === 'all' || item.category === filterType || item.type === filterType
    );
    const filtered = filterCatalogItems(byType, catalogFilters);
    return sortCatalogItems(filtered, sort);
  }, [catalog, catalogPatches, filterType, sort, catalogFilters]);

  const resetFilters = () => {
    const next = new URLSearchParams(searchParams);
    for (const key of FILTER_URL_KEYS) next.delete(key);
    next.delete('page');
    setSearchParams(next);
  };

  const { page: safePage, pageCount, slice } = useMemo(
    () => paginateItems(filteredSorted, page, CATALOG_PAGE_SIZE),
    [filteredSorted, page]
  );

  useEffect(() => {
    if (safePage !== page) {
      setQuery({ page: String(safePage) }, true);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [safePage, page]);

  useEffect(() => {
    const qs = searchParams.toString();
    saveCatalogNav({
      ids: filteredSorted.map((it) => it.id),
      listPath: qs ? `/shop?${qs}` : '/shop',
    });
  }, [filteredSorted, searchParams]);

  const handleBuyClick = async (item: ShopItem) => {
    if (isMerchItem(item)) {
      try {
        const { allowed, reason } = await paymentApi.canPurchaseMerch();
        setMerchAllowed(allowed);
        if (!allowed) {
          setMerchBlockReason(reason || 'Покупка мерча недоступна без активной подписки');
          setNotification({
            type: 'error',
            message: `${reason || 'Покупка мерча недоступна'}. Оформите подписку.`,
            link: { label: 'Перейти к подписке', path: '/subscription' },
          });
          return;
        }
      } catch {
        setMerchAllowed(false);
        setMerchBlockReason('Не удалось проверить доступ к мерчу');
        setNotification({
          type: 'error',
          message: 'Не удалось проверить доступ к покупке мерча',
        });
        return;
      }
    } else {
      setMerchAllowed(null);
      setMerchBlockReason('');
    }
    setSelectedItem(item);
    setShowModal(true);
  };

  const handleConfirmPurchase = async () => {
    if (!selectedItem) return;
    const boughtId = selectedItem.id;

    try {
      const result: any = await useShopStore.getState().buyItem(boughtId);
      const remaining =
        typeof result?.remaining_stock === 'number'
          ? result.remaining_stock
          : typeof result?.remainingStock === 'number'
            ? result.remainingStock
            : null;

      setCatalog((prev) =>
        prev.map((it) => {
          if (it.id !== boughtId) return it;
          const nextStock =
            remaining !== null
              ? remaining
              : typeof it.stock === 'number'
                ? Math.max(0, it.stock - 1)
                : it.stock;
          return { ...it, owned: true, stock: nextStock, available: (nextStock ?? 1) > 0 };
        })
      );

      setNotification({
        type: 'success',
        message: `${selectedItem.name} успешно куплен!`,
      });
      setShowModal(false);
      setSelectedItem(null);
    } catch (error: any) {
      setNotification({
        type: 'error',
        message: error.message || 'Ошибка при покупке',
      });
    }
  };

  const handleConfirmCancel = async () => {
    if (!cancelTarget) return;
    const itemId = cancelTarget.item_id;
    const name = cancelTarget.item.name;
    try {
      const result: any = await cancelPurchase(itemId);
      const remaining =
        typeof result?.remaining_stock === 'number'
          ? result.remaining_stock
          : typeof result?.remainingStock === 'number'
            ? result.remainingStock
            : null;
      const refunded =
        result?.refunded_amount ??
        result?.refundedAmount ??
        cancelTarget.purchase_price ??
        cancelTarget.item.price_tickets;

      setCatalog((prev) =>
        prev.map((it) => {
          if (it.id !== itemId) return it;
          const nextStock =
            remaining !== null
              ? remaining
              : typeof it.stock === 'number'
                ? it.stock + 1
                : it.stock;
          return { ...it, owned: false, stock: nextStock, available: true };
        })
      );

      setNotification({
        type: 'success',
        message: `${name}: возвращено ${formatTicketAmount(refunded)}.`,
      });
      setCancelTarget(null);
    } catch (error: any) {
      setNotification({
        type: 'error',
        message: error.message || 'Ошибка при отмене покупки',
      });
    }
  };

  const token = localStorage.getItem('accessToken');

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-void text-text-secondary">
        <IconLabel name="lock" iconClassName="h-4 w-4">
          Войдите в аккаунт, чтобы просматривать магазин
        </IconLabel>
      </div>
    );
  }

  if ((loading || catalogLoading) && catalog.length === 0) {
    return (
      <div className="min-h-screen bg-void">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <PageShell width="wide">
      <PageHeader
        title={
          <IconLabel name="gift" iconClassName="h-7 w-7 text-horizon-gold">
            Магазин
          </IconLabel>
        }
        subtitle="Тратьте билетики на крутые предметы!"
        onBack={() => navigate('/')}
        backLabel="На главную"
        actions={
          <span className="flex items-center gap-1.5 rounded-sm border border-horizon-gold/30 bg-horizon-gold/10 px-3 py-1.5 font-hud text-sm tabular-nums text-horizon-gold">
            <Icon name="ticket" className="h-4 w-4" />
            {balance}
          </span>
        }
      />

      {error && <Notification type="error" message={error} onClose={clearError} />}

      {notification && (
        <>
          <Notification
            type={notification.type}
            message={notification.message}
            onClose={() => setNotification(null)}
          />
          {notification.link && (
            <div className="mb-4 text-center">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setNotification(null);
                  navigate(notification.link!.path);
                }}
              >
                {notification.link.label}
              </Button>
            </div>
          )}
        </>
      )}

      <div className="mb-6 flex flex-wrap gap-2">
        <FilterChip active={activeTab === 'shop'} onClick={() => setActiveTab('shop')}>
          <IconLabel name="cart">Товары</IconLabel>
        </FilterChip>
        <FilterChip
          active={activeTab === 'inventory'}
          onClick={() => {
            setActiveTab('inventory');
            fetchInventory();
          }}
        >
          <IconLabel name="backpack">Мой инвентарь ({inventory.length})</IconLabel>
        </FilterChip>
      </div>

      {activeTab === 'shop' ? (
        <>
          <div className="mb-3 flex flex-wrap items-center gap-2">
            {itemTypes.map((type) => (
              <FilterChip
                key={type.value}
                active={filterType === type.value}
                onClick={() => setQuery({ type: type.value, page: '1' })}
              >
                {type.icon ? (
                  <IconLabel name={type.icon}>{type.label}</IconLabel>
                ) : (
                  type.label
                )}
              </FilterChip>
            ))}
            <label className="ml-auto flex items-center gap-2 text-sm text-text-secondary">
              <span className="text-text-muted">Сортировка</span>
              <select
                className="rounded-sm border border-white/10 bg-nebula px-2 py-1.5 text-sm text-text-primary"
                value={sort}
                onChange={(e) =>
                  setQuery({ sort: e.target.value as CatalogSort, page: '1' })
                }
              >
                {CATALOG_SORT_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <CatalogFiltersPanel
            filters={catalogFilters}
            facets={facets}
            onChange={(patch) => setQuery(patch)}
            onReset={resetFilters}
            queryDraft={qDraft}
            onQueryDraftChange={setQDraft}
          />

          <p className="mb-4 text-sm text-text-muted">
            Показано {slice.length} · отфильтровано {filteredSorted.length} из {catalog.length}
          </p>

          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {slice.length === 0 ? (
              <p className="col-span-full py-16 text-center text-text-secondary">
                Нет товаров по текущим фильтрам
              </p>
            ) : (
              slice.map((item) => (
                <ShopItemCard
                  key={item.id}
                  item={item}
                  balance={balance}
                  onBuyClick={handleBuyClick}
                />
              ))
            )}
          </div>

          <CatalogPager
            page={safePage}
            pageCount={pageCount}
            total={filteredSorted.length}
            disabled={catalogLoading}
            onPageChange={(p) => setQuery({ page: String(p) })}
          />
        </>
      ) : (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {inventory.length === 0 ? (
            <p className="col-span-full flex items-center justify-center gap-2 py-16 text-text-secondary">
              <Icon name="backpack" className="h-5 w-5" />
              У вас пока нет купленных предметов
            </p>
          ) : (
            inventory.map((purchased) => {
              // Shop /inventory often has empty image_url; catalog (inventory svc) has images[].
              const fromCatalog = catalogById.get(purchased.item_id);
              const img =
                purchased.item.image_url ||
                purchased.item.images?.[0] ||
                purchased.item.icon_url ||
                fromCatalog?.image_url ||
                fromCatalog?.images?.[0] ||
                fromCatalog?.icon_url ||
                '';
              const title = purchased.item.name || fromCatalog?.name || 'Товар';
              const fallbackIcon = itemFallbackIcon({
                category: purchased.item.category || fromCatalog?.category,
                type: purchased.item.type || fromCatalog?.type,
                game_id: purchased.item.game_id || fromCatalog?.game_id,
              });
              return (
                <Card key={purchased.id} className="flex h-full flex-col">
                  <button
                    type="button"
                    className="flex min-h-0 flex-1 flex-col text-left text-inherit"
                    onClick={() => navigate(`/shop/item/${purchased.item_id}`)}
                  >
                    <CardImage
                      src={img}
                      alt={title}
                      className="w-full shrink-0"
                      fit="cover"
                      fallback={
                        <Icon name={fallbackIcon} className="h-12 w-12 text-horizon-gold" />
                      }
                    />
                    <div className="mt-4 min-w-0 flex-1">
                      <h4 className="font-display text-lg font-semibold leading-snug text-text-primary line-clamp-2">
                        {title}
                      </h4>
                      <p className="mt-1 text-xs text-text-muted">
                        Куплено: {new Date(purchased.purchased_at).toLocaleDateString()}
                      </p>
                    </div>
                  </button>
                  <div className="mt-3 flex justify-end border-t border-white/5 pt-3">
                    {purchased.can_cancel ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="shrink-0 gap-1.5 px-2"
                        disabled={cancelling}
                        title="Отменить покупку"
                        aria-label={`Отменить покупку ${title}`}
                        onClick={() => setCancelTarget(purchased)}
                      >
                        <Icon name="undo" className="h-3.5 w-3.5" />
                        <span className="hidden sm:inline">Отменить</span>
                      </Button>
                    ) : (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="shrink-0 gap-1.5 px-2 opacity-50"
                        disabled
                        title={
                          purchased.fulfilled_at
                            ? 'Товар уже отправлен — возврат недоступен.'
                            : 'Срок возврата истёк (7 дней с покупки).'
                        }
                        aria-label="Возврат недоступен"
                      >
                        <Icon name="undo" className="h-3.5 w-3.5" />
                        <span className="hidden sm:inline">Отменить</span>
                      </Button>
                    )}
                  </div>
                </Card>
              );
            })
          )}
        </div>
      )}

      <PurchaseModal
        isOpen={showModal}
        item={selectedItem}
        balance={balance}
        onConfirm={handleConfirmPurchase}
        onClose={() => {
          setShowModal(false);
          setSelectedItem(null);
        }}
        loading={useShopStore.getState().buying}
        merchAllowed={selectedItem && isMerchItem(selectedItem) ? merchAllowed : true}
        merchBlockReason={merchBlockReason}
        onGoSubscription={() => navigate('/subscription')}
      />

      <CancelPurchaseModal
        isOpen={!!cancelTarget}
        item={cancelTarget?.item ?? null}
        refundAmount={
          cancelTarget?.purchase_price || cancelTarget?.item.price_tickets || 0
        }
        loading={cancelling}
        onConfirm={() => void handleConfirmCancel()}
        onClose={() => setCancelTarget(null)}
      />
    </PageShell>
  );
};
