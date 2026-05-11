import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 100 },   // разогрев до 100
        { duration: '1m', target: 1000 },   // подъём до 1000
        { duration: '1m', target: 1000 },   // стабильная нагрузка
        { duration: '30s', target: 0 },     // спад
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],   // 95% запросов < 500ms
        http_req_failed: ['rate<0.01'],     // ошибки < 1%
    },
};

export default function() {
    const payload = JSON.stringify({
        user_id: `user-${__VU}`,
        game_id: 'hexagon',
        level: 3,
        seed: `seed-${__VU}`,
        moves: [],
    });
    
    const res = http.post('http://localhost:8080/api/game/submit', payload, {
        headers: { 'Content-Type': 'application/json' },
    });
    
    check(res, {
        'status is 200': (r) => r.status === 200,
        'success is true': (r) => JSON.parse(r.body).success === true,
    });
    
    sleep(0.1);
}