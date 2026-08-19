# Load testing

Three tools, because they answer different questions. Running all three is not
redundancy: when they disagree, the disagreement is itself the finding.

| Tool | Model | Answers |
|---|---|---|
| **k6** | Virtual users and arrival rates, JavaScript scenarios with pass/fail thresholds | "Does the service still meet its latency budget under a realistic traffic mix?" Good in CI, because a threshold breach exits non-zero. |
| **Locust** | Virtual users with think time, Python task weights, live web UI | "What does the service look like with N concurrent human-shaped users?" Best for exploring interactively and for workloads with real session state. |
| **Vegeta** | Constant request rate, no think time | "What exactly is latency at N requests per second?" The most precise of the three, because it holds the offered rate regardless of how the service responds, so it cannot mask a slowdown by backing off. |

A practical division of labour: use **Vegeta** to establish ground truth for a
single endpoint, **k6** to guard those numbers in CI, and **Locust** to explore
a mixed workload while watching it happen.

## Contents

```
loadtest/
  seed.sh                        creates the account, token and paste handles the tools need
  docker-compose.loadtest.yaml   compose override that raises the rate limits
  k6/
    lib/api.js                   shared client, metrics and expected statuses
    smoke.js                     one pass over every endpoint, strict thresholds, CI friendly
    read.js                      ramping read-heavy load
    mixed.js                     readers, writers and signed-in users at fixed arrival rates
    login.js                     login only, to find the bcrypt throughput ceiling
  locust/
    locustfile.py                weighted user classes
    requirements.txt
  vegeta/
    run.sh                       target generation plus attack, report, histogram and plot
  results/                       generated output, ignored by git
```

## Setup

### 1. Raise the rate limits

The shipped limits are production values: one login attempt per five seconds,
one anonymous write per second, fifty public reads per second, all per client
IP. A load test from a single host hits them immediately, and the run then
measures the limiter rather than the service.

```sh
docker compose -f docker-compose.yaml -f loadtest/docker-compose.loadtest.yaml \
  --profile all up -d --build
```

Never apply that override to an environment reachable from the internet.

> If a run reports a large `rate_limited` count or a wall of 429s, this step is
> missing. That is the limiter working, not the service failing.

### 2. Seed the data

Paste handles are random and only exist once created, so they cannot be
hardcoded in a target file. Seeding is a separate step so the tools spend their
run generating load rather than registering users.

```sh
./loadtest/seed.sh                       # defaults to http://localhost:8000
./loadtest/seed.sh http://staging:8000   # or point it somewhere else
```

It writes `loadtest/.seed.env` (git-ignored) with the base URL, an account, its
bearer token, ten readable paste handles and one secret-protected handle. All
three tools read it.

## k6

Install: `brew install k6`, `winget install k6`, or see
<https://grafana.com/docs/k6/latest/set-up/install-k6/>. No install needed with
Docker, shown below.

k6 takes configuration through the environment, so export the seed file first:

```sh
set -a; source loadtest/.seed.env; set +a
```

```sh
# One pass over every endpoint. Fails on any error. Run this in CI.
k6 run loadtest/k6/smoke.js

# Read-heavy ramp: 50 virtual users by default.
VUS=100 RAMP_UP=30s STEADY=2m RAMP_DOWN=15s k6 run loadtest/k6/read.js

# Realistic mix at fixed arrival rates.
READ_RATE=200 WRITE_RATE=40 AUTH_RATE=15 DURATION=2m k6 run loadtest/k6/mixed.js

# Login only, to find the bcrypt ceiling.
VUS=16 DURATION=1m k6 run loadtest/k6/login.js
```

With Docker instead of a local install:

```sh
set -a; source loadtest/.seed.env; set +a
docker run --rm --network host -v "$PWD/loadtest:/loadtest:ro" \
  -e BASE_URL -e TOKEN -e SLUGS -e SECRET_SLUG -e SECRET -e EMAIL -e PASSWORD -e USERNAME \
  grafana/k6 run /loadtest/k6/smoke.js
```

**Reading the output.** Two custom metrics matter more than the built-ins:

* `business_errors` is the real error rate. A request the service was right to
  reject, such as a 401 on the deliberately unauthorized read, is not counted.
* `rate_limited` counts 429s separately, so throttling never inflates the error
  rate.
* `dropped_iterations` above zero means k6 could not keep up. The run then
  understated the offered load, and its latency numbers should be discarded
  rather than interpreted.

## Locust

```sh
pip install -r loadtest/locust/requirements.txt
```

Interactive, with the web UI on <http://localhost:8089>:

```sh
locust -f loadtest/locust/locustfile.py --host http://localhost:8000
```

Headless, which is what a repeatable run wants:

```sh
locust -f loadtest/locust/locustfile.py --host http://localhost:8000 \
  --headless --users 100 --spawn-rate 10 --run-time 2m
```

With Docker, no Python install required:

```sh
docker run --rm --network host -v "$PWD/loadtest:/mnt/locust:ro" locustio/locust \
  -f /mnt/locust/locust/locustfile.py --host http://localhost:8000 \
  --headless --users 100 --spawn-rate 10 --run-time 2m
```

Distributed across cores, since one Locust process is single-threaded and
becomes the bottleneck well before the service does:

```sh
locust -f loadtest/locust/locustfile.py --master &
locust -f loadtest/locust/locustfile.py --worker &   # once per core
```

The user classes are weighted to resemble real traffic: `AnonymousReader` (8),
`AnonymousWriter` (2), `AuthenticatedUser` (1) and `HealthCheckUser` (1).
Change the `weight` attributes to reshape the mix. Locust reads
`loadtest/.seed.env` directly, and environment variables override it.

## Vegeta

```sh
go install github.com/tsenart/vegeta/v12@latest
```

```sh
./loadtest/vegeta/run.sh <scenario> [rate] [duration]

./loadtest/vegeta/run.sh read 500 1m     # 500 reads per second for a minute
./loadtest/vegeta/run.sh write 50 30s
./loadtest/vegeta/run.sh mixed 300 1m
./loadtest/vegeta/run.sh all             # every scenario at the defaults
```

Scenarios: `read`, `raw`, `index`, `write`, `profile`, `mixed`, `all`.
Defaults: 200 requests per second for 30 seconds.

Each run prints a summary and a latency histogram, and leaves the raw results
in `loadtest/results/` so they can be re-reported without attacking again:

```sh
vegeta report loadtest/results/read.bin
vegeta report -type=json loadtest/results/read.bin
vegeta plot loadtest/results/read.bin > plot.html
```

**Finding the knee.** Vegeta's constant rate makes capacity easy to bracket:
step the rate up until latency climbs or the success ratio drops.

```sh
for rate in 100 250 500 1000 2000 4000; do
  ./loadtest/vegeta/run.sh read $rate 15s | grep -E 'Latencies|Success'
done
```

## Baseline

Measured on the compose stack with MySQL 8.4, on a 24-core developer machine
with the service, the database and the load generator all on the same host.
Treat these as the shape to expect, not as a target to hit on other hardware.

| Endpoint | p50 | p95 | Notes |
|---|---|---|---|
| `GET /u/anonymous/:slug` | 0.35 ms | 0.6 ms | Sustained 2000 req/s with idle CPU |
| `GET /u/anonymous/:slug/raw` | 0.4 ms | 0.7 ms | |
| `GET /u/anonymous/posts` | 0.5 ms | 0.9 ms | |
| `POST /u/anonymous/post` | 2.9 ms | 4.1 ms | Dominated by the InnoDB commit |
| `GET /u/anonymous/:slug` (secret-protected) | 160 ms | 165 ms | One bcrypt comparison at cost 12 |
| `POST /login` | 164 ms | 169 ms | One bcrypt comparison; ceiling near 49 logins/s at 8 concurrent |

Two things fall out of that table.

**Password hashing is the entire cost model.** Every other endpoint is a sub-
millisecond index lookup; anything touching bcrypt is three orders of magnitude
slower, and deliberately so. Login throughput is therefore roughly
`cores / hash time`, so `BCRYPT_COST` is a direct throughput decision rather
than a free security dial. Use `loadtest/k6/login.js` to measure the ceiling
before changing it.

**That cost is why the public reads are rate limited.** Reading a
secret-protected paste costs about 250 times an ordinary read on a route that
requires no credential, which is an amplification lever: a handful of requests
per second from one client can occupy a core. `RATE_LIMIT_READ_RPS` exists to
bound it. This limit was added after the first load test run made the asymmetry
visible.

## Interpreting a disagreement between the tools

The tools measure the same service, so a large gap between them is usually the
load generator, not the service. The most common cause is under-provisioned
virtual users: k6's arrival-rate executors report generator-side queueing as
request latency, which showed up here as a 316 ms write p95 that Vegeta
measured at 4 ms at the same offered rate. `dropped_iterations` is the tell.

The first seconds of a run are also unrepresentative. Writes show a one-off
spike while the database connection pool grows from its idle size to what the
load demands, so a `max` in the hundreds of milliseconds alongside a median of
a few milliseconds is warm-up, not a tail. Judge a run by its median and p95
over a duration long enough to dwarf that ramp, a minute or more, rather than
by its maximum.

When k6 and Locust look slow but Vegeta does not, suspect the generator.
When all three agree, believe them.

## In CI

The smoke test is the one worth running on every change: it is fast, it
exercises every endpoint, and its thresholds fail the build on any error.

```sh
docker compose -f docker-compose.yaml -f loadtest/docker-compose.loadtest.yaml \
  --profile all up -d --build
./loadtest/seed.sh
set -a; source loadtest/.seed.env; set +a
k6 run loadtest/k6/smoke.js       # non-zero exit if a threshold is crossed
```

Full load runs belong on a schedule against a dedicated environment, not on
pull requests: on a shared CI runner the generator and the service compete for
the same cores, and the numbers stop meaning anything.
