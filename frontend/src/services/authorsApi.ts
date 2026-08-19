import api from './api';

export interface Author {
  id: string;
  user_id: string;
  display_name: string;
  bio: string;
  avatar_url: string;
  active: boolean;
  created_at_unix: number;
  updated_at_unix: number;
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
    return data;
  },
};
