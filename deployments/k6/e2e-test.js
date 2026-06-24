import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '10s',
};

export default function () {
  const baseURL = 'http://localhost:8079';
  
  // Генерируем уникальный email для каждой итерации
  const email = `test_${Date.now()}_${__ITER}@example.com`;
  const nickname = `Player_${Date.now()}_${__ITER}`;

  // 1. Регистрация
  const registerPayload = JSON.stringify({
    email: email,
    password: 'secret123',
    nickname: nickname,
  });

  const registerRes = http.post(`${baseURL}/api/auth/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(registerRes, {
    '✅ регистрация успешна': (r) => r.status === 200 || r.status === 201,
  });

  // Небольшая задержка, чтобы регистрация успела завершиться
  sleep(0.5);

  // 2. Логин
  const loginPayload = JSON.stringify({
    email: email,
    password: 'secret123',
  });

  const loginRes = http.post(`${baseURL}/api/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  let token = null;
  if (loginRes.status === 200) {
    const body = JSON.parse(loginRes.body);
    token = body.access_token;  // ← исправлено: .access_token
    check(loginRes, {
      '✅ логин успешен': (r) => r.status === 200,
    });
  }

  if (!token) {
    console.error('❌ Не удалось получить токен');
    return;
  }

  // 3. Отправка рекорда (игра)
  const scorePayload = JSON.stringify({
    game_id: 'hexagon',
    user_id: 'test-user',
    level: 1,
    score: 150,
    seed: 'test_seed',
    moves: [],
  });

  const scoreRes = http.post(`${baseURL}/api/game/submit`, scorePayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });

  check(scoreRes, {
    '✅ рекорд отправлен': (r) => r.status === 200,
  });

  // 4. Проверка лидерборда
  const leaderboardRes = http.get(`${baseURL}/api/leaderboard?game_id=hexagon&limit=10`);
  
  check(leaderboardRes, {
    '✅ лидерборд доступен': (r) => r.status === 200,
  });

  sleep(1);
}