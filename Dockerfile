# Pinned to the patch release that carries the current standard library
# security fixes. Bump it deliberately, not implicitly through a floating tag.
FROM golang:1.26.6-alpine AS builder

WORKDIR /src

# Copy the manifests on their own first so the dependency download is cached
# independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO is off so the binary is static and runs on a minimal base image.
# -trimpath keeps build machine paths out of panics.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pastebomb ./cmd/api

FROM alpine:3.22

# ca-certificates is needed for any outbound TLS; wget backs the healthcheck.
RUN apk add --no-cache ca-certificates wget \
    && adduser -D -u 10001 pastebomb \
    && mkdir -p /var/log/pastebomb \
    && chown pastebomb:pastebomb /var/log/pastebomb

COPY --from=builder /out/pastebomb /usr/local/bin/pastebomb

ENV AUDIT_LOG_DIR=/var/log/pastebomb
WORKDIR /var/log/pastebomb

# Running as a non-root user means a remote code execution bug does not start
# out with permission to rewrite the image.
USER pastebomb:pastebomb

EXPOSE 8000

# A GET, not wget's --spider, which issues a HEAD the router does not serve.
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --output-document=/dev/null http://127.0.0.1:8000/health || exit 1

ENTRYPOINT ["/usr/local/bin/pastebomb"]
