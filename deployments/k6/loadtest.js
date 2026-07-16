import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Rate } from 'k6/metrics';

// Кастомная метрика для отслеживания ошибок
const errorRate = new Rate('errors');

// Настройки нагрузки (для 10K RPS)
export const options = {
  stages: [
    { duration: '30s', target: 50 },    // 50 пользователей
    { duration: '1m', target: 200 },    // 200
    { duration: '1m', target: 500 },    // 500
    { duration: '30s', target: 0 },     // спад
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],    // 95% запросов < 500ms
    http_req_failed: ['rate<0.01'],      // Ошибок < 1%
    errors: ['rate<0.1'],                // Кастомная метрика ошибок < 10%
  },
};

// Генерация тестовых пользователей
const testUsers = new SharedArray('users', function () {
  const users = [];
  for (let i = 0; i < 100; i++) {
    users.push({
      email: `testuser${i}@example.com`,
      password: 'password123',
      nickname: `Player${i}`,
    });
  }
  return users;
});

// Используем setup для логина и получения userId
export function setup() {
  const baseUrl = 'http://localhost:8079';
  const user = testUsers[0]; // Берем первого пользователя для setup
  
  const loginRes = http.post(`${baseUrl}/api/auth/login`, JSON.stringify({
    email: user.email,
    password: user.password,
  }), { headers: { 'Content-Type': 'application/json' } });
  
  const token = loginRes.json('access_token');
  const userId = loginRes.json('user_id');
  
  return { token, userId };
}

// Основная функция
export default function (data) {
  const baseUrl = 'http://localhost:8079';
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];

  // 1. Регистрация (10% пользователей)
  if (Math.random() < 0.1) {
    const registerRes = http.post(`${baseUrl}/api/auth/register`, JSON.stringify({
      email: user.email,
      password: user.password,
      nickname: user.nickname,
    }), { headers: { 'Content-Type': 'application/json' } });

    const success = check(registerRes, {
      'register status 200': (r) => r.status === 200,
    });
    errorRate.add(!success);
    sleep(1);
  }

  // 2. Логин (получение токена)
  const loginRes = http.post(`${baseUrl}/api/auth/login`, JSON.stringify({
    email: user.email,
    password: user.password,
  }), { headers: { 'Content-Type': 'application/json' } });

  const loginCheck = check(loginRes, {
    'login status 200': (r) => r.status === 200,
    'login has token': (r) => r.json('access_token') !== undefined,
  });
  errorRate.add(!loginCheck);

  const token = loginRes.json('access_token');
  if (!token) {
    sleep(1);
    return;
  }

  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  // 3. Получить баланс
  const balanceRes = http.get(`${baseUrl}/api/billing/balance/all`, { headers });
  const balanceCheck = check(balanceRes, {
    'balance status 200': (r) => r.status === 200,
    'balance has tickets': (r) => r.json('tickets') !== undefined,
  });
  errorRate.add(!balanceCheck);
  sleep(0.5);

  // 4. Получить лидерборд (для всех игр)
  const games = ['hexagon', 'flappy', 'memory', 'towers'];
  const game = games[Math.floor(Math.random() * games.length)];
  const leaderboardRes = http.get(`${baseUrl}/api/leaderboard?game_id=${game}&limit=10`, { headers });
  const lbCheck = check(leaderboardRes, {
    'leaderboard status 200': (r) => r.status === 200,
  });
  errorRate.add(!lbCheck);
  sleep(0.5);

  // 5. Получить профиль
  const profileRes = http.get(`${baseUrl}/api/profile`, { headers });
  const profileCheck = check(profileRes, {
    'profile status 200': (r) => r.status === 200,
  });
  errorRate.add(!profileCheck);
  sleep(0.5);

  // 6. Получить товары магазина
  const shopRes = http.get(`${baseUrl}/api/shop/items`, { headers });
  const shopCheck = check(shopRes, {
    'shop items status 200': (r) => r.status === 200,
    'shop items array': (r) => Array.isArray(r.json()),
  });
  errorRate.add(!shopCheck);
  sleep(0.5);

  // 7. Получить инвентарь
  const inventoryRes = http.get(`${baseUrl}/api/shop/inventory`, { headers });
  const invCheck = check(inventoryRes, {
    'inventory status 200': (r) => r.status === 200,
  });
  errorRate.add(!invCheck);
  sleep(0.5);

  // 8. Отправить рекорд (используем реальный userId из setup)
  const submitRes = http.post(`${baseUrl}/api/game/submit`, JSON.stringify({
    user_id: data.userId, // ← используем реальный UUID из setup
    game_id: game,
    level: 1,
    score: Math.floor(Math.random() * 1000),
    user_email: user.email,
    nickname: user.nickname,
    seed: `k6_${Date.now()}`,
    moves: [],
  }), { headers });

  const submitCheck = check(submitRes, {
    'submit score status 200': (r) => r.status === 200,
    'submit has success': (r) => r.json('success') === true,
  });
  errorRate.add(!submitCheck);
  sleep(1);

  // 9. WebSocket (опционально, если нужно проверить подключение)
  // const ws = new WebSocket(`ws://localhost:8079/ws/leaderboard`);
  // ws.onopen = () => { ws.close(); };
}

export function handleSummary(data) {
  // Вывод результатов в консоль
  console.log(`
    ==================================================
    📊 РЕЗУЛЬТАТЫ НАГРУЗОЧНОГО ТЕСТА
    ==================================================
    ✅ Всего запросов:    ${data.metrics.http_reqs.values.count}
    ✅ Успешных:          ${data.metrics.http_reqs.values.rate}
    ❌ Ошибок:            ${data.metrics.http_req_failed.values.rate}%
    ⏱️  Средняя latency:  ${data.metrics.http_req_duration.values.avg.toFixed(2)} ms
    ⏱️  p95 latency:      ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)} ms
    ⏱️  Максимальная:     ${data.metrics.http_req_duration.values.max.toFixed(2)} ms
    ==================================================
  `);
}