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
  type: 'game_skin' | 'profile_theme' | 'merch';
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
  
  fetchItems: () => Promise<void>;
  fetchInventory: () => Promise<void>;
  fetchBalance: () => Promise<void>;
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

      fetchItems: async () => {
        const { lastFetch, items } = get();
        const now = Date.now();
        
        // Проверяем кеш
        if (items.length > 0 && (now - lastFetch) < CACHE_DURATION) {
          console.log('📦 Используем кеш товаров');
          return;
        }

        set({ loading: true, error: null });
        try {
          const userId = localStorage.getItem('userId');
          if (!userId) throw new Error('Пользователь не авторизован');
          
          const response = await getShopItems();
          // Преобразуем данные из API в формат ShopItem
          const shopItems: ShopItem[] = response.data.items.map((item: any) => ({
            id: item.id,
            name: item.name,
            description: item.description || '',
            price_tickets: item.price,
            icon_url: item.image_url || '',
            type: item.category || 'merch',
            category: item.category || 'merch',
            game_id: item.game_id,
            image_url: item.image_url,
            available: item.available,
            owned: item.owned || false,
          }));
          
          set({ 
            items: shopItems, 
            loading: false,
            lastFetch: now 
          });
        } catch (error: any) {
          console.error('❌ Ошибка загрузки товаров:', error);
          set({ 
            error: error.response?.data?.message || 'Ошибка загрузки товаров',
            loading: false 
          });
          throw error;
        }
      },

      fetchInventory: async () => {
        set({ loading: true, error: null });
        try {
          const response = await getInventory();
          // Преобразуем данные из API в формат PurchasedItem
          const inventoryItems: PurchasedItem[] = response.data.items.map((item: any) => ({
            id: item.id,
            item_id: item.id,
            purchased_at: new Date().toISOString(),
            item: {
              id: item.id,
              name: item.name,
              description: item.description || '',
              price_tickets: item.price,
              icon_url: item.image_url || '',
              type: item.category || 'merch',
              category: item.category || 'merch',
              game_id: item.game_id,
              image_url: item.image_url,
              available: item.available,
              owned: true,
            }
          }));
          
          set({ 
            inventory: inventoryItems, 
            loading: false 
          });
        } catch (error: any) {
          console.error('❌ Ошибка загрузки инвентаря:', error);
          set({ 
            error: error.response?.data?.message || 'Ошибка загрузки инвентаря',
            loading: false 
          });
          throw error;
        }
      },

      fetchBalance: async () => {
        try {
          const userId = localStorage.getItem('userId');
          if (!userId) return;
          
          const response = await getAllBalances(userId);
          // Ищем баланс билетиков
          const ticketsBalance = response.data.balances?.find(
            (b: any) => b.currency === 'TICKETS' || b.currency === 'tickets'
          );
          set({ balance: ticketsBalance?.balance || 0 });
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
          
          // Обновляем баланс после покупки
          await get().fetchBalance();
          
          // Обновляем инвентарь
          await get().fetchInventory();
          
          // Обновляем список товаров (чтобы обновить статус owned)
          await get().fetchItems();
          
          set({ buying: false });
          return response.data;
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || 'Ошибка при покупке';
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
