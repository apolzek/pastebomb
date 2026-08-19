# pastebomb

A pastebin API in Go: anonymous or authenticated pastes, unguessable handles,
optional access secrets and self-expiring content.

## Architecture

The codebase follows domain-driven design with a strict dependency rule: every
import points inward. The domain has no idea that Gin, GORM or HTTP exist.

```
cmd/api                        process entry point, signal handling

internal/
  domain/                      the model, free of any framework
    shared/                    errors, pagination, clock and hasher ports
    user/                      User aggregate, value objects, repository port
    paste/                     Paste aggregate, access guard, repository port

  application/                 use cases, one method per user intention
                               plus the outbound ports (tokens, audit trail)

  infrastructure/              the adapters that implement those ports
    config/                    environment loading and validation
    persistence/               GORM repositories and row mapping
    security/                  bcrypt, JWT, crypto/rand handle generation
    audit/                     rotating JSON audit log

  interfaces/http/             delivery layer
    handler/                   one handler per endpoint
    middleware/                auth, rate limiting, logging, headers
    dto/                       the wire contract
    httperr/                   domain error to HTTP status mapping
    reqctx/                    request-scoped values

  bootstrap/                   composition root, wires everything together
  testsupport/                 in-memory repositories and deterministic fakes
```

Layer rules that the code holds to:

* **Aggregates own their invariants.** Every field is private, and the only way
  to build a `User` or a `Paste` is through a constructor that validates. There
  is no path to an invalid instance.
* **Value objects instead of bare strings.** `Email`, `Username`, `Slug`,
  `Expiration` and `SecretHash` each carry their own rules, so a validation
  cannot be forgotten at one of several call sites.
* **Ports in the domain, adapters in infrastructure.** The domain declares that
  it needs a `Hasher` and a `SlugGenerator`; infrastructure decides that means
  bcrypt and `crypto/rand`.
* **No package-level mutable state.** Configuration, the database handle and
  the clock are all injected, which is what makes the whole stack testable
  without a database.

## Running it

Copy the environment template and fill it in. `JWT_SECRET` and `DB_PASSWORD`
have no defaults: the process refuses to start without them.

```sh
cp .env.example .env
openssl rand -hex 32   # paste the result into JWT_SECRET
```

### With docker compose

```sh
docker compose --profile all up --build
```

### Locally, against the containerized database

```sh
docker compose --profile depends up -d
go run ./cmd/api
# or, with live reload:
air
```

> http://localhost:8000/health

### Migrations

Compose runs them automatically. To apply them by hand:

```sh
make migrate
```

## Configuration

Every setting is read once at startup and validated; a bad value fails the boot
with a list of the problems rather than surfacing later under load.

| Variable | Default | Notes |
|---|---|---|
| `APP_PORT` | `:8000` | Listen address |
| `DB_DRIVER` | `mysql` | `mysql` or `pgsql` (the migrations are MySQL-only) |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER` | `localhost`, `3306`, `pastebomb`, `user` | |
| `DB_PASSWORD` | none | **Required** |
| `JWT_SECRET` | none | **Required**, at least 32 characters, example values are rejected |
| `JWT_ISSUER` | `pastebomb` | Verified on every incoming token |
| `JWT_TTL` | `24h` | Token lifetime |
| `BCRYPT_COST` | `12` | Must be between 10 and 15 |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:8080` | Comma-separated, `*` disables the check |
| `HTTP_TRUSTED_PROXIES` | empty | Only these may set `X-Forwarded-For` |
| `HTTP_MAX_BODY_BYTES` | `2097152` | Request body cap |
| `RATE_LIMIT_LOGIN_RPS` / `_BURST` | `0.2` / `5` | Per client, login |
| `RATE_LIMIT_ANON_RPS` / `_BURST` | `1` / `10` | Per client, anonymous writes |
| `RATE_LIMIT_READ_RPS` / `_BURST` | `50` / `100` | Per client, public reads |
| `PURGE_INTERVAL` | `1m` | How often expired pastes are swept |
| `AUDIT_LOG_DIR` | `.` | Where `audit.log` is written |
| `LOG_LEVEL` | info | Set to `debug` for verbose logs |

## API

Successful responses are `{"message": ..., "data": ..., "meta": ...}`; failures
are `{"message": ..., "code": ..., "request_id": ...}` where `code` is stable
and machine-readable.

