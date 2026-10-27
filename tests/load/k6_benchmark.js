import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '5s', target: 50 },  // Ramp-up
    { duration: '15s', target: 200 }, // High-concurrency sustained
    { duration: '5s', target: 0 },   // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<50'], // 95% of requests must complete below 50ms
    http_req_failed: ['rate<0.01'],  // Error rate < 1%
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8080/api/v1';

export default function () {
  // 1. Homepage Aggregated Feed
  const homeRes = http.get(`${BASE_URL}/homepage`);
  check(homeRes, {
    'homepage status is 200': (r) => r.status === 200,
    'homepage response time < 50ms': (r) => r.timings.duration < 50,
  });

  // 2. Paginated Articles Feed
  const feedRes = http.get(`${BASE_URL}/articles?limit=12`);
  check(feedRes, {
    'articles status is 200': (r) => r.status === 200,
  });

  // 3. Trending Articles
  const trendingRes = http.get(`${BASE_URL}/trending?window=today&limit=5`);
  check(trendingRes, {
    'trending status is 200': (r) => r.status === 200,
  });

  sleep(0.01);
}
