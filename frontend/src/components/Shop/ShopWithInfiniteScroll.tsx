// frontend/src/components/Shop/ShopWithInfiniteScroll.tsx
import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import ShopItemCard from './ShopItemCard';
import PurchaseModal from './PurchaseModal';
import Notification from '../Common/Notification/Notification';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { useShopStore, type ShopItem } from '../../store/shopStore';
import { inventoryApi } from '../../services/inventoryApi';
import { inventoryToShopItem } from '../../lib/shopItemMap';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';
import { Icon, IconLabel, type IconName } from '../ui/Icon';

const ITEMS_PER_PAGE = 20;

export const ShopWithInfiniteScroll: React.FC = () => {
  const navigate = useNavigate();
  const {
    inventory,
    balance,
    error,
    fetchInventory,
    fetchBalance,
    clearError,
  } = useShopStore();

  const [items, setItems] = useState<ShopItem[]>([]);
  const [allItems, setAllItems] = useState<ShopItem[]>([]);
  const [displayedItems, setDisplayedItems] = useState<ShopItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [catalogLoading, setCatalogLoading] = useState(true);
  const [selectedItem, setSelectedItem] = useState<ShopItem | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [activeTab, setActiveTab] = useState<'shop' | 'inventory'>('shop');
  const [filterType, setFilterType] = useState<string>('all');
  const [notification, setNotification] = useState<{
    type: 'success' | 'error';
    message: string;
  } | null>(null);
  const [totalItems, setTotalItems] = useState(0);

  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadMoreRef = useRef<HTMLDivElement>(null);

  const itemTypes: { value: string; label: string; icon?: IconName }[] = [
    { value: 'all', label: 'Все' },
    { value: 'карточка', label: 'Карточки', icon: 'cards' },
    { value: 'game_skin', label: 'Скины', icon: 'palette' },
    { value: 'profile_theme', label: 'Темы', icon: 'palette' },
    { value: 'merch', label: 'Мерч', icon: 'gift' },
  ];

  // Catalog once on mount; owned flags patched when inventory arrives (no re-fetch).
  useEffect(() => {
    let cancelled = false;
    const loadItems = async () => {
      setCatalogLoading(true);
      try {
        const ownedIds = new Set(
          useShopStore.getState().inventory.map((p) => p.item_id)
        );
        const response = await inventoryApi.searchAllItems();
        if (cancelled) return;
        const shopItems: ShopItem[] = (response.items ?? []).map((item) =>
          inventoryToShopItem(item, ownedIds.has(item.id))
        );
        setAllItems(shopItems);
        setTotalItems(response.total);
        setItems(shopItems);
        setDisplayedItems(shopItems.slice(0, ITEMS_PER_PAGE));
        setHasMore(shopItems.length > ITEMS_PER_PAGE);
      } catch (error) {
        console.error('Failed to load items:', error);
      } finally {
        if (!cancelled) setCatalogLoading(false);
      }
    };
    loadItems();
    return () => {
      cancelled = true;
    };
  }, []);

  // Загрузка инвентаря и баланса
  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      fetchBalance();
      fetchInventory();
    }
  }, [fetchBalance, fetchInventory]);

  // Обновление отображаемых товаров при изменении фильтра
  useEffect(() => {
    setPage(1);
    const filtered = allItems.filter(item =>
      filterType === 'all' || item.category === filterType
    );
    setItems(filtered);
    setDisplayedItems(filtered.slice(0, ITEMS_PER_PAGE));
    setHasMore(filtered.length > ITEMS_PER_PAGE);
  }, [filterType, allItems]);

  // Обновляем owned статус из инвентаря
  useEffect(() => {
    const ownedIds = new Set(inventory.map(p => p.item_id));
    setAllItems(prev => {
      if (prev.length === 0) return prev;
      let changed = false;
      const next = prev.map(item => {
        const owned = ownedIds.has(item.id);
        if (item.owned === owned) return item;
        changed = true;
        return { ...item, owned };
      });
      return changed ? next : prev;
    });
  }, [inventory]);

  // Intersection Observer для бесконечной прокрутки
  useEffect(() => {
    if (observerRef.current) {
      observerRef.current.disconnect();
    }

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore) {
          loadMore();
        }
      },
      { threshold: 0.1 }
    );

    if (loadMoreRef.current) {
      observerRef.current.observe(loadMoreRef.current);
    }

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [hasMore, loadingMore, displayedItems]);

  const loadMore = useCallback(() => {
    if (loadingMore || !hasMore) return;

    setLoadingMore(true);
    const nextPage = page + 1;
    const start = (nextPage - 1) * ITEMS_PER_PAGE;
    const end = start + ITEMS_PER_PAGE;
    const newItems = items.slice(start, end);

    if (newItems.length === 0) {
      setHasMore(false);
      setLoadingMore(false);
      return;
    }

    setDisplayedItems(prev => [...prev, ...newItems]);
    setPage(nextPage);
    setHasMore(end < items.length);
    setLoadingMore(false);
  }, [items, page, loadingMore, hasMore]);

  const handleBuyClick = (item: ShopItem) => {
    setSelectedItem(item);
    setShowModal(true);
  };

  const handleConfirmPurchase = async () => {
    if (!selectedItem) return;

    try {
      await useShopStore.getState().buyItem(selectedItem.id);
      setNotification({
        type: 'success',
        message: `${selectedItem.name} успешно куплен!`,
      });
      setShowModal(false);
      setSelectedItem(null);
      // Обновляем owned статус
      await fetchInventory();
      await fetchBalance();
    } catch (error: any) {
      setNotification({
        type: 'error',
        message: error.message || 'Ошибка при покупке',
      });
    }
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedItem(null);
  };

  const handleBack = () => {
    navigate('/');
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

  if (catalogLoading && allItems.length === 0) {
    return <LoadingSpinner fullscreen />;
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
        onBack={handleBack}
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
        <Notification
          type={notification.type}
          message={notification.message}
          onClose={() => setNotification(null)}
        />
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
            <div className="mb-6 flex flex-wrap gap-2">
              {itemTypes.map((type) => (
                <FilterChip
                  key={type.value}
                  active={filterType === type.value}
                  onClick={() => setFilterType(type.value)}
                >
                  {type.icon ? (
                    <IconLabel name={type.icon}>{type.label}</IconLabel>
                  ) : (
                    type.label
                  )}
                </FilterChip>
              ))}
            </div>

            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {displayedItems.length === 0 ? (
                <p className="col-span-full py-16 text-center text-text-secondary">Нет товаров выбранного типа</p>
              ) : (
                displayedItems.map((item) => (
                  <ShopItemCard
                    key={item.id}
                    item={item}
                    balance={balance}
                    onBuyClick={handleBuyClick}
                  />
                ))
              )}
            </div>

            {/* Элемент для Intersection Observer — триггер подгрузки */}
            <div ref={loadMoreRef} className="my-5 h-5">
              {loadingMore && (
                <div className="flex flex-col items-center gap-2 py-5">
                  <Spinner size={32} />
                  <p className="text-sm text-text-secondary">Загрузка ещё...</p>
                </div>
              )}
              {!hasMore && displayedItems.length > 0 && (
                <p className="py-5 text-center text-text-secondary">
                  Все товары загружены ({totalItems} шт.)
                </p>
              )}
            </div>
          </>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {inventory.length === 0 ? (
              <p className="col-span-full flex items-center justify-center gap-2 py-16 text-text-secondary">
                <Icon name="backpack" className="h-5 w-5" />
                У вас пока нет купленных предметов
              </p>
            ) : (
              inventory.map((purchased) => (
                <div key={purchased.id} className="flex items-center gap-4 rounded-md border border-white/10 bg-nebula p-4">
                  <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-md bg-white/5">
                    {purchased.item.image_url ? (
                      <img
                        src={purchased.item.image_url}
                        alt={purchased.item.name}
                        className="h-full w-full rounded-md object-cover"
                        onError={(e) => {
                          e.currentTarget.style.display = 'none';
                          const fallback = e.currentTarget.nextElementSibling as HTMLElement | null;
                          if (fallback) fallback.classList.remove('hidden');
                        }}
                      />
                    ) : null}
                    <Icon
                      name="gift"
                      className={`h-8 w-8 text-text-muted${purchased.item.image_url ? ' hidden' : ''}`}
                    />
                  </div>
                  <div className="min-w-0">
                    <h4 className="truncate font-display text-sm font-semibold text-text-primary">{purchased.item.name}</h4>
                    <p className="truncate text-xs text-text-secondary">{purchased.item.description}</p>
                    <span className="text-xs text-text-muted">
                      Куплено: {new Date(purchased.purchased_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        <PurchaseModal
          isOpen={showModal}
          item={selectedItem}
          balance={balance}
          onConfirm={handleConfirmPurchase}
          onClose={handleCloseModal}
          loading={useShopStore.getState().buying}
        />
    </PageShell>
  );
};
