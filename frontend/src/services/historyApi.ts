import api from './api';

export interface HistoryEvent {
  id: string;
  user_id: string;
  event_type: string;
  payload_json: string;
  created_at_unix: number;
}

export const historyApi = {
  list: async (
    limit = 50,
    offset = 0,
    eventType = '',
  ): Promise<{ events: HistoryEvent[]; total: number }> => {
    const { data } = await api.get('/history', {
      params: {
        limit,
        offset,
        ...(eventType ? { event_type: eventType } : {}),
      },
    });
    return {
      events: data.events ?? [],
      total: data.total ?? 0,
    };
  },
};
