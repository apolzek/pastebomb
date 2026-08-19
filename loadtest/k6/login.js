// Login-only load. Kept separate because /login runs bcrypt at the configured
// cost, so it is orders of magnitude more expensive than any other endpoint
// and would otherwise dominate a mixed run's latency percentiles.
//
// Its throughput ceiling is a deliberate product of BCRYPT_COST: roughly
// (cores / hash time) logins per second. Use this script to find that ceiling
// before choosing a cost.
import { api } from './lib/api.js';

export const options = {
  vus: Number(__ENV.VUS || 10),
  duration: __ENV.DURATION || '30s',
  thresholds: {
    business_errors: ['rate<0.01'],
    // Generous on purpose: bcrypt at cost 12 is tens of milliseconds of CPU
    // per attempt by design.
    'write_latency{endpoint:login}': ['p(95)<2000'],
  },
};

export default function () {
  api.login(__ENV.EMAIL, __ENV.PASSWORD);
}
