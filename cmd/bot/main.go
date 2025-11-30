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

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	if err := logger.InitGlobal(cfg.LogLevel); err != nil {
		log.Fatalf("Logger error: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		logger.S.Info("Running migrations...")
		if err := runMigrations(cfg); err != nil {
			logger.S.Fatalf("Migration failed: %v", err)
		}
		logger.S.Info("Migrations completed")
		return
	}

	logger.S.Info("Starting application...")
	app := app.New(cfg)

	if err := app.Run(); err != nil {
		logger.S.Errorf("Application failed: %v", err)
		os.Exit(1)
	}
}

func runMigrations(cfg *config.Config) error {
	// Используем стандартный database/sql
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	command := "up"
	if len(os.Args) > 2 {
		command = os.Args[2]
	}

	// Запускаем миграции через goose
	if err := goose.RunContext(context.Background(), command, db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
