import api from './api';

export type AdminRole = 'user' | 'author' | 'admin';

export interface AdminUserSubscription {
  active: boolean;
  plan: string;
  status: string;
}

export interface AdminUser {
  user_id: string;
  email: string;
  role: AdminRole;
  nickname: string;
  created_at: string;
  lamps: number;
  tickets: number;
  subscription: AdminUserSubscription;
}

export interface AdminUsersResponse {
  users: AdminUser[];
  total: number;
  limit: number;
  offset: number;
}

export interface TopExpensiveItem {
  id: string;
  name: string;
  price: number;
  author_id: string;
}

export interface InventoryStats {
  total_items: number;
  total_stock: number;
  by_type: Record<string, number>;
  by_author: Record<string, number>;
  author_emails: Record<string, string>;
  top_expensive: TopExpensiveItem[];
}

export const adminApi = {
  listUsers: async (params: {
    q?: string;
    limit?: number;
    offset?: number;
  }): Promise<AdminUsersResponse> => {
    const { data } = await api.get('/admin/users', { params });
    return {
      users: data.users ?? [],
      total: data.total ?? 0,
      limit: data.limit ?? 50,
      offset: data.offset ?? 0,
    };
  },

  updateRole: async (userId: string, role: AdminRole): Promise<void> => {
    await api.post('/auth/update-role', { user_id: userId, role });
  },

  getInventoryStats: async (): Promise<InventoryStats> => {
    const { data } = await api.get('/inventory/stats');
    return {
      total_items: data.total_items ?? 0,
      total_stock: data.total_stock ?? 0,
      by_type: data.by_type ?? {},
      by_author: data.by_author ?? {},
      author_emails: data.author_emails ?? {},
      top_expensive: Array.isArray(data.top_expensive) ? data.top_expensive : [],
    };
  },
};
