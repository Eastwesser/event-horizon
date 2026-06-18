// loadtest.js
import http from 'k6/http';
import { sleep } from 'k6';

const TARGET = __ENV.TARGET ? parseInt(__ENV.TARGET) : 100;

export const options = {
    stages: [
        { duration: '30s', target: TARGET },
        { duration: '60s', target: TARGET },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<5000'],
    },
};

const BASE_URL = 'http://localhost:8080';
const USER_EMAIL = 'k6load@test.com';
const USER_PASSWORD = 'secret123';
const USER_ID = 'ccd79af5-9fd9-45db-961a-818a577164ee';

let token = '';

export function setup() {
    const loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
        email: USER_EMAIL,
        password: USER_PASSWORD,
    }), { headers: { 'Content-Type': 'application/json' } });

    if (loginRes.status === 200) {
        token = loginRes.json('access_token');
        console.log(`✅ Token obtained`);
    }
    return { token };
}

export default function (data) {
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
    };

    // 1. Hexagon
    http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: USER_ID,
        game_id: 'hexagon',
        level: 1,
        score: Math.floor(Math.random() * 500) + 50,
        user_email: USER_EMAIL,
        nickname: `LoadTest`,
        seed: `seed_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers, timeout: '5s' });

    // 2. Memory
    http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: USER_ID,
        game_id: 'memory',
        level: 1,
        score: Math.floor(Math.random() * 900) + 100,
        user_email: USER_EMAIL,
        nickname: `LoadTest`,
        seed: `memory_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers, timeout: '5s' });

    // 3. Flappy
    http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: USER_ID,
        game_id: 'flappy',
        level: 1,
        score: Math.floor(Math.random() * 50) + 10,
        user_email: USER_EMAIL,
        nickname: `LoadTest`,
        seed: `flappy_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers, timeout: '5s' });

    // 4. Towers
    http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: USER_ID,
        game_id: 'towers',
        level: 1,
        score: Math.floor(Math.random() * 500) + 50,
        user_email: USER_EMAIL,
        nickname: `LoadTest`,
        seed: `towers_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers, timeout: '5s' });

    sleep(0.5);
}