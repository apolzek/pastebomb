# pastebomb

### Run with docker compose

Up app and database
```sh
docker-compose --profile all up
```
### Run app localy

Up database and start app
```sh
docker-compose --profile depends up   

go mod tidy
go run cmd/main.go
# or 
air
```
> http://localhost:8000/health

Run migrations
```sh
migrate -path database/migrations/ -database "mysql://user:1234@tcp(0.0.0.0:3306)/go_gin_gonic?charset=utf8mb4&parseTime=True&loc=Local" up
```

## Testing

The test suite is made of **pure unit tests** — no database, no `.env`, no network. You can run it from a fresh clone without `docker-compose up`.

Run everything:
```sh
go test ./...
```

Verbose output (recommended while developing):
```sh
go test -count=1 ./... -v
```

Run a single package:
```sh
go test -v ./utils/...
go test -v ./security/...
```

Run a single test by name:
```sh
go test -v ./security/... -run TestGenerateAndDecodeToken
```

With coverage:
```sh
go test -cover ./...
# or, for an HTML report:
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### What is covered

| Package    | Tests                                                                                          |
|------------|------------------------------------------------------------------------------------------------|
| `utils`    | `GenerateRandomString`, `EncodeContent` (base64), `RandomString`, `RandomFileName`, `FileValidationByExtension`, `RemoveFile` |
| `security` | `GenerateToken` / `VerifyToken` / `DecodeToken` — happy path, malformed token, wrong signing method |

Controller, repository and database tests were intentionally removed because they required a live MySQL instance and a populated `.env`, which made `go test ./...` fail on any machine that was not already running the full stack. Add them back as proper integration tests (e.g. behind a `//go:build integration` tag) when the project has a test database fixture.
