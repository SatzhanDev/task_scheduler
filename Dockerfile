# ===== STAGE 1: build =====
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Сначала зависимости (для кэша)
COPY go.mod go.sum ./
RUN go mod download

# Потом весь код
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/api

# ===== STAGE 2: runtime =====
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/app /app/app

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/app"]