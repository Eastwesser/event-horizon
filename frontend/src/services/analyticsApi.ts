import api from './api';

export interface DayCount {
  day: string;
  count: number;
}

export interface RetentionPoint {
  day_n: number;
  rate: number;
}

/** Admin-only analytics endpoints (gateway RequireRole admin). */
export const analyticsApi = {
  dau: async (days = 30): Promise<{ days: DayCount[] }> => {
    const { data } = await api.get('/api/analytics/dau', { params: { days } });
    return data;
  },
  mau: async (days = 30): Promise<{ mau: number; window_days: number }> => {
    const { data } = await api.get('/api/analytics/mau', { params: { days } });
    return data;
  },
  retention: async (cohortDaysAgo = 7, windowDays = 7) => {
    const { data } = await api.get('/api/analytics/retention', {
      params: { cohort_days_ago: cohortDaysAgo, window_days: windowDays },
    });
    return data as {
      cohort_day: string;
      cohort_size: number;
      points: RetentionPoint[];
    };
  },
};
