/**
 * highload/seed-users.js — регистрация пула пользователей для 10k-probe.js.
 *
 * Создаёт N пользователей с префиксом k6_u_<timestamp>_<i>@event-horizon.test
 * и сохраняет их email/password в load_test_results/k6_users.json.
 *
 * ВАЖНО: bcrypt 12 + rate limit login 5/s. 1000 регистраций = ~200s минимум.
 * Скрипт идёт последовательно (1 VU), с паузой между регистрациями,
 * чтобы не упираться в rate limiter и не плодить 429 в отчёте.
 *
 * Usage:
 *   BASE_URL=http://localhost:8079 \
 *   EH_SEED_COUNT=10 \
 *   k6 run deployments/k6/highload/seed-users.js
 *
 * Потом:
 *   EH_K6_USERS_FILE=load_test_results/k6_users.json \
 *   EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
 *   k6 run deployments/k6/highload/10k-probe.js
 *
 * Чистка после теста (SQL):
 *   DELETE FROM users WHERE email LIKE 'k6_u_%@event-horizon.test';
 *   -- и профили/балансы/авторы по этим user_id (если регал как author)
 */

import http from 'k6/http';
import { check } from 'k6';

const baseURL = __ENV.BASE_URL || 'http://localhost:8079';
const COUNT = Number(__ENV.EH_SEED_COUNT || 10);
const PAUSE_MS = Number(__ENV.EH_SEED_PAUSE_MS || 250); // 4 req/s, ниже 5/s лимитера
const PASSWORD = __ENV.EH_SEED_PASSWORD || 'k6_secret_2026';

export const options = {
  // Последовательно, 1 VU — регистрация это write path, не надо параллелить.
  scenarios: {
    seed: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: COUNT,
      maxDuration: '30m',
    },
  },
  thresholds: {
    // Регистрация должна проходить; 429 = мы слишком быстрые.
    http_req_failed: ['rate<0.10'],
  },
};

function sleepMs(ms) {
  // k6 sleep — секунды, но с дробной частью.
  return new Promise((r) => setTimeout(r, ms));
}

export default function () {
  // Уникальность: timestamp + iter, чтобы повторный запуск не падал на 409.
  const stamp = Date.now();
  const i = __ITER;
  const email = `k6_u_${stamp}_${i}@event-horizon.test`;
  const nickname = `k6_${stamp}_${i}`;

  const res = http.post(
    `${baseURL}/api/v1/auth/register`,
    JSON.stringify({ email, password: PASSWORD, nickname }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const body = (() => { try { return res.json(); } catch { return null; } })();
  const ok = check(res, {
    'register 200/201': (r) => r.status === 200 || r.status === 201,
    'register has user_id': () => body && (body.user_id || body.id),
  });

  if (!ok) {
    console.warn(
      `seed: register failed status=${res.status} body=${String(res.body).slice(0, 160)}`,
    );
  }

  // Пауза, чтобы не пробить rate limiter login/register (5/s).
  const until = Date.now() + PAUSE_MS;
  while (Date.now() < until) {
    // busy-wait — k6 не даёт sleep с миллисекундами в старом API.
    // Для 250ms это ок; для больших пауз лучше sleep(1).
  }
}

export function handleSummary(data) {
  // k6 не умеет писать файл из default-функции. Поэтому собираем
  // email/password здесь — но у нас нет доступа к списку из default.
  // Решение: default пишет в __ENV через console.log, а мы парсим
  // в шелле. См. ниже — простой вариант через stdout.
  //
  // Практичнее: после прогона собрать JSON из БД:
  //   docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon \
  //     -c "SELECT email FROM users WHERE email LIKE 'k6_u_%' ORDER BY created_at DESC LIMIT 1000;" \
  //     -t -A > /tmp/k6_emails.txt
  //   jq -Rn '[inputs | {email: ., password: "k6_secret_2026"}]' \
  //     < /tmp/k6_emails.txt > load_test_results/k6_users.json
  //
  // Ниже — просто человекочитаемый отчёт.
  const m = data.metrics;
  const reqs = m.http_reqs?.values?.count ?? 0;
  const failed = (m.http_req_failed?.values?.rate ?? 0) * 100;

  const text = `
==================================================
 seed-users — COUNT=${COUNT}  PAUSE_MS=${PAUSE_MS}
==================================================
  Requests : ${reqs}
  Failed   : ${failed.toFixed(2)} %
  Password : ${PASSWORD}
==================================================
  Дальше:
    1) Собери JSON из БД (см. комментарий в handleSummary):
       docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon \\
         -c "SELECT email FROM users WHERE email LIKE 'k6_u_%' ORDER BY created_at DESC LIMIT ${COUNT};" \\
         -t -A > /tmp/k6_emails.txt
       jq -Rn --arg pwd '${PASSWORD}' \\
         '[inputs | {email: ., password: \$pwd}]' \\
         < /tmp/k6_emails.txt > load_test_results/k6_users.json
    2) EH_K6_USERS_FILE=load_test_results/k6_users.json \\
       EH_PROFILE=50 ... k6 run deployments/k6/highload/10k-probe.js
==================================================
`;
  return { stdout: text };
}