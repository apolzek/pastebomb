.PHONY: help build run test test-race cover vet fmt fmt-check lint tidy vuln check docker-up docker-down migrate migrate-down \
	loadtest-up loadtest-down loadtest-seed loadtest-smoke loadtest-k6 loadtest-locust loadtest-vegeta

GO         ?= go
PKG        := ./...
BIN        := bin/pastebomb
DB_USER    ?= user
DB_PASSWORD ?= $(shell grep -E '^DB_PASSWORD=' .env 2>/dev/null | cut -d= -f2- | tr -d '"')
DB_NAME    ?= pastebomb
DB_DSN     := mysql://$(DB_USER):$(DB_PASSWORD)@tcp(127.0.0.1:3306)/$(DB_NAME)?charset=utf8mb4&parseTime=True&loc=UTC

help: ## List the available targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Compile the API binary into bin/
	$(GO) build -trimpath -o $(BIN) ./cmd/api

run: ## Run the API from source
	$(GO) run ./cmd/api

test: ## Run the unit test suite
	$(GO) test -count=1 $(PKG)

test-race: ## Run the suite under the race detector
	$(GO) test -count=1 -race $(PKG)

cover: ## Produce an HTML coverage report
	$(GO) test -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "report written to coverage.html"

vet: ## Run go vet
	$(GO) vet $(PKG)

fmt: ## Format the tree
	$(GO) fmt $(PKG)

fmt-check: ## Fail if anything is unformatted
	@unformatted=$$(gofmt -l . | grep -v '^$$'); \
	if [ -n "$$unformatted" ]; then echo "unformatted files:"; echo "$$unformatted"; exit 1; fi

tidy: ## Sync go.mod and go.sum
	$(GO) mod tidy

vuln: ## Scan dependencies and the standard library for known vulnerabilities
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKG)

check: fmt-check vet test-race vuln ## Everything CI runs

docker-up: ## Start the full stack
	docker compose --profile all up -d --build

docker-down: ## Stop the stack
	docker compose --profile all down

migrate: ## Apply the pending migrations against a local database
	migrate -path database/migrations -database "$(DB_DSN)" up

migrate-down: ## Roll back the most recent migration
	migrate -path database/migrations -database "$(DB_DSN)" down 1

# ---------------------------------------------------------------------------
# Load testing. See loadtest/README.md.
#
# The stack must run with the rate limit override, otherwise every run measures
# the limiter instead of the service.
# ---------------------------------------------------------------------------

COMPOSE_LOADTEST := docker compose -f docker-compose.yaml -f loadtest/docker-compose.loadtest.yaml

loadtest-up: ## Start the stack with rate limits raised for benchmarking
	$(COMPOSE_LOADTEST) --profile all up -d --build

loadtest-down: ## Stop the benchmarking stack
	$(COMPOSE_LOADTEST) --profile all down

loadtest-seed: ## Create the account, token and paste handles the tools need
	./loadtest/seed.sh

loadtest-smoke: ## One k6 pass over every endpoint, fails on any error
	set -a; . ./loadtest/.seed.env; set +a; k6 run loadtest/k6/smoke.js

loadtest-k6: ## k6 mixed workload (READ_RATE, WRITE_RATE, AUTH_RATE, DURATION)
	set -a; . ./loadtest/.seed.env; set +a; k6 run loadtest/k6/mixed.js

loadtest-locust: ## Locust headless run (USERS, SPAWN_RATE, RUN_TIME)
	locust -f loadtest/locust/locustfile.py --host $(or $(BASE_URL),http://localhost:8000) \
		--headless --users $(or $(USERS),100) --spawn-rate $(or $(SPAWN_RATE),10) --run-time $(or $(RUN_TIME),2m)

loadtest-vegeta: ## Vegeta run (SCENARIO, RATE, DURATION)
	./loadtest/vegeta/run.sh $(or $(SCENARIO),mixed) $(or $(RATE),200) $(or $(DURATION),30s)
