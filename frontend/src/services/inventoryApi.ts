import api from './api';

export interface InventoryItem {
  id: string;
  author_id: string;
  type: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  attributes: Record<string, any>;
  images: string[];
  created_at: string;
  updated_at: string;
}

export interface CreateItemRequest {
  type: string;
  name: string;
  description?: string;
  price: number;
  stock?: number;
  attributes?: Record<string, any>;
  images?: string[];
}

export interface UpdateItemRequest {
  type?: string;
  name?: string;
  description?: string;
  price?: number;
  stock?: number;
  attributes?: Record<string, any>;
  images?: string[];
}

export interface SearchItemsRequest {
  author_id?: string;
  type?: string;
  price_min?: number;
  price_max?: number;
  query?: string;
  limit?: number;
  offset?: number;
}

export interface SearchItemsResponse {
  items: InventoryItem[];
  total: number;
}

const BASE_URL = '/api/inventory/items';

export const inventoryApi = {
  // Создать товар
  createItem: async (data: CreateItemRequest): Promise<InventoryItem> => {
    const response = await api.post(BASE_URL, data);
    return response.data.item;
  },

  // Получить список товаров с фильтрами
  searchItems: async (params: SearchItemsRequest): Promise<SearchItemsResponse> => {
    const response = await api.get(BASE_URL, { params });
    return response.data;
  },

  // Получить товар по ID
  getItem: async (id: string): Promise<InventoryItem> => {
    const response = await api.get(`${BASE_URL}/${id}`);
    return response.data.item;
  },

  // Обновить товар
  updateItem: async (id: string, data: UpdateItemRequest): Promise<InventoryItem> => {
    const response = await api.put(`${BASE_URL}/${id}`, data);
    return response.data.item;
  },

  // Удалить товар
  deleteItem: async (id: string): Promise<void> => {
    await api.delete(`${BASE_URL}/${id}`);
  },

  // Получить товары автора
  getByAuthor: async (authorId: string): Promise<InventoryItem[]> => {
    const response = await api.get(BASE_URL, {
      params: { author_id: authorId, limit: 100 }
    });
    return response.data.items;
  },

  // Получить товары по типу
  getByType: async (type: string): Promise<InventoryItem[]> => {
    const response = await api.get(BASE_URL, {
      params: { type, limit: 100 }
    });
    return response.data.items;
  },
};