// Тип блина (начинка)
export type PancakeType = 
  | 'nutella'
  | 'strawberry'
  | 'fish'
  | 'sausage'
  | 'chicken'
  | 'caesar'
  | 'cranberry'
  | 'pancake';

// Стопка блинов
export interface Stack {
  type: PancakeType;
  count: number;
}

// Гекс с стопкой блинов
export interface HexTile {
  coord: { q: number; r: number };
  stack: Stack | null;
}

// Пользователь
export interface User {
  id: string;
  email: string;
}

// Баланс
export interface Balance {
  lamps: number;
  tickets: number;
}

// Запись в лидерборде
export interface LeaderboardEntry {
  rank: number;
  userId: string;
  userEmail: string;
  score: number;
}

// Ход для отправки на сервер
export interface GameMove {
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  timestamp: number;
}

// Ответ от сервера при отправке рекорда
export interface SubmitScoreResponse {
  success: boolean;
  newHighscore: number;
  rank: number;
  message: string;
  lampsEarned: number;
  ticketsEarned: number;
}