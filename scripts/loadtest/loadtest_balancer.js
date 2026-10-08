/**
 * LEGACY balancer blast — retargeted to public edge :8079.
 * Prefer CORE: make test-k6 (browse.js) / make test-k6-purchase.
 *
 *   BASE_URL=http://localhost:8079 k6 run scripts/loadtest/loadtest_balancer.js
 */
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '30s', target: 50 },
    { duration: '30s', target: 100 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8079';
const USER_EMAIL = __ENV.EH_K6_EMAIL || 'k6load@test.com';
const USER_PASSWORD = __ENV.EH_K6_PASSWORD || 'secret123';
const USER_ID = __ENV.EH_K6_USER_ID || 'ccd79af5-9fd9-45db-961a-818a577164ee';

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/api/auth/login`,
    JSON.stringify({ email: USER_EMAIL, password: USER_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  let token = '';
  if (loginRes.status === 200) {
    token = loginRes.json('access_token');
    console.log('Token obtained');
  } else {
    console.warn(`login failed: ${loginRes.status} (legacy script — use browse.js for CORE)`);
  }
  return { token };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${data.token}`,
  };

  for (const game_id of ['hexagon', 'memory', 'flappy', 'towers']) {
    http.post(
      `${BASE_URL}/api/game/submit`,
      JSON.stringify({
        user_id: USER_ID,
        game_id,
        level: 1,
        score: Math.floor(Math.random() * 200) + 20,
        user_email: USER_EMAIL,
        nickname: 'LoadTest',
        seed: `${game_id}_${__VU}_${Date.now()}`,
        moves: [],
      }),
      { headers, timeout: '5s' },
    );
  }

  sleep(1);
}
