
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем и делаем исполняемым
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/bot ./cmd/bot
RUN chmod +x /go/bin/bot

WORKDIR /go/bin

EXPOSE 8080

CMD ["./bot"]