// frontend/src/store/shopStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { getShopItems, buyShopItem, getInventory, getAllBalances } from '../services/api';

export interface ShopItem {
  id: string;
  name: string;
  description: string;
  price_tickets: number;
  icon_url: string;
  type: string;
  category: string;
  game_id?: string;
  image_url?: string;
  available: boolean;
  owned: boolean;
}

export interface PurchasedItem {
  id: string;
  item_id: string;
  purchased_at: string;
  item: ShopItem;
}

interface ShopState {
  items: ShopItem[];
  inventory: PurchasedItem[];
  balance: number;
  loading: boolean;
  buying: boolean;
  error: string | null;
  lastFetch: number;
  
  fetchItems: (force?: boolean) => Promise<void>;
  fetchInventory: () => Promise<void>;
  fetchBalance: (force?: boolean) => Promise<void>;
  buyItem: (itemId: string) => Promise<void>;
  clearError: () => void;
}

const CACHE_DURATION = 5 * 60 * 1000; // 5 минут

export const useShopStore = create<ShopState>()(
  persist(
    (set, get) => ({
      items: [],
      inventory: [],
      balance: 0,
      loading: false,
      buying: false,
      error: null,
      lastFetch: 0,

      fetchItems: async (force = false) => {
        const { lastFetch, items } = get();
        const now = Date.now();
        
        // Проверяем кеш только если не force
        if (!force && items.length > 0 && (now - lastFetch) < CACHE_DURATION) {
          console.log('📦 Используем кеш товаров');
          return;
        }

        set({ loading: true, error: null });
        try {
          const userId = localStorage.getItem('userId');
          if (!userId) throw new Error('Пользователь не авторизован');
          
          const response = await getShopItems();
          console.log('📦 Ответ от API /shop/items:', response.data);
          
          // Обрабатываем разные форматы ответа
          let itemsData = response.data;
          if (response.data && response.data.items && Array.isArray(response.data.items)) {
            itemsData = response.data.items;
          } else if (Array.isArray(response.data)) {
            itemsData = response.data;
          } else if (response.data && response.data.data && Array.isArray(response.data.data)) {
            itemsData = response.data.data;
          } else {
            console.error('❌ Неизвестный формат ответа:', response.data);
            throw new Error('Неверный формат данных от сервера');
          }
          
          // Получаем инвентарь для проверки owned
          const inventoryResponse = await getInventory();
          let inventoryData = inventoryResponse.data?.items || inventoryResponse.data || [];
          if (!Array.isArray(inventoryData)) {
            inventoryData = [];
          }
          const ownedIds = new Set(inventoryData.map((item: any) => item.id || item.item_id));
          
          const shopItems: ShopItem[] = itemsData.map((item: any) => ({
            id: item.id || item.Id || '',
            name: item.name || item.Name || 'Без названия',
            description: item.description || item.Description || '',
            price_tickets: item.price || item.Price || 0,
            icon_url: item.image_url || item.ImageUrl || '',
            type: item.category || item.Category || 'other',
            category: item.category || item.Category || 'other',
            game_id: item.game_id || item.GameId || undefined,
            image_url: item.image_url || item.ImageUrl || '',
            available: item.available !== undefined ? item.available : (item.Available !== undefined ? item.Available : true),
            owned: ownedIds.has(item.id || item.Id || ''),
          }));
          
          console.log('✅ Загружено товаров:', shopItems.length, 'из них куплено:', shopItems.filter(i => i.owned).length);
          
          set({ 
            items: shopItems, 
            loading: false,
            lastFetch: now 
          });
        } catch (error: any) {
          console.error('❌ Ошибка загрузки товаров:', error);
          set({ 
            error: error.response?.data?.message || error.message || 'Ошибка загрузки товаров',
            loading: false 
          });
          throw error;
        }
      },

      fetchInventory: async () => {
        set({ loading: true, error: null });
        try {
          const response = await getInventory();
          console.log('📦 Ответ от API /shop/inventory:', response.data);
          
          let inventoryData = response.data;
          if (response.data && response.data.items && Array.isArray(response.data.items)) {
            inventoryData = response.data.items;
          } else if (Array.isArray(response.data)) {
            inventoryData = response.data;
          } else if (response.data && response.data.data && Array.isArray(response.data.data)) {
            inventoryData = response.data.data;
          } else {
            console.warn('⚠️ Инвентарь пуст или неверный формат:', response.data);
            inventoryData = [];
          }
          
          const inventoryItems: PurchasedItem[] = inventoryData.map((item: any) => ({
            id: item.id || item.Id || crypto.randomUUID(),
            item_id: item.id || item.Id || '',
            // purchased_at: item.purchased_at || item.PurchasedAt || new Date().toISOString(),
            purchased_at: item.purchased_at || item.created_at || item.PurchasedAt || new Date().toISOString(),
            item: {
              id: item.id || item.Id || '',
              name: item.name || item.Name || 'Без названия',
              description: item.description || item.Description || '',
              price_tickets: item.price || item.Price || 0,
              icon_url: item.image_url || item.ImageUrl || '',
              type: item.category || item.Category || 'other',
              category: item.category || item.Category || 'other',
              game_id: item.game_id || item.GameId || undefined,
              image_url: item.image_url || item.ImageUrl || '',
              available: item.available !== undefined ? item.available : (item.Available !== undefined ? item.Available : true),
              owned: true,
            }
          }));
          
          console.log('✅ Загружено предметов в инвентаре:', inventoryItems.length);
          
          set({ 
            inventory: inventoryItems, 
            loading: false 
          });
        } catch (error: any) {
          console.error('❌ Ошибка загрузки инвентаря:', error);
          set({ 
            error: error.response?.data?.message || error.message || 'Ошибка загрузки инвентаря',
            loading: false 
          });
          throw error;
        }
      },

      fetchBalance: async (force = false) => {
        try {
          const userId = localStorage.getItem('userId');
          if (!userId) return;
          
          // Если force=false и есть кеш - используем его
          if (!force) {
            const cached = localStorage.getItem('shop_balance_cache');
            if (cached) {
              try {
                const { balance, timestamp } = JSON.parse(cached);
                if (Date.now() - timestamp < 30000) { // 30 секунд кеш
                  set({ balance });
                  return;
                }
              } catch (e) {
                // ignore
              }
            }
          }
          
          const response = await getAllBalances(userId);
          console.log('📦 Ответ от API /billing/balance/all:', response.data);
          
          let tickets = 0;
          if (response.data && typeof response.data === 'object') {
            tickets = response.data.tickets || response.data.TICKETS || 0;
          }
          
          set({ balance: tickets });
          // Сохраняем в кеш
          localStorage.setItem('shop_balance_cache', JSON.stringify({
            balance: tickets,
            timestamp: Date.now()
          }));
          console.log('💰 Баланс загружен:', { tickets });
        } catch (error) {
          console.error('❌ Ошибка загрузки баланса:', error);
        }
      },

      buyItem: async (itemId: string) => {
        set({ buying: true, error: null });
        try {
          const userId = localStorage.getItem('userId');
          if (!userId) throw new Error('Пользователь не авторизован');
          
          const response = await buyShopItem(itemId);
          console.log('📦 Ответ от API /shop/purchase:', response.data);
          
          // Принудительно обновляем баланс (force=true)
          await get().fetchBalance(true);
          
          // Обновляем инвентарь
          await get().fetchInventory();
          
          // Обновляем список товаров (чтобы обновить статус owned)
          await get().fetchItems(true);
          
          set({ buying: false });
          return response.data;
        } catch (error: any) {
          const errorMessage = error.response?.data?.error || error.response?.data?.message || error.message || 'Ошибка при покупке';
          console.error('❌ Ошибка покупки:', error);
          set({ 
            error: errorMessage,
            buying: false 
          });
          throw new Error(errorMessage);
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'shop-storage',
      partialize: (state) => ({
        items: state.items,
        lastFetch: state.lastFetch,
        // Не сохраняем баланс, т.к. он должен быть актуальным
      }),
    }
  )
);
