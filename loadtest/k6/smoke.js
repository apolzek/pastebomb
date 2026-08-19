// Smoke test: one virtual user walks every endpoint once. It answers "is the
// service healthy enough to be worth load testing?" and is the right thing to
// run in CI, where a full load run would be noise.
import { sleep } from 'k6';
import { api } from './lib/api.js';

export const options = {
  vus: 1,
  iterations: 20,
  thresholds: {
    // A smoke run must be perfect: any failure is a real defect, not load.
    business_errors: ['rate==0'],
    http_req_failed: ['rate==0'],
    http_req_duration: ['p(95)<1000'],
  },
};

export default function () {
  api.health();
  api.ready();
  api.publicIndex();
  api.readPaste();
  api.readRaw();
  api.readProtectedWithoutSecret();
  api.readProtectedWithSecret();
  api.publicProfile(__ENV.USERNAME || 'nobody');
  api.profile();
  api.myPastes();
  api.createAnonymousPaste();
  api.createOwnedPaste();
  api.login(__ENV.EMAIL || '', __ENV.PASSWORD || '');

  sleep(0.1);
}
