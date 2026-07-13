// frontend/src/services/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
});

// Добавляем JWT токен к каждому запросу
api.interceptors.request.use((config) => {
  console.log('📡 API Request:', config.method, config.url, config.data);
  console.log('📡 Headers:', config.headers);
  const token = localStorage.getItem('accessToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Обработка ошибок
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('❌ API Error:', {
      url: error.config?.url,
      method: error.config?.method,
      status: error.response?.status,
      message: error.response?.data?.message || error.message,
    });

    if (error.response?.status === 401) {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('userId');
      window.dispatchEvent(new Event('authChange'));
    }

    return Promise.reject(error);
  }
);

// ========== AUTH ==========
export const register = (email: string, password: string) =>
  api.post('/auth/register', { email, password });

export const login = (email: string, password: string) =>
  api.post('/auth/login', { email, password });

// ========== GAME ==========
export const submitScore = (userId: string, gameId: string, level: number, seed: string, moves: any[]) =>
  api.post('/game/submit', { user_id: userId, game_id: gameId, level, seed, moves });

// ========== BILLING ==========
export const getBalance = (userId: string, currency: 'lamps' | 'tickets') =>
  api.get('/billing/balance', { params: { user_id: userId, currency } });

export const getAllBalances = (userId: string) =>
  api.get('/billing/balance/all', { params: { user_id: userId } });

// ========== SHOP ==========
export const getShopItems = () =>
  api.get('/shop/items');

export const buyShopItem = (itemId: string) =>
  api.post('/shop/purchase', { item_id: itemId });

export const getInventory = () =>
  api.get('/shop/inventory');

// ========== PROFILE ==========
export const getProfile = () =>
  api.get('/profile');

export default api;