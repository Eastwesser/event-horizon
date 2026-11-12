import { create } from 'zustand';
import { inventoryApi, type InventoryItem, type SearchItemsRequest } from '../../services/inventoryApi';

interface InventoryState {
  items: InventoryItem[];
  total: number;
  loading: boolean;
  error: string | null;
  selectedItem: InventoryItem | null;

  fetchItems: (params?: SearchItemsRequest) => Promise<void>;
  fetchItem: (id: string) => Promise<void>;
  createItem: (data: any) => Promise<void>;
  updateItem: (id: string, data: any) => Promise<void>;
  deleteItem: (id: string) => Promise<void>;
  softDeleteItem: (id: string) => Promise<void>;
  clearSelected: () => void;
  clearError: () => void;
}

export const useInventoryStore = create<InventoryState>((set, get) => ({
  items: [],
  total: 0,
  loading: false,
  error: null,
  selectedItem: null,

  fetchItems: async (params = {}) => {
    set({ loading: true, error: null });
    try {
      const response = await inventoryApi.searchItems(params);
      set({
        items: response.items ?? [],
        total: response.total ?? 0,
        loading: false,
      });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Не удалось загрузить товары',
        loading: false,
      });
    }
  },

  fetchItem: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const item = await inventoryApi.getItem(id);
      set({ selectedItem: item, loading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Не удалось загрузить товар',
        loading: false,
      });
    }
  },

  createItem: async (data: any) => {
    set({ loading: true, error: null });
    try {
      const item = await inventoryApi.createItem(data);
      set((state) => ({
        items: [item, ...state.items],
        total: state.total + 1,
        loading: false,
      }));
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Не удалось создать товар',
        loading: false,
      });
      throw error;
    }
  },

  updateItem: async (id: string, data: any) => {
    set({ loading: true, error: null });
    try {
      const item = await inventoryApi.updateItem(id, data);
      set((state) => ({
        items: state.items.map((i) => (i.id === id ? item : i)),
        selectedItem: state.selectedItem?.id === id ? item : state.selectedItem,
        loading: false,
      }));
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Не удалось обновить товар',
        loading: false,
      });
      throw error;
    }
  },

  deleteItem: async (id: string) => {
    // 1. Сохраняем текущее состояние для отката
    const prevItems = get().items;
    const prevTotal = get().total;

    // 2. Оптимистично убираем карточку из стейта
    set({
      items: prevItems.filter((i) => i.id !== id),
      total: prevTotal - 1,
      error: null,
    });

    // 3. Отправляем запрос на бэк
    try {
      await inventoryApi.deleteItem(id);
    } catch (error: any) {
      // 4. Откат: если упало, возвращаем карточку и показываем ошибку
      set({
        items: prevItems,
        total: prevTotal,
        error: error.response?.data?.error || 'Не удалось удалить товар',
      });
      throw error;
    }
  },

  softDeleteItem: async (id: string) => {
    const prevItems = get().items;
    set({
      items: prevItems.map((i) => (i.id === id ? { ...i, deleted: true } : i)),
      error: null,
    });
    try {
      await inventoryApi.softDeleteItem(id);
    } catch (error: any) {
      set({ items: prevItems, error: error?.response?.data?.error || 'Ошибка' });
      throw error;
    }
  },

  clearSelected: () => set({ selectedItem: null }),
  clearError: () => set({ error: null }),
}));
