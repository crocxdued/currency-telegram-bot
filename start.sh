#!/bin/bash
set -e

echo "🚀 Starting Currency Telegram Bot..."

# Делаем бинарник исполняемым (на всякий случай)
chmod +x ./bot

# Выполняем миграции
if [ -n "$DB_URL" ]; then
    echo "📦 Running database migrations..."
    ./bot migrate
fi

echo "🤖 Starting bot..."
exec ./bot