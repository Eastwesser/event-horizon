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

const BASE_URL = '/inventory/items';

/** Backend validates limit ∈ [0, 100]. */
export const INVENTORY_PAGE_SIZE = 100;

export const inventoryApi = {
  // Создать товар
  createItem: async (data: CreateItemRequest): Promise<InventoryItem> => {
    const response = await api.post(BASE_URL, data);
    return response.data.item;
  },

  // Получить список товаров с фильтрами
  searchItems: async (params: SearchItemsRequest): Promise<SearchItemsResponse> => {
    const response = await api.get(BASE_URL, { params });
    return {
      items: response.data?.items ?? [],
      total: response.data?.total ?? 0,
    };
  },

  /**
   * Fetch all matching items by paging with limit≤100.
   * Caps page size at INVENTORY_PAGE_SIZE even if caller asks for more.
   */
  searchAllItems: async (
    params: Omit<SearchItemsRequest, 'limit' | 'offset'> = {}
  ): Promise<SearchItemsResponse> => {
    const pageSize = INVENTORY_PAGE_SIZE;
    const first = await inventoryApi.searchItems({
      ...params,
      limit: pageSize,
      offset: 0,
    });
    const total = first.total ?? 0;
    const items = [...(first.items ?? [])];
    let offset = items.length;
    while (offset < total) {
      const page = await inventoryApi.searchItems({
        ...params,
        limit: pageSize,
        offset,
      });
      const batch = page.items ?? [];
      if (batch.length === 0) break;
      items.push(...batch);
      offset += batch.length;
    }
    return { items, total };
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
    const response = await inventoryApi.searchAllItems({ author_id: authorId });
    return response.items;
  },

  // Получить товары по типу
  getByType: async (type: string): Promise<InventoryItem[]> => {
    const response = await inventoryApi.searchAllItems({ type });
    return response.items;
  },
};

/** Upload image for inventory item; returns public URL path (e.g. /uploads/….png). */
export async function uploadInventoryImage(file: File): Promise<string> {
  const form = new FormData();
  form.append('file', file);
  const response = await api.post<{ url: string }>('/uploads', form, {
    timeout: 60000,
  });
  const url = response.data?.url;
  if (!url) {
    throw new Error('Сервер не вернул URL');
  }
  return url;
}
