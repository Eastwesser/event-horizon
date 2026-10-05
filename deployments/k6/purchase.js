/**
 * Wave 4 — purchase → cancel (idempotency) load.
 *
 * Each VU gets its own unowned non-merch item (buy → cancel → cancel again).
 * Cancel restores ownership so the same VU can loop without 409.
 *
 * Usage:
 *   BASE_URL=http://localhost:8079 \
 *   EH_K6_EMAIL=... EH_K6_PASSWORD=... \
 *   K6_VUS=1 K6_DURATION=30s \
 *   k6 run deployments/k6/purchase.js
 *
 * Optional: EH_K6_ITEM_ID — pin as VU1's item only if that id is currently unowned.
 * If fewer unowned items than VUs, extra VUs idle (console warning).
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const baseURL = __ENV.BASE_URL || 'http://localhost:8079';
const requestedVUs = Number(__ENV.K6_VUS || 5);

export const options = {
  vus: requestedVUs,
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

/** Non-merch, unowned, price > 0 — cheapest first. */
function unownedCandidates(items) {
  return (items || [])
    .filter((it) => {
      if (!it || !it.id) return false;
      if (it.owned === true) return false;
      const type = String(it.type || it.item_type || '').toLowerCase();
      const cat = String(it.category || '').toLowerCase();
      if (type === 'merch' || cat === 'merch') return false;
      const price = Number(it.price);
      return Number.isFinite(price) && price > 0;
    })
    .sort((a, b) => Number(a.price) - Number(b.price));
}

function pickItemIDs(items, need) {
  const candidates = unownedCandidates(items);
  const ids = candidates.map((it) => it.id);
  const pinned = (__ENV.EH_K6_ITEM_ID || '').trim();
  if (pinned) {
    const idx = ids.indexOf(pinned);
    if (idx === -1) {
      console.warn(
        `EH_K6_ITEM_ID=${pinned} is missing or already owned — ignoring pin`,
      );
    } else {
      ids.splice(idx, 1);
      ids.unshift(pinned);
    }
  }
  return ids.slice(0, need);
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

  const itemIDs = pickItemIDs(shopRes.json(), requestedVUs);
  if (itemIDs.length === 0) {
    throw new Error(
      'no unowned non-merch items (cancel some purchases or seed more skins)',
    );
  }
  if (itemIDs.length < requestedVUs) {
    console.warn(
      `only ${itemIDs.length} unowned item(s) for ${requestedVUs} VUs — VUs > ${itemIDs.length} will idle`,
    );
  }
  console.log(`purchase.js: assigned ${itemIDs.length} item(s) to VUs: ${itemIDs.join(', ')}`);
  return { token, itemIDs };
}

export default function (data) {
  const headers = authHeaders(data.token);
  const idx = __VU - 1;
  if (idx >= data.itemIDs.length) {
    sleep(1);
    return;
  }
  const itemID = data.itemIDs[idx];

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
