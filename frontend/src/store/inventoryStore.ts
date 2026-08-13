import { create } from 'zustand';
import { inventoryApi, type InventoryItem, type SearchItemsRequest } from '../services/inventoryApi';

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
  clearSelected: () => void;
  clearError: () => void;
}

export const useInventoryStore = create<InventoryState>((set) => ({
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
        items: response.items,
        total: response.total,
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
    set({ loading: true, error: null });
    try {
      await inventoryApi.deleteItem(id);
      set((state) => ({
        items: state.items.filter((i) => i.id !== id),
        total: state.total - 1,
        selectedItem: state.selectedItem?.id === id ? null : state.selectedItem,
        loading: false,
      }));
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Не удалось удалить товар',
        loading: false,
      });
      throw error;
    }
  },

  clearSelected: () => set({ selectedItem: null }),
  clearError: () => set({ error: null }),
}));
