// frontend/src/components/Shop/Shop.tsx
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ShopItemCard from './ShopItemCard';
import PurchaseModal from './PurchaseModal';
import Notification from '../Common/Notification/Notification';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { paymentApi } from '../../services/paymentApi';
import './Shop.css';
import { useShopStore, type ShopItem } from '../../store/shopStore';

function isMerchItem(item: ShopItem): boolean {
  const cat = (item.category || '').toLowerCase();
  const type = (item.type || '').toLowerCase();
  return cat.includes('merch') || type.includes('merch');
}

export const Shop: React.FC = () => {
  const navigate = useNavigate();
  const {
    items,
    inventory,
    balance,
    loading,
    error,
    fetchItems,
    fetchInventory,
    fetchBalance,
    clearError,
  } = useShopStore();

  const [selectedItem, setSelectedItem] = useState<any>(null);
  const [showModal, setShowModal] = useState(false);
  const [activeTab, setActiveTab] = useState<'shop' | 'inventory'>('shop');
  const [filterType, setFilterType] = useState<string>('all');
  const [notification, setNotification] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
    link?: { label: string; path: string };
  } | null>(null);
  const [merchAllowed, setMerchAllowed] = useState<boolean | null>(null);
  const [merchBlockReason, setMerchBlockReason] = useState('');

  const itemTypes = [
    { value: 'all', label: 'Все' },
    { value: 'game_skin', label: '🎨 Скины' },
    { value: 'profile_theme', label: '🎨 Темы' },
    { value: 'merch', label: '🎁 Мерч' },
  ];

  // useEffect(() => {
  //   const token = localStorage.getItem('accessToken');
  //   if (token) {
  //     fetchItems();
  //     fetchBalance();
  //     if (activeTab === 'inventory') {
  //       fetchInventory();
  //     }
  //   }
  // }, [fetchItems, fetchBalance, fetchInventory, activeTab]);
  useEffect(() => {
  const token = localStorage.getItem('accessToken');
    if (token) {
      fetchItems();
      fetchBalance();
      fetchInventory();
    }
  }, [fetchItems, fetchBalance, fetchInventory]);

  const handleBuyClick = async (item: ShopItem) => {
    if (isMerchItem(item)) {
      try {
        const { allowed, reason } = await paymentApi.canPurchaseMerch();
        setMerchAllowed(allowed);
        if (!allowed) {
          setMerchBlockReason(reason || 'Покупка мерча недоступна без активной подписки');
          setNotification({
            type: 'error',
            message: `❌ ${reason || 'Покупка мерча недоступна'}. Оформите подписку.`,
            link: { label: 'Перейти к подписке', path: '/subscription' },
          });
          return;
        }
      } catch {
        setMerchAllowed(false);
        setMerchBlockReason('Не удалось проверить доступ к мерчу');
        setNotification({
          type: 'error',
          message: '❌ Не удалось проверить доступ к покупке мерча',
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
    
    try {
      await useShopStore.getState().buyItem(selectedItem.id);
      setNotification({
        type: 'success',
        message: `✅ ${selectedItem.name} успешно куплен!`,
      });
      setShowModal(false);
      setSelectedItem(null);
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

  if (loading && items.length === 0) {
    return (
      <div className="shop-container">
        <LoadingSpinner />
      </div>
    );
  }

  // Группируем товары по имени и объединяем статус owned
  const uniqueItems = items.reduce((acc, item) => {
    const existing = acc.find(i => i.name === item.name);
    if (existing) {
      // Если хотя бы один экземпляр owned - помечаем весь товар как owned
      existing.owned = existing.owned || item.owned;
      return acc;
    }
    acc.push({ ...item });
    return acc;
  }, [] as ShopItem[]);

  // Фильтруем по категории
  const filteredItems = uniqueItems.filter(item => 
    filterType === 'all' || item.category === filterType
  );

  return (
    <div className="shop-container">
      {/* Кнопка Назад */}
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
        <>
          <Notification
            type={notification.type}
            message={notification.message}
            onClose={() => setNotification(null)}
          />
          {notification.link && (
            <div style={{ textAlign: 'center', marginBottom: '1rem' }}>
              <button
                className="filter-btn active"
                onClick={() => {
                  setNotification(null);
                  navigate(notification.link!.path);
                }}
              >
                {notification.link.label}
              </button>
            </div>
          )}
        </>
      )}

      {/* Вкладки */}
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
          {/* Фильтры */}
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

          {/* Товары */}
          <div className="shop-grid">
            {filteredItems.length === 0 ? (
              <div className="shop-empty">
                <p>Нет товаров выбранного типа</p>
              </div>
            ) : (
              filteredItems.map((item) => (
                <ShopItemCard
                  key={item.id}
                  item={item}
                  balance={balance}
                  onBuyClick={handleBuyClick}
                />
              ))
            )}
          </div>
        </>
      ) : (
        // Инвентарь
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
        merchAllowed={selectedItem && isMerchItem(selectedItem) ? merchAllowed : true}
        merchBlockReason={merchBlockReason}
        onGoSubscription={() => navigate('/subscription')}
      />
    </div>
  );
};
