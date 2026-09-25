// frontend/src/components/Shop/ShopWithInfiniteScroll.tsx
import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import ShopItemCard from './ShopItemCard';
import PurchaseModal from './PurchaseModal';
import Notification from '../Common/Notification/Notification';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { useShopStore, type ShopItem } from '../../store/shopStore';
import { inventoryApi } from '../../services/inventoryApi';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Spinner } from '../ui/Spinner';
import { FilterChip } from '../ui/FilterChip';

const ITEMS_PER_PAGE = 20;

export const ShopWithInfiniteScroll: React.FC = () => {
  const navigate = useNavigate();
  const {
    inventory,
    balance,
    loading,
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
  const [selectedItem, setSelectedItem] = useState<any>(null);
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

  const itemTypes = [
    { value: 'all', label: 'Все' },
    { value: 'game_skin', label: '🎨 Скины' },
    { value: 'profile_theme', label: '🎨 Темы' },
    { value: 'merch', label: '🎁 Мерч' },
  ];

  // Загрузка всех товаров
  useEffect(() => {
    const loadItems = async () => {
      try {
        const response = await inventoryApi.searchItems({ limit: 1000 });
        const shopItems: ShopItem[] = response.items.map((item: any) => ({
          id: item.id,
          name: item.name,
          description: item.description || '',
          price_tickets: item.price || 0,
          icon_url: item.images?.[0] || '',
          type: item.type || 'other',
          category: item.type || 'other',
          game_id: undefined,
          image_url: item.images?.[0] || '',
          available: true,
          owned: false,
        }));
        setAllItems(shopItems);
        setTotalItems(response.total);
        setItems(shopItems);
        setDisplayedItems(shopItems.slice(0, ITEMS_PER_PAGE));
        setHasMore(shopItems.length > ITEMS_PER_PAGE);
      } catch (error) {
        console.error('Failed to load items:', error);
      }
    };
    loadItems();
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
    setAllItems(prev => prev.map(item => ({
      ...item,
      owned: ownedIds.has(item.id)
    })));
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
        message: `✅ ${selectedItem.name} успешно куплен!`,
      });
      setShowModal(false);
      setSelectedItem(null);
      // Обновляем owned статус
      await fetchInventory();
      await fetchBalance();
    } catch (error: any) {
      setNotification({
        type: 'error',
        message: error.message || '❌ Ошибка при покупке',
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
        🔒 Войдите в аккаунт, чтобы просматривать магазин
      </div>
    );
  }

  if (loading && allItems.length === 0) {
    return (
      <div className="min-h-screen bg-void">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <PageShell width="wide">
      <PageHeader
        title="🎁 Магазин"
        subtitle="Тратьте билетики на крутые предметы!"
        onBack={handleBack}
        backLabel="На главную"
        actions={
          <span className="flex items-center gap-1.5 rounded-sm border border-horizon-gold/30 bg-horizon-gold/10 px-3 py-1.5 font-hud text-sm tabular-nums text-horizon-gold">
            🎟️ {balance} билетиков
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
          🛒 Товары
        </FilterChip>
        <FilterChip
          active={activeTab === 'inventory'}
          onClick={() => {
            setActiveTab('inventory');
            fetchInventory();
          }}
        >
          🎒 Мой инвентарь ({inventory.length})
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
                  {type.label}
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
                  🎉 Все товары загружены ({totalItems} шт.)
                </p>
              )}
            </div>
          </>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {inventory.length === 0 ? (
              <p className="col-span-full py-16 text-center text-text-secondary">У вас пока нет купленных предметов 🎒</p>
            ) : (
              inventory.map((purchased) => (
                <div key={purchased.id} className="flex items-center gap-4 rounded-md border border-white/10 bg-nebula p-4">
                  <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-md bg-white/5 text-3xl">
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
                    <span className={purchased.item.image_url ? 'hidden' : undefined}>🎁</span>
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
