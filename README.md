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

Run tests (*optional*)
```sh
go test -count=1 ./... -v
```
