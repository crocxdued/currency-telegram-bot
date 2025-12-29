package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/crocxdued/currency-telegram-bot/internal/app"
	"github.com/crocxdued/currency-telegram-bot/internal/config"
	"github.com/crocxdued/currency-telegram-bot/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	log.Println("🚀 Starting Currency Bot...")

	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// 2. Инициализация логгера
	if err := logger.InitGlobal(cfg.LogLevel); err != nil {
		log.Fatalf("Logger error: %v", err)
	}

	// 3. АВТОМАТИЧЕСКИЕ МИГРАЦИИ
	// Теперь они запускаются всегда перед стартом бота
	logger.S.Info("Checking and running migrations...")
	if err := runMigrations(cfg); err != nil {
		logger.S.Fatalf("Migration failed: %v", err)
	}
	logger.S.Info("Migrations status: OK")

	// 4. Запуск приложения
	logger.S.Info("Starting application...")
	application := app.New(cfg)

	if err := application.Run(); err != nil {
		logger.S.Errorf("Application failed: %v", err)
		os.Exit(1)
	}
}

func runMigrations(cfg *config.Config) error {
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Указываем папку, где лежит ваш файл 001_create_user_favorites.sql
	// В Docker она должна быть скопирована в корень или рядом с исполняемым файлом
	migrationDir := "migrations"

	if err := goose.RunContext(context.Background(), "up", db, migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
