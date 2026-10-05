import api from './api';

export interface Author {
  id: string;
  user_id: string;
  display_name: string;
  bio: string;
  avatar_url: string;
  portfolio?: string;
  active: boolean;
  created_at_unix: number;
  updated_at_unix: number;
  verified_at_unix?: number;
}

export interface AuthorApplication {
  id: string;
  user_id: string;
  status: 'pending' | 'approved' | 'rejected' | string;
  display_name: string;
  portfolio: string;
  motivation: string;
  contact_email: string;
  created_at_unix: number;
  reviewed_at_unix?: number;
  reviewed_by?: string;
  reviewer_note?: string;
}

export interface AuthorSaleRow {
  id: string;
  item_id: string;
  item_name?: string;
  buyer_email: string;
  price: number;
  status: string;
  purchased_at: string;
  refunded_at?: string;
}

export interface AuthorSalesResponse {
  sales_count: number;
  tickets_earned: number;
  purchases: AuthorSaleRow[];
  total: number;
  author_id: string;
}

export const authorsApi = {
  getMe: async (): Promise<Author> => {
    const { data } = await api.get('/authors/me');
    return data;
  },
  upsertMe: async (body: {
    display_name: string;
    bio?: string;
    avatar_url?: string;
    portfolio?: string;
  }): Promise<Author> => {
    const { data } = await api.put('/authors/me', body);
    return data;
  },
  getSales: async (params?: {
    author_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<AuthorSalesResponse> => {
    const { data } = await api.get('/authors/me/sales', { params });
    return {
      sales_count: data?.sales_count ?? 0,
      tickets_earned: data?.tickets_earned ?? 0,
      purchases: Array.isArray(data?.purchases) ? data.purchases : [],
      total: data?.total ?? 0,
      author_id: data?.author_id ?? '',
    };
  },
  get: async (userId: string): Promise<Author> => {
    const { data } = await api.get(`/authors/${userId}`);
    return data;
  },
  list: async (limit = 20, offset = 0): Promise<{ authors: Author[]; total: number }> => {
    const { data } = await api.get('/authors', { params: { limit, offset } });
    return {
      authors: Array.isArray(data?.authors) ? data.authors : [],
      total: typeof data?.total === 'number' ? data.total : 0,
    };
  },
  apply: async (body: {
    display_name: string;
    portfolio?: string;
    motivation: string;
    contact_email?: string;
  }): Promise<AuthorApplication> => {
    const { data } = await api.post('/authors/apply', body);
    return data;
  },
  getMyApplication: async (): Promise<AuthorApplication> => {
    const { data } = await api.get('/authors/me/application');
    return data;
  },
};