| Method | Path | Access | Purpose |
|---|---|---|---|
| `GET` | `/health` | public | Liveness |
| `GET` | `/ready` | public | Readiness, checks the database |
| `POST` | `/login` | public, rate limited | Issue a token |
| `POST` | `/u` | public, rate limited | Register |
| `GET` | `/u/:handle` | public | Public profile by id or username |
| `GET` | `/u/anonymous/posts` | public | Public index, metadata only |
| `POST` | `/u/anonymous/post` | public, rate limited | Create an anonymous paste |
| `GET` | `/u/anonymous/:slug` | public | Read a paste |
| `GET` | `/u/anonymous/:slug/raw` | public | Read the body as `text/plain` |
| `GET` | `/u/me` | token | My profile |
| `PATCH` | `/u/me` | token | Update my profile |
| `DELETE` | `/u/me`, `/u/:id` | token | Deactivate my account |
| `GET` | `/u/me/posts` | token | My pastes, metadata only |
| `GET` | `/u/me/posts/content` | token | My pastes with bodies |
| `POST` | `/u/me/post` | token | Create a paste I own |
| `GET` | `/u/all`, `/u`, `/u/paginate` | admin token | List accounts |

Pastes protected by a secret are read by sending it in the `X-Post-Secret`
header, which keeps it out of URLs and therefore out of access logs.

Admin rights come from the `is_admin` claim in the token, which is set from the
`users.is_admin` column. Grant it out of band:

```sql
UPDATE users SET is_admin = 1 WHERE email = 'you@example.com';
```

`RESTClient/` has ready-made requests for the VS Code REST Client extension and
a Postman collection.

## Security posture

* Passwords and paste secrets are bcrypt hashed; neither is ever returned or
  logged, and the audit trail redacts credential-shaped fields.
* Pastes are addressed by a 12-character `crypto/rand` handle. Numeric ids are
  never accepted, so pastes cannot be enumerated.
* Expired pastes stop being served immediately; the purge job only reclaims
  storage.
* Login failures are indistinguishable from one another, so the endpoint cannot
  be used to discover which addresses are registered.
* Tokens pin the signing algorithm and the issuer, which closes the `alg: none`
  and algorithm-confusion attacks.
* Login, anonymous writes and public reads are rate limited per client, with
  bounded memory. Reads are limited because a secret-protected read costs a
  bcrypt comparison, roughly 250 times an ordinary read, on a route that needs
  no credential.
* Every response carries `nosniff`, `X-Frame-Options`, a restrictive CSP and a
  correlation id. Raw pastes are served as a plain-text attachment so a paste
  containing HTML cannot execute in a browser.
* The server has explicit timeouts, a request body cap and graceful shutdown.

Run the dependency and standard library scan with:

```sh
make vuln
```

## Testing

The suite is pure unit and integration-in-memory: no database, no `.env`, no
network. It runs from a fresh clone.

```sh
make test        # go test ./...
make test-race   # under the race detector
make cover       # HTML coverage report
make check       # formatting, vet, race tests and the vulnerability scan
```

What is covered:

| Package | Focus |
|---|---|
| `domain/shared` | Error classification, pagination clamping |
| `domain/user` | Value object rules, registration, authentication, deactivation |
| `domain/paste` | Creation rules, expiration, secret unlocking, access guard |
| `application` | Use cases against in-memory repositories: uniqueness, ownership, purging |
| `infrastructure/security` | JWT round trip, `alg: none`, wrong key, wrong issuer, expiry; bcrypt salting; handle uniqueness |
| `infrastructure/config` | Required settings, weak secrets, out-of-range values |
| `infrastructure/audit` | JSON output and credential redaction |
| `interfaces/http` | The real router end to end: authentication, admin gating, secret-protected reads, index leakage, pagination, headers |
| `interfaces/http/middleware` | Rate limiting and bucket eviction |

The GORM repository adapters are the one layer without automated coverage: they
need a live MySQL instance. Add them behind a `//go:build integration` tag when
a test database fixture exists.

## Load testing

`loadtest/` holds three suites, one per tool, sharing a seeding script:

```sh
make loadtest-up      # stack with the rate limits raised for benchmarking
make loadtest-seed    # account, token and paste handles
make loadtest-smoke   # k6 pass over every endpoint, fails on any error
```

k6 guards latency budgets in CI, Locust explores a mixed workload
interactively, and Vegeta measures precise latency at a fixed request rate.
Measured baselines and how to interpret a disagreement between the tools are in
[loadtest/README.md](loadtest/README.md).
