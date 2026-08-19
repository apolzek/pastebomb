// Read-heavy load: ramps virtual users against the endpoints a pastebin
// actually serves in bulk. Nothing here writes, so the run isolates read path
// cost (query, base64 decode, JSON encoding) from insert contention.
import { sleep } from 'k6';
import { api } from './lib/api.js';

export const options = {
  stages: [
    { duration: __ENV.RAMP_UP || '30s', target: Number(__ENV.VUS || 50) },
    { duration: __ENV.STEADY || '1m', target: Number(__ENV.VUS || 50) },
    { duration: __ENV.RAMP_DOWN || '15s', target: 0 },
  ],
  thresholds: {
    business_errors: ['rate<0.01'],
    http_req_failed: ['rate<0.01'],
    // Reads are index lookups: anything slower points at the database or at
    // connection pool starvation.
    'read_latency{endpoint:read_paste}': ['p(95)<200'],
    'read_latency{endpoint:public_index}': ['p(95)<300'],
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
  },
};

export default function () {
  // Weighted to match a real pastebin: most traffic follows a shared link to
  // one paste, a minority browses the index.
  const roll = Math.random();
  if (roll < 0.55) {
    api.readPaste();
  } else if (roll < 0.8) {
    api.readRaw();
  } else if (roll < 0.95) {
    api.publicIndex(Math.ceil(Math.random() * 3));
  } else {
    api.readProtectedWithoutSecret();
  }

  sleep(Math.random() * 0.5);
}
