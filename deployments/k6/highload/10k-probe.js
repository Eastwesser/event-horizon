/**
 * highload/10k-probe.js — лестница нагрузки для Event Horizon.
 *
 * ЦЕЛЬ: не «10k VU сразу», а измеримая кривая. Ты сам выбираешь PROFILE.
 *
 * Профили (рекомендую именно в этом порядке):
 *   EH_PROFILE=10    — санити, всё живо
 *   EH_PROFILE=50    — реалистичный потолок на 8 GiB all-in-one (CORE+)
 *   EH_PROFILE=200   — за пределами CORE, смотрим деградацию
 *   EH_PROFILE=1000  — явно за границей, документируем «где упало»
 *   EH_PROFILE=10000 — только если готов ронять стек и смотреть в Grafana
 *
 * ВАЖНО: 10k DAU ≠ 10k VU. Peak ~30–50 RPS. 10k одновременных VU —
 * это стресс-тест «где сломается», а не «наш профиль».
 *
 * ВАЖНО: env-имена EH_*, а не K6_* — K6_* k6 читает как конфиг и
 * перезатирает scenarios. Если в шелле есть K6_VUS/K6_DURATION —
 * запускай через `env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS`.
 *
 * ПУЛ ЮЗЕРОВ (рекомендуется): EH_K6_USERS_JSON — JSON-строка массива
 * [{"email":"...","password":"..."}, ...]. Передаётся через env, без
 * файлов и без open() — надёжнее всего.
 *
 *   EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
 *   EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
 *   k6 run deployments/k6/highload/10k-probe.js
 *
 * Один admin-токен (fallback):
 *   EH_K6_EMAIL=admin@eventhorizon.local EH_K6_PASSWORD=... \
 *   EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
 *   k6 run deployments/k6/highload/10k-probe.js
 *
 * Сценарии (веса):
 *   browse   (70%) — shop/items, inventory/items, leaderboard, profile
 *   submit   (20%) — game/submit → outbox → NATS
 *   purchase (10%) — shop/purchase → cancel → cancel (идемпотентность)
 *
 * Thresholds мягкие — нам нужна кривая, а не pass/fail.
 * Жёсткие пороги — в CORE: browse.js / purchase.js.
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// ---- метрики ---------------------------------------------------------
const errorRate       = new Rate('errors');
const browseErrors    = new Rate('browse_errors');
const submitErrors    = new Rate('submit_errors');
const purchaseErrors  = new Rate('purchase_errors');
const purchaseOwned   = new Rate('purchase_owned'); // 409 item already owned
const status409       = new Rate('status_409');
const status429       = new Rate('status_429');
const status5xx       = new Rate('status_5xx');

// Trend по сценариям
const browseDur   = new Trend('browse_duration', true);
const submitDur   = new Trend('submit_duration', true);
const purchaseDur = new Trend('purchase_duration', true);

// Trend по ручкам browse (диагностика: кто даёт 2.5с)
const shopDur = new Trend('shop_dur', true);
const invDur  = new Trend('inv_dur', true);
const lbDur   = new Trend('lb_dur', true);
const profDur = new Trend('prof_dur', true);

// ---- конфиг ---------------------------------------------------------
const baseURL  = __ENV.BASE_URL || 'http://localhost:8079';
const PROFILE  = Number(__ENV.EH_PROFILE || 50);
const DURATION = __ENV.EH_DURATION || '60s';
const RAMP     = __ENV.EH_RAMP || '30s';

const USERS_JSON_ENV = __ENV.EH_K6_USERS_JSON || '';
const USERS_FILE_LEGACY = __ENV.EH_K6_USERS_FILE || '';

function sanitizeUsers(arr) {
  return arr
    .map((u) => ({
      email: String(u.email || '').replace(/[\r\n]/g, '').trim(),
      password: String(u.password || '').trim(),
    }))
    .filter((u) => u.email && u.password);
}

const usersPool = USERS_JSON_ENV
  ? new SharedArray('k6-users-env', () => {
      const arr = JSON.parse(USERS_JSON_ENV);
      if (!Array.isArray(arr) || arr.length === 0) {
        throw new Error('EH_K6_USERS_JSON must be a non-empty JSON array');
      }
      return sanitizeUsers(arr);
    })
  : USERS_FILE_LEGACY
  ? new SharedArray('k6-users-file', () => {
      const text = String(open(USERS_FILE_LEGACY, 'b'));
      if (!text || text[0] !== '[') {
        throw new Error(
          `users file ${USERS_FILE_LEGACY} did not read as JSON array; ` +
          `first 120 chars: ${text.slice(0, 120)}`,
        );
      }
      const arr = JSON.parse(text);
      if (!Array.isArray(arr) || arr.length === 0) {
        throw new Error(`${USERS_FILE_LEGACY} must be a non-empty JSON array`);
      }
      return sanitizeUsers(arr);
    })
  : null;

export const options = {
  scenarios: {
    browse: {
      executor: 'ramping-vus',
      exec: 'browseScenario',
      startVUs: 0,
      stages: [
        { duration: RAMP,     target: Math.floor(PROFILE * 0.7) },
        { duration: DURATION, target: Math.floor(PROFILE * 0.7) },
        { duration: RAMP,     target: 0 },
      ],
      gracefulRampDown: '10s',
    },
    submit: {
      executor: 'ramping-vus',
      exec: 'submitScenario',
      startVUs: 0,
      stages: [
        { duration: RAMP,     target: Math.floor(PROFILE * 0.2) },
        { duration: DURATION, target: Math.floor(PROFILE * 0.2) },
        { duration: RAMP,     target: 0 },
      ],
      gracefulRampDown: '10s',
    },
    purchase: {
      executor: 'ramping-vus',
      exec: 'purchaseScenario',
      startVUs: 0,
      stages: [
        { duration: RAMP,     target: Math.max(1, Math.floor(PROFILE * 0.1)) },
        { duration: DURATION, target: Math.max(1, Math.floor(PROFILE * 0.1)) },
        { duration: RAMP,     target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000', 'p(99)<5000'],
    http_req_failed:   ['rate<0.10'],
    errors:            ['rate<0.15'],
  },
};

export default function (data) {
  return browseScenario(data);
}

// ---- helpers --------------------------------------------------------
function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
}

function safeJSON(r) {
  try { return r.json(); } catch { return null; }
}

function bodyPreview(r) {
  const s = String(r.body || '');
  return s.length > 160 ? s.slice(0, 160) + '…' : s;
}

function login(email, password) {
  const res = http.post(
    `${baseURL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (res.status !== 200) {
    throw new Error(`login failed for ${email}: ${res.status} ${bodyPreview(res)}`);
  }
  const body = safeJSON(res);
  if (!body || !body.access_token) {
    throw new Error(`login response missing access_token for ${email}`);
  }
  return { token: body.access_token, userId: body.user_id || '' };
}

function pickToken(data) {
  if (data.pool && data.pool.length > 0) {
    return data.pool[(__VU - 1) % data.pool.length];
  }
  return { token: data.token, userId: data.userId };
}

// ---- setup ----------------------------------------------------------
export function setup() {
  let pool = null;
  let primary = { token: '', userId: '' };

  if (usersPool && usersPool.length > 0) {
    const limit = Number(__ENV.EH_K6_POOL_LIMIT || usersPool.length);
    const slice = usersPool.slice(0, limit);
    pool = [];
    for (let i = 0; i < slice.length; i++) {
      const u = slice[i];
      try {
        const r = login(u.email, u.password);
        pool.push(r);
      } catch (e) {
        console.warn(`pool login failed for ${u.email}: ${e.message}`);
      }
    }
    if (pool.length === 0) {
      throw new Error('users pool: all logins failed');
    }
    primary = pool[0];
    console.log(
      `10k-probe: PROFILE=${PROFILE} pool=${pool.length}/${slice.length} (from ${usersPool.length})`,
    );
  } else {
    const email = __ENV.EH_K6_EMAIL;
    const password = __ENV.EH_K6_PASSWORD;
    if (!email || !password) {
      throw new Error(
        'Set EH_K6_USERS_JSON, or both EH_K6_EMAIL and EH_K6_PASSWORD',
      );
    }
    primary = login(email, password);
    console.log(`10k-probe: PROFILE=${PROFILE} single-token mode (${email})`);
  }

  // Для purchase — список unowned non-merch items.
  const shopRes = http.get(`${baseURL}/api/v1/shop/items`, {
    headers: authHeaders(primary.token),
  });
  let itemIDs = [];
  const shopBody = safeJSON(shopRes);
  if (shopRes.status === 200 && Array.isArray(shopBody)) {
    itemIDs = shopBody
      .filter((it) => {
        if (!it || !it.id) return false;
        if (it.owned === true) return false;
        const type = String(it.type || it.item_type || '').toLowerCase();
        const cat = String(it.category || '').toLowerCase();
        if (type === 'merch' || cat === 'merch') return false;
        const price = Number(it.price);
        return Number.isFinite(price) && price > 0;
      })
      .sort((a, b) => Number(a.price) - Number(b.price))
      .map((it) => it.id);
  } else {
    console.warn(
      `setup: shop/items non-200: status=${shopRes.status} body=${bodyPreview(shopRes)}`,
    );
  }

  console.log(
    `10k-probe: PROFILE=${PROFILE} DURATION=${DURATION} RAMP=${RAMP} ` +
    `unowned_items=${itemIDs.length} pool=${pool ? pool.length : 'single'}`,
  );
  if (itemIDs.length === 0) {
    console.warn(
      '10k-probe: no unowned non-merch items — purchase scenario will idle',
    );
  }

  return { token: primary.token, userId: primary.userId, pool, itemIDs };
}

// ---- browse (70%): read path ----------------------------------------
export function browseScenario(data) {
  const { token } = pickToken(data);
  const headers = authHeaders(token);
  const game = ['hexagon', 'flappy', 'memory', 'towers'][
    Math.floor(Math.random() * 4)
  ];

  group('browse', () => {
    const t0 = Date.now();

    // shop/items
    const tShop = Date.now();
    const shop = http.get(`${baseURL}/api/v1/shop/items`, { headers });
    shopDur.add(Date.now() - tShop);
    const shopBody = safeJSON(shop);
    const shopOK = check(shop, {
      'shop/items 200': (r) => r.status === 200,
      'shop/items array': () => Array.isArray(shopBody),
    });
    if (!shopOK) {
      console.warn(`shop/items non-200: status=${shop.status} body=${bodyPreview(shop)}`);
    }
    errorRate.add(!shopOK);
    status409.add(shop.status === 409);
    status429.add(shop.status === 429);
    status5xx.add(shop.status >= 500);

    // inventory/items
    const tInv = Date.now();
    const inv = http.get(`${baseURL}/api/v1/inventory/items`, { headers });
    invDur.add(Date.now() - tInv);
    const invBody = safeJSON(inv);
    const invOK = check(inv, {
      'inventory/items 200': (r) => r.status === 200,
      'inventory/items has items': () => invBody && Array.isArray(invBody.items),
    });
    if (!invOK) {
      console.warn(`inventory/items non-200: status=${inv.status} body=${bodyPreview(inv)}`);
    }
    errorRate.add(!invOK);
    status409.add(inv.status === 409);
    status429.add(inv.status === 429);
    status5xx.add(inv.status >= 500);

    // leaderboard
    const tLb = Date.now();
    const lb = http.get(
      `${baseURL}/api/v1/leaderboard?game_id=${game}&limit=10`,
      { headers },
    );
    lbDur.add(Date.now() - tLb);
    const lbOK = check(lb, { 'leaderboard 200': (r) => r.status === 200 });
    if (!lbOK) {
      console.warn(`leaderboard non-200: status=${lb.status} body=${bodyPreview(lb)}`);
    }
    errorRate.add(!lbOK);
    status409.add(lb.status === 409);
    status429.add(lb.status === 429);
    status5xx.add(lb.status >= 500);

    // profile
    const tProf = Date.now();
    const prof = http.get(`${baseURL}/api/v1/profile`, { headers });
    profDur.add(Date.now() - tProf);
    const profOK = check(prof, { 'profile 200': (r) => r.status === 200 });
    if (!profOK) {
      console.warn(`profile non-200: status=${prof.status} body=${bodyPreview(prof)}`);
    }
    errorRate.add(!profOK);
    status409.add(prof.status === 409);
    status429.add(prof.status === 429);
    status5xx.add(prof.status >= 500);

    browseErrors.add(!(shopOK && invOK && lbOK && profOK));
    browseDur.add(Date.now() - t0);
  });

  sleep(0.3);
}

// ---- submit (20%): write path → outbox → NATS -----------------------
export function submitScenario(data) {
  const { token, userId } = pickToken(data);
  const headers = authHeaders(token);
  const game = ['hexagon', 'flappy', 'memory', 'towers'][
    Math.floor(Math.random() * 4)
  ];

  group('submit', () => {
    const t0 = Date.now();
    const res = http.post(
      `${baseURL}/api/v1/game/submit`,
      JSON.stringify({
        user_id: userId,
        game_id: game,
        level: 1,
        score: Math.floor(Math.random() * 1000),
        seed: `k6_${Date.now()}_${__VU}_${__ITER}`,
        moves: [],
      }),
      { headers },
    );
    const body = safeJSON(res);
    const ok = check(res, {
      'submit 200': (r) => r.status === 200,
      'submit success': () => body && body.success === true,
    });
    if (!ok) {
      console.warn(`submit non-200: status=${res.status} body=${bodyPreview(res)}`);
    }
    errorRate.add(!ok);
    submitErrors.add(!ok);
    status409.add(res.status === 409);
    status429.add(res.status === 429);
    status5xx.add(res.status >= 500);
    submitDur.add(Date.now() - t0);
  });

  sleep(1);
}

// ---- purchase (10%): buy → cancel → cancel (idempotency) ------------
export function purchaseScenario(data) {
  if (!data.itemIDs || data.itemIDs.length === 0) {
    sleep(1);
    return;
  }
  const { token } = pickToken(data);
  const headers = authHeaders(token);
  const itemID = data.itemIDs[(__VU - 1) % data.itemIDs.length];

  group('purchase', () => {
    const t0 = Date.now();

    const buy = http.post(
      `${baseURL}/api/v1/shop/purchase`,
      JSON.stringify({ item_id: itemID }),
      { headers },
    );
    const buyBody = safeJSON(buy);
    const buyOK = check(buy, {
      'purchase 200': (r) => r.status === 200,
      'purchase success': () => buyBody && buyBody.success === true,
    });

    // 409 item already owned — не ошибка системы, а «товар уже куплен».
    // Пропускаем итерацию, не считаем в purchaseErrors.
    const owned = buy.status === 409;
    if (owned) {
      purchaseOwned.add(true);
      errorRate.add(false);        // не считаем ошибкой
      status409.add(true);
      // На 409 не идём в cancel — покупки нет.
      purchaseDur.add(Date.now() - t0);
      sleep(0.5);
      return;
    }
    purchaseOwned.add(false);

    if (!buyOK) {
      console.warn(`purchase non-200: status=${buy.status} body=${bodyPreview(buy)}`);
    }
    errorRate.add(!buyOK);
    status409.add(false);
    status429.add(buy.status === 429);
    status5xx.add(buy.status >= 500);

    let purchaseID = itemID;
    if (buyBody) {
      purchaseID = buyBody.purchase_id || buyBody.id || itemID;
    }

    const c1 = http.post(
      `${baseURL}/api/v1/shop/purchase/${purchaseID}/cancel`,
      null,
      { headers },
    );
    const c1OK = check(c1, { 'cancel1 200': (r) => r.status === 200 });
    if (!c1OK) {
      console.warn(`cancel1 non-200: status=${c1.status} body=${bodyPreview(c1)}`);
    }
    errorRate.add(!c1OK);
    status409.add(c1.status === 409);
    status429.add(c1.status === 429);
    status5xx.add(c1.status >= 500);

    const c2 = http.post(
      `${baseURL}/api/v1/shop/purchase/${purchaseID}/cancel`,
      null,
      { headers },
    );
    const c2Body = safeJSON(c2);
    // Валидный ответ на повторный cancel:
    //   успех + (already_refunded=true ИЛИ refunded_amount>0 ИЛИ balance не изменился)
    // Бэк возвращает {success:true, already_refunded:false, refunded_amount:750}
    // на ПЕРВЫЙ cancel — это ОК, деньги вернулись. На ВТОРОЙ — already_refunded:true.
    const c2OK = check(c2, {
      'cancel2 200': (r) => r.status === 200,
      'cancel2 idempotent': () => {
        if (!c2Body || c2Body.success !== true) return false;
        // Первый cancel: already_refunded=false + refunded_amount>0
        if (c2Body.already_refunded === false &&
            Number(c2Body.refunded_amount || 0) > 0) return true;
        // Второй cancel: already_refunded=true (деньги не двигаются)
        if (c2Body.already_refunded === true) return true;
        // Fallback: refunded_amount=0 (нечего возвращать)
        if (Number(c2Body.refunded_amount || 0) === 0) return true;
        return false;
      },
    });
    if (!c2OK) {
      console.warn(`cancel2 non-200: status=${c2.status} body=${bodyPreview(c2)}`);
    }
    errorRate.add(!c2OK);
    status409.add(c2.status === 409);
    status429.add(c2.status === 429);
    status5xx.add(c2.status >= 500);

    // purchaseErrors — только если buy/c1/c2 реально сломались (не 409).
    purchaseErrors.add(!(buyOK && c1OK && c2OK));
    purchaseDur.add(Date.now() - t0);
  });

  sleep(0.5);
}

// ---- summary ---------------------------------------------------------
export function handleSummary(data) {
  const m = data.metrics;
  const g = (name, k) => (m[name]?.values?.[k] ?? 0);
  const f = (x) => (typeof x === 'number' ? x.toFixed(1) : '0.0');
  const pct = (x) => ((x || 0) * 100).toFixed(2);
  const failRate = (m.http_req_failed?.values?.rate ?? 0) * 100;
  const reqs = m.http_reqs?.values?.count ?? 0;

  // p50/p99 в k6 2.x иногда лежат как p(50.00)/p(99.00).
  // Берём с fallback: p50 → avg, p99 → max.
  const p50 = g('http_req_duration','p(50)') || g('http_req_duration','avg');
  const p99 = g('http_req_duration','p(99)') || g('http_req_duration','max');

  const text = `
==================================================
 10k-probe — PROFILE=${PROFILE}  DURATION=${DURATION}  RAMP=${RAMP}
==================================================
  Requests total : ${reqs}
  Failed         : ${failRate.toFixed(2)} %
  p50            : ${f(p50)} ms   (avg fallback)
  p95            : ${f(g('http_req_duration','p(95)'))} ms
  p99            : ${f(p99)} ms   (max fallback)
  max            : ${f(g('http_req_duration','max'))} ms

  browse   p95   : ${f(g('browse_duration','p(95)'))} ms   errors: ${pct(g('browse_errors','rate'))}%
  submit   p95   : ${f(g('submit_duration','p(95)'))} ms   errors: ${pct(g('submit_errors','rate'))}%
  purchase p95   : ${f(g('purchase_duration','p(95)'))} ms   errors: ${pct(g('purchase_errors','rate'))}%

  browse breakdown (p95):
    shop/items      : ${f(g('shop_dur','p(95)'))} ms
    inventory/items : ${f(g('inv_dur','p(95)'))} ms
    leaderboard     : ${f(g('lb_dur','p(95)'))} ms
    profile         : ${f(g('prof_dur','p(95)'))} ms

  status_409     : ${pct(g('status_409','rate'))} %
  status_429     : ${pct(g('status_429','rate'))} %
  status_5xx     : ${pct(g('status_5xx','rate'))} %
  purchase_owned : ${pct(g('purchase_owned','rate'))} %   (409 item already owned — не ошибка)
==================================================
`;
  const stamp = new Date().toISOString().replace(/[:.]/g, '-');
  const jsonPath = `load_test_results/10k-probe_P${PROFILE}_${stamp}.json`;
  return { stdout: text, [jsonPath]: JSON.stringify(data, null, 2) };
}