# ---------- Builder ----------
    FROM golang:1.24-alpine AS builder

    # gcc + libpq-dev для CGO (lib/pq)
    RUN apk add --no-cache gcc musl-dev
    
    WORKDIR /app
    
    # Кэшируем зависимости
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Копируем весь код
    COPY . .
    
    # Собираем бинарник в корень образа builder'а (абсолютный путь!)
    RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /currency-bot ./cmd/bot
    
    # ---------- Runtime ----------
    FROM alpine:latest
    
    RUN apk --no-cache add ca-certificates tzdata
    
    WORKDIR /app
    
    # Копируем точно из корня builder'а — 100% найдёт
    COPY --from=builder /currency-bot ./currency-bot
    COPY --from=builder /app/migrations ./migrations
    
    # Делаем исполняемым
    RUN chmod +x ./currency-bot
    
    EXPOSE 8080
    
    CMD ["./currency-bot"]