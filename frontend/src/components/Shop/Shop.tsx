// frontend/src/components/Shop/Shop.tsx
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ShopItemCard from './ShopItemCard';
import PurchaseModal from './PurchaseModal';
import Notification from '../Common/Notification/Notification';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { paymentApi } from '../../services/paymentApi';
import { useShopStore, type ShopItem } from '../../store/shopStore';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';

function isMerchItem(item: ShopItem): boolean {
  const cat = (item.category || '').toLowerCase();
  const type = (item.type || '').toLowerCase();
  return cat.includes('merch') || type.includes('merch');
}

const itemTypes = [
  { value: 'all', label: 'Все' },
  { value: 'game_skin', label: '🎨 Скины' },
  { value: 'profile_theme', label: '🎨 Темы' },
  { value: 'merch', label: '🎁 Мерч' },
];

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

  const handleBack = () => navigate('/');

  const token = localStorage.getItem('accessToken');

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-void text-text-secondary">
        🔒 Войдите в аккаунт, чтобы просматривать магазин
      </div>
    );
  }

  if (loading && items.length === 0) {
    return (
      <div className="min-h-screen bg-void">
        <LoadingSpinner />
      </div>
    );
  }

  // Группируем товары по имени и объединяем статус owned
  const uniqueItems = items.reduce((acc, item) => {
    const existing = acc.find(i => i.name === item.name);
    if (existing) {
      existing.owned = existing.owned || item.owned;
      return acc;
    }
    acc.push({ ...item });
    return acc;
  }, [] as ShopItem[]);

  const filteredItems = uniqueItems.filter(item =>
    filterType === 'all' || item.category === filterType
  );

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
              {filteredItems.length === 0 ? (
                <p className="col-span-full py-16 text-center text-text-secondary">Нет товаров выбранного типа</p>
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
          merchAllowed={selectedItem && isMerchItem(selectedItem) ? merchAllowed : true}
          merchBlockReason={merchBlockReason}
          onGoSubscription={() => navigate('/subscription')}
        />
    </PageShell>
  );
};
