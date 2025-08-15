# Build stage
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем основное приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

# Собираем producer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o producer ./internal/infrastructure/kafka/producer/producer.go

# Runtime stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/producer .
COPY --from=builder /app/internal/infrastructure/kafka/producer/orders ./orders
COPY --from=builder /app/config ./config
COPY --from=builder /app/ui ./ui

EXPOSE 8080

CMD ["./main"]
