FROM golang:1.21-alpine as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/

# new image
FROM alpine:3.14

COPY --from=builder /app/app /app/app

EXPOSE 8000

CMD ["/app/app"]
