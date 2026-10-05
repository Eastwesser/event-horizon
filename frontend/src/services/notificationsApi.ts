import api from './api';

export interface AppNotification {
  id: string;
  user_id: string;
  title: string;
  body: string;
  link: string;
  created_at_unix: number;
  read_at_unix: number;
}

export const notificationsApi = {
  list: async (
    limit = 30,
    offset = 0,
  ): Promise<{ notifications: AppNotification[]; total: number; unread_count: number }> => {
    const { data } = await api.get('/notifications', { params: { limit, offset } });
    return {
      notifications: data.notifications ?? [],
      total: data.total ?? 0,
      unread_count: data.unread_count ?? 0,
    };
  },

  markRead: async (id: string | 'all'): Promise<number> => {
    const { data } = await api.post(`/notifications/${id}/read`);
    return data.marked ?? 0;
  },
};
