// Mixed workload: independent scenarios running side by side at fixed arrival
// rates, which is how the service is actually used. Arrival rate rather than
// virtual users means a slowdown shows up as a queue instead of quietly
// reducing the offered load.
import { sleep } from 'k6';
import { api } from './lib/api.js';

const readRate = Number(__ENV.READ_RATE || 100);
const writeRate = Number(__ENV.WRITE_RATE || 20);
const authRate = Number(__ENV.AUTH_RATE || 10);
const duration = __ENV.DURATION || '1m';

export const options = {
  scenarios: {
    readers: {
      executor: 'constant-arrival-rate',
      rate: readRate,
      timeUnit: '1s',
      duration,
      // Virtual users needed is roughly rate multiplied by latency, and a
      // read takes well under a millisecond. Over-allocating them costs the
      // load generator more than it costs the service, which shows up as
      // phantom latency in the results.
      preAllocatedVUs: Math.max(30, Math.ceil(readRate / 4)),
      maxVUs: Math.max(60, readRate),
      exec: 'read',
    },
    writers: {
      executor: 'constant-arrival-rate',
      rate: writeRate,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: Math.max(20, writeRate),
      maxVUs: Math.max(40, writeRate * 2),
      exec: 'write',
    },
    authenticated: {
      executor: 'constant-arrival-rate',
      rate: authRate,
      timeUnit: '1s',
      duration,
      // This one sleeps between requests, so it genuinely needs VUs.
      preAllocatedVUs: Math.max(20, authRate * 4),
      maxVUs: Math.max(60, authRate * 8),
      exec: 'authenticated',
    },
  },
  // Latency thresholds encode an expectation about a specific environment.
  // These are calibrated for the compose stack on a developer machine, where a
  // read measures well under a millisecond and a write a few milliseconds; the
  // headroom absorbs a busy laptop. Re-baseline them against your own
  // environment rather than treating them as universal.
  thresholds: {
    business_errors: ['rate<0.01'],
    http_req_failed: ['rate<0.01'],
    'read_latency{endpoint:read_paste}': ['p(95)<100'],
    'write_latency{endpoint:create_anonymous}': ['p(95)<250'],
    http_req_duration: ['p(95)<500'],
    // A dropped iteration means the load generator could not keep up, so the
    // run understated the offered load and its numbers cannot be trusted.
    dropped_iterations: ['count<10'],
  },
};

export function read() {
  Math.random() < 0.7 ? api.readPaste() : api.publicIndex();
}

export function write() {
  api.createAnonymousPaste();
}

export function authenticated() {
  api.profile();
  api.myPastes();
  if (Math.random() < 0.3) {
    api.createOwnedPaste();
  }
  sleep(0.2);
}
