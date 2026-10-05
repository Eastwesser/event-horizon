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

export const authorsApi = {
  upsertMe: async (body: { display_name: string; bio?: string; avatar_url?: string }): Promise<Author> => {
    const { data } = await api.put('/authors/me', body);
    return data;
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
