// frontend/src/services/api.ts
import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import {
  clearAuth,
  getAccessToken,
  getRefreshToken,
  hydrateAuth,
  setTokens,
} from '../lib/auth';

export const API_BASE = '/api';

hydrateAuth();

const api = axios.create({
  baseURL: API_BASE,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

const AUTH_ATTEMPT_PATHS = ['/auth/login', '/auth/register', '/auth/refresh'];

function isAuthAttempt(url?: string): boolean {
  if (!url) return false;
  return AUTH_ATTEMPT_PATHS.some((p) => url.includes(p));
}

api.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  console.log('📡 API Request:', config.method, config.url, config.data);
  console.log('📡 Headers:', config.headers);
  return config;
});

let refreshPromise: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  const refresh = getRefreshToken();
  if (!refresh) return null;
  try {
    // Raw axios — avoid interceptor recursion on /auth/refresh
    const { data } = await axios.post(`${API_BASE}/auth/refresh`, {
      refresh_token: refresh,
    });
    const access = data.access_token as string | undefined;
    const nextRefresh = (data.refresh_token as string | undefined) || refresh;
    if (!access) return null;
    setTokens(access, nextRefresh);
    return access;
  } catch {
    return null;
  }
}

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    console.error('❌ API Error:', {
      url: error.config?.url,
      method: error.config?.method,
      status: error.response?.status,
      message:
        (error.response?.data as { error?: string; message?: string })?.error ||
        (error.response?.data as { message?: string })?.message ||
        error.message,
    });

    const original = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    if (error.response?.status === 401 && original && !original._retry) {
      const url = original.url;
      if (isAuthAttempt(url)) {
        return Promise.reject(error);
      }

      const sentAuth = Boolean(original.headers?.Authorization);
      if (!sentAuth) {
        return Promise.reject(error);
      }

      // Access expired (TTL ~15m) — try refresh once before logging out.
      original._retry = true;
      if (!refreshPromise) {
        refreshPromise = refreshAccessToken().finally(() => {
          refreshPromise = null;
        });
      }
      const newToken = await refreshPromise;
      if (newToken) {
        original.headers.Authorization = `Bearer ${newToken}`;
        return api(original);
      }
      clearAuth();
    }

    return Promise.reject(error);
  },
);

export const register = (email: string, password: string) =>
  api.post('/auth/register', { email, password });

export const login = (email: string, password: string) =>
  api.post('/auth/login', { email, password });

export const submitScore = (userId: string, gameId: string, level: number, seed: string, moves: any[]) =>
  api.post('/game/submit', { user_id: userId, game_id: gameId, level, seed, moves });

export const getLeaderboard = (gameId: string, limit = 10) =>
  api.get('/leaderboard', { params: { game_id: gameId, limit } });

export const getBalance = (userId: string, currency: 'lamps' | 'tickets') =>
  api.get('/billing/balance', { params: { user_id: userId, currency } });

export const getAllBalances = (userId: string) =>
  api.get('/billing/balance/all', { params: { user_id: userId } });

export const getShopItems = async () => {
  const response = await api.get('/shop/items');
  // Gateway may historically return null for an empty list; always expose [].
  const data = response.data;
  if (data == null) {
    return { ...response, data: [] as unknown[] };
  }
  if (Array.isArray(data)) {
    return response;
  }
  if (Array.isArray(data.items)) {
    return { ...response, data: data.items };
  }
  return { ...response, data: [] as unknown[] };
};

export const buyShopItem = (itemId: string) =>
  api.post('/shop/purchase', { item_id: itemId });

export const getInventory = () => api.get('/shop/inventory');

export const getProfile = () => api.get('/profile');

export default api;
