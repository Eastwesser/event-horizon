/**
 * Wave 4 — browsing load: shop catalog + inventory catalog list DTOs.
 *
 * Usage:
 *   BASE_URL=http://localhost:8079 \
 *   EH_K6_EMAIL=... EH_K6_PASSWORD=... \
 *   k6 run deployments/k6/browse.js
 *
 * Optional: K6_VUS=20 K6_DURATION=30s
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const baseURL = __ENV.BASE_URL || 'http://localhost:8079';

export const options = {
  vus: Number(__ENV.K6_VUS || 20),
  duration: __ENV.K6_DURATION || '30s',
  thresholds: {
    http_req_duration: ['p(95)<800'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.1'],
  },
};

export function setup() {
  const email = __ENV.EH_K6_EMAIL;
  const password = __ENV.EH_K6_PASSWORD;
  if (!email || !password) {
    throw new Error('EH_K6_EMAIL and EH_K6_PASSWORD are required');
  }
  const loginRes = http.post(
    `${baseURL}/api/v1/auth/login`,
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
  return { token };
}

export default function (data) {
  const headers = {
    Authorization: `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  const shopRes = http.get(`${baseURL}/api/v1/shop/items`, { headers });
  const shopOK = check(shopRes, {
    'shop items 200': (r) => r.status === 200,
    'shop items is array': (r) => Array.isArray(r.json()),
  });
  errorRate.add(!shopOK);

  const invRes = http.get(`${baseURL}/api/v1/inventory/items`, { headers });
  const invOK = check(invRes, {
    'inventory items 200': (r) => r.status === 200,
    'inventory items has items': (r) => {
      const body = r.json();
      return body && Array.isArray(body.items);
    },
  });
  errorRate.add(!invOK);

  sleep(0.3);
}
