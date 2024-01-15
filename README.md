# pastebomb

#### This project uses:

- mysql
- gin-gonic
- migrate
- jwt

#### Steps

1. Up database with docker
```sh
docker compose --profile depends up
docker compose --profile all up    
```

2. Run migrations
```sh
migrate -path database/migrations/ -database "mysql://user:1234@tcp(0.0.0.0:3306)/go_gin_gonic?charset=utf8mb4&parseTime=True&loc=Local" up
```

3. Run app localy
```sh
go mod tidy
go run cmd/main.go
# or 
air
```

4. Run tests
```sh
go test -count=1 ./... -v
```
