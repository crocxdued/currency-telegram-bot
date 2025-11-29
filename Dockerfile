
FROM golang:1.24-alpine

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bot ./cmd/bot
RUN chmod +x /app/bot

WORKDIR /app
CMD ["./bot"]