.PHONY: build run test vet tidy fmt cover docker-up docker-down migrate

GO ?= go
PKG := ./...

build:
	$(GO) build -o bin/app ./cmd

run:
	$(GO) run ./cmd

test:
	$(GO) test -count=1 $(PKG)

vet:
	$(GO) vet $(PKG)

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt $(PKG)

cover:
	$(GO) test -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -html=coverage.out

docker-up:
	docker compose --profile all up -d

docker-down:
	docker compose --profile all down

migrate:
	migrate -path database/migrations -database "mysql://user:1234@tcp(0.0.0.0:3306)/go_gin_gonic?charset=utf8mb4&parseTime=True&loc=Local" up
