/**
 * Wave 4 — purchase → cancel (idempotency) load.
 *
 * Picks the cheapest non-merch shop item, buys it, cancels twice
 * (second cancel must report already_refunded / succeed without error).
 *
 * Usage:
 *   BASE_URL=http://localhost:8079 \
 *   EH_K6_EMAIL=... EH_K6_PASSWORD=... \
 *   k6 run deployments/k6/purchase.js
 *
 * Account must have enough tickets for at least one purchase.
 * Prefer a seed admin / funded user. Optional: EH_K6_ITEM_ID to pin an item.
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const baseURL = __ENV.BASE_URL || 'http://localhost:8079';

export const options = {
  vus: Number(__ENV.K6_VUS || 5),
  duration: __ENV.K6_DURATION || '30s',
  thresholds: {
    http_req_duration: ['p(95)<1200'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.15'],
  },
};

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
}

function pickItem(items) {
  const pinned = __ENV.EH_K6_ITEM_ID;
  if (pinned) {
    return pinned;
  }
  const candidates = (items || [])
    .filter((it) => {
      if (!it || !it.id) return false;
      const type = String(it.type || it.item_type || '').toLowerCase();
      const cat = String(it.category || '').toLowerCase();
      if (type === 'merch' || cat === 'merch') return false;
      const price = Number(it.price);
      return Number.isFinite(price) && price > 0;
    })
    .sort((a, b) => Number(a.price) - Number(b.price));
  return candidates.length ? candidates[0].id : null;
}

export function setup() {
  const email = __ENV.EH_K6_EMAIL;
  const password = __ENV.EH_K6_PASSWORD;
  if (!email || !password) {
    throw new Error('EH_K6_EMAIL and EH_K6_PASSWORD are required');
  }
  const loginRes = http.post(
    `${baseURL}/api/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (loginRes.status !== 200) {
    throw new Error(`login failed: ${loginRes.status} ${loginRes.body}`);
  }
  const token = loginRes.json('access_token');
  if (!token) {
    throw new Error('login response missing access_token');
  }

  const shopRes = http.get(`${baseURL}/api/shop/items`, { headers: authHeaders(token) });
  if (shopRes.status !== 200) {
    throw new Error(`shop items failed: ${shopRes.status}`);
  }
  const itemID = pickItem(shopRes.json());
  if (!itemID) {
    throw new Error('no purchasable non-merch item found (set EH_K6_ITEM_ID)');
  }
  return { token, itemID };
}

export default function (data) {
  const headers = authHeaders(data.token);
  const itemID = data.itemID;

  const buyRes = http.post(
    `${baseURL}/api/shop/purchase`,
    JSON.stringify({ item_id: itemID }),
    { headers },
  );
  const buyOK = check(buyRes, {
    'purchase 200': (r) => r.status === 200,
    'purchase success': (r) => r.json('success') === true,
  });
  errorRate.add(!buyOK);
  if (!buyOK) {
    sleep(0.5);
    return;
  }

  const cancel1 = http.post(`${baseURL}/api/shop/purchase/${itemID}/cancel`, null, { headers });
  const cancel1OK = check(cancel1, {
    'cancel1 200': (r) => r.status === 200,
    'cancel1 success': (r) => r.json('success') === true,
  });
  errorRate.add(!cancel1OK);

  const cancel2 = http.post(`${baseURL}/api/shop/purchase/${itemID}/cancel`, null, { headers });
  const cancel2OK = check(cancel2, {
    'cancel2 200': (r) => r.status === 200,
    'cancel2 idempotent': (r) =>
      r.json('success') === true &&
      (r.json('already_refunded') === true || Number(r.json('refunded_amount') || 0) === 0),
  });
  errorRate.add(!cancel2OK);

  sleep(0.5);
}
