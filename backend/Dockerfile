# ─── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# ─── Runtime stage ─────────────────────────────────────────────────────────────
FROM scratch

# ca-certificates needed for TLS connections (RabbitMQ, Postgres)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
