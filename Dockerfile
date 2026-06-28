FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/

FROM alpine:3.20

COPY --from=builder /app/app /app/app

EXPOSE 8000

CMD ["/app/app"]
