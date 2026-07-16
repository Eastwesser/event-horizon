🎯 Следующие сервисы (приоритет)

Payment — реальные платежи (Boosty/Stripe) — 🔥 ВАЖНО

Notification — уведомления (Telegram/Email) — 🔥 ВАЖНО

Analytics — DAU, MAU, Retention — 🟡 ПОТОМ

Почему Payment в первую очередь: Это даёт реальные деньги и мотивацию игрокам.

📋 План на завтра (16.07)

Нагрузочное тестирование (K6) — постреляем по API

Индексы в БД — оптимизация

Документирование — что уже есть

🌍 Текущая карта местности Event Horizon
✅ Что уже работает (v1.0.5 — багфикс):
Бэкенд:

Auth (JWT, регистрация, логин)

Game (валидация рекордов)

Billing (баланс, лампочки/билетики)

Leaderboard (топ, восстановление из БД)

Profile (агрегация данных)

Shop (магазин, покупки, инвентарь, purchased_at)

NATS кластер (3 ноды) + мониторинг

Gateway + Balancer

Фронтенд:

4 игры (Flappy, Hexagon, Towers, Memory)

Магазин (карточки, фильтры, инвентарь)

Профиль (баланс, рекорды, достижения)

Лидерборд (WebSocket работает)

Переключение скинов в играх

Инфраструктура:

Prometheus + Grafana + Jaeger

NATS кластер + Exporter

Docker Compose

🎯 Что дальше? (приоритеты)
Нагрузочное тестирование (K6) — сегодняшняя задача

Прогнать все сценарии: auth, game, leaderboard, shop, billing

Замерить RPS и latency

Найти узкие места

Оптимизация индексов в БД — следующая задача после K6

Индексы для таблиц: highscores, user_currencies, inventory, purchases

Анализ медленных запросов

Новые сервисы (после оптимизации):

Payment (реальные платежи)

Notification (уведомления)

Analytics (аналитика)

Инфраструктура (в фоне):

k3s (Kubernetes)

CI/CD

Ansible

🚀 Давай начнём с K6!
1. Проверь, что K6 установлен
bash
k6 --version
Если нет — установи:

bash
# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
2. Напиши тестовый скрипт для K6
javascript
// deployments/k6/load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Разгон до 10 пользователей
    { duration: '1m', target: 50 },   // До 50
    { duration: '1m', target: 100 },  // До 100
    { duration: '30s', target: 0 },   // Спад
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% запросов < 500ms
    http_req_failed: ['rate<0.01'],   // <1% ошибок
  },
};

// Токен для авторизации
let token = '';

export function setup() {
  // Логин для получения токена
  const loginRes = http.post('http://localhost:8079/api/auth/login', JSON.stringify({
    email: 'tuzer@example.com',
    password: 'tuzer1'
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(loginRes, { 'login successful': (r) => r.status === 200 });
  token = loginRes.json('access_token');
  return { token };
}

export default function (data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };
  
  // 1. Получить баланс
  const balanceRes = http.get('http://localhost:8079/api/billing/balance/all', { headers });
  check(balanceRes, { 'balance status 200': (r) => r.status === 200 });
  
  // 2. Получить лидерборд
  const leaderboardRes = http.get('http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10', { headers });
  check(leaderboardRes, { 'leaderboard status 200': (r) => r.status === 200 });
  
  // 3. Получить товары
  const shopRes = http.get('http://localhost:8079/api/shop/items', { headers });
  check(shopRes, { 'shop items status 200': (r) => r.status === 200 });
  
  // 4. Получить инвентарь
  const inventoryRes = http.get('http://localhost:8079/api/shop/inventory', { headers });
  check(inventoryRes, { 'inventory status 200': (r) => r.status === 200 });
  
  // 5. Отправить рекорд
  const scoreRes = http.post('http://localhost:8079/api/game/submit', JSON.stringify({
    user_id: '7fc8a659-1bb2-4d7c-b60e-c140239d5c62',
    game_id: 'hexagon',
    level: 1,
    score: Math.floor(Math.random() * 100),
    user_email: 'tuzer@example.com',
    seed: `test_${Date.now()}`,
    moves: [],
  }), { headers });
  check(scoreRes, { 'score submit status 200': (r) => r.status === 200 });
  
  sleep(1);
}
3. Запусти тест
bash
cd /home/denismatveev/event_horizon/deployments/k6
k6 run load-test.js
4. Мониторинг во время теста
Открой Grafana: http://localhost:3000

Смотри метрики: RPS, latency, ошибки

Нагрузка на БД и NATS

📊 Что будем смотреть после теста:
Узкие места — какие запросы самые медленные

Индексы — что можно добавить в БД

Пороги — превышен ли порог 500ms