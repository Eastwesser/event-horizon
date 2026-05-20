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

// Auth
export const register = (email: string, password: string) =>
  api.post('/auth/register', { email, password });

export const login = (email: string, password: string) =>
  api.post('/auth/login', { email, password });

// Game
export const submitScore = (userId: string, gameId: string, level: number, seed: string, moves: any[]) =>
  api.post('/game/submit', { user_id: userId, game_id: gameId, level, seed, moves });

// Billing
export const getBalance = (userId: string, currency: 'lamps' | 'tickets') =>
  api.get('/billing/balance', { params: { user_id: userId, currency } });

export const getAllBalances = (userId: string) =>
  api.get('/billing/balance/all', { params: { user_id: userId } });

export default api;