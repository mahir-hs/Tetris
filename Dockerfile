# ---- Builder Stage ----
FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o tetris ./cmd/tetris

# ---- Runtime Stage (tiny: ~15MB) ----
FROM alpine:3.20
WORKDIR /app
# Install gotty (terminal→WebSocket)
RUN apk add --no-cache curl && \
    curl -fsSL https://github.com/yudai/gotty/releases/download/v2.0.0/gotty_2.0.0_linux_amd64.tar.gz | tar -xz -C /usr/local/bin
COPY --from=builder /app/tetris .
COPY --from=builder /app/settings.json .
# SQLite DB path (persistent volume on Koyeb)
ENV TETRIS_DB_PATH=/data/tetris.db
# Koyeb injects PORT env var (default 8080)
EXPOSE 8080
CMD ["gotty", "-w", "--port", "8080", "--permit-write", "--once", "./tetris"]