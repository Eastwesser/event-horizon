// frontend/src/components/Shop/ShopWithInfiniteScroll.tsx
import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import ShopItemCard from './ShopItemCard';
import PurchaseModal from './PurchaseModal';
import Notification from '../Common/Notification/Notification';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import './Shop.css';
import { useShopStore, type ShopItem } from '../../store/shopStore';
import { inventoryApi } from '../../services/inventoryApi';

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

  // Фильтрация товаров
  const filteredItems = allItems.filter(item =>
    filterType === 'all' || item.category === filterType
  );

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
      <div className="shop-container">
        <div className="shop-empty">
          <p>🔒 Войдите в аккаунт, чтобы просматривать магазин</p>
        </div>
      </div>
    );
  }

  if (loading && allItems.length === 0) {
    return (
      <div className="shop-container">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="shop-container">
      <button onClick={handleBack} className="back-btn" title="На главную">
        ← Назад
      </button>

      <div className="shop-header">
        <div className="shop-title">
          <h1>🎁 Магазин</h1>
          <p className="shop-subtitle">Тратьте билетики на крутые предметы!</p>
        </div>
        <div className="shop-balance">
          <span className="balance-icon">🎟️</span>
          <span className="balance-amount">{balance}</span>
          <span className="balance-label">билетиков</span>
        </div>
      </div>

      {error && (
        <Notification
          type="error"
          message={error}
          onClose={clearError}
        />
      )}

      {notification && (
        <Notification
          type={notification.type}
          message={notification.message}
          onClose={() => setNotification(null)}
        />
      )}

      <div className="shop-tabs">
        <button
          className={`tab ${activeTab === 'shop' ? 'active' : ''}`}
          onClick={() => setActiveTab('shop')}
        >
          🛒 Товары
        </button>
        <button
          className={`tab ${activeTab === 'inventory' ? 'active' : ''}`}
          onClick={() => {
            setActiveTab('inventory');
            fetchInventory();
          }}
        >
          🎒 Мой инвентарь ({inventory.length})
        </button>
      </div>

      {activeTab === 'shop' ? (
        <>
          <div className="shop-filters">
            {itemTypes.map(type => (
              <button
                key={type.value}
                className={`filter-btn ${filterType === type.value ? 'active' : ''}`}
                onClick={() => setFilterType(type.value)}
              >
                {type.label}
              </button>
            ))}
          </div>

          <div className="shop-grid">
            {displayedItems.length === 0 ? (
              <div className="shop-empty">
                <p>Нет товаров выбранного типа</p>
              </div>
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
          <div ref={loadMoreRef} style={{ height: '20px', margin: '20px 0' }}>
            {loadingMore && (
              <div style={{ textAlign: 'center', padding: '20px' }}>
                <LoadingSpinner />
                <p style={{ color: '#8a94a8', marginTop: '8px' }}>Загрузка ещё...</p>
              </div>
            )}
            {!hasMore && displayedItems.length > 0 && (
              <div style={{ textAlign: 'center', padding: '20px', color: '#8a94a8' }}>
                🎉 Все товары загружены ({totalItems} шт.)
              </div>
            )}
          </div>
        </>
      ) : (
        <div className="inventory-grid">
          {inventory.length === 0 ? (
            <div className="shop-empty">
              <p>У вас пока нет купленных предметов 🎒</p>
            </div>
          ) : (
            inventory.map((purchased) => (
              <div key={purchased.id} className="inventory-item">
                <div className="item-icon">
                  {purchased.item.image_url ? (
                    <img src={purchased.item.image_url} alt={purchased.item.name} />
                  ) : (
                    <span className="item-emoji">🎁</span>
                  )}
                </div>
                <div className="item-info">
                  <h4>{purchased.item.name}</h4>
                  <p>{purchased.item.description}</p>
                  <span className="purchased-date">
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
    </div>
  );
};
