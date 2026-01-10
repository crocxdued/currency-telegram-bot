package app

import (
	"context"
	"fmt"
	"time"

	"github.com/crocxdued/currency-telegram-bot/internal/config"
	"github.com/crocxdued/currency-telegram-bot/internal/domain/services"
	"github.com/crocxdued/currency-telegram-bot/internal/interfaces/exchanger/cbr"
	"github.com/crocxdued/currency-telegram-bot/internal/interfaces/exchanger/exchangeratehost"
	"github.com/crocxdued/currency-telegram-bot/internal/interfaces/handlers"
	"github.com/crocxdued/currency-telegram-bot/internal/interfaces/repository/cache"
	"github.com/crocxdued/currency-telegram-bot/internal/interfaces/repository/postgres"
	"github.com/crocxdued/currency-telegram-bot/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type App struct {
	config *config.Config
	db     *sqlx.DB
	bot    *tgbotapi.BotAPI
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) initDB(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, "postgres", a.config.DBURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	a.db = db
	logger.S.Info("Database connection established")
	return nil
}

func (a *App) initBot() error {
	bot, err := tgbotapi.NewBotAPI(a.config.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = false 
	a.bot = bot

	logger.S.Infof("Authorized on account %s", bot.Self.UserName)
	return nil
}

func (a *App) initServices() (*handlers.BotHandler, error) {
	
	ratesCache := cache.NewRatesCache(a.config.CacheTTLMinutes)

	providers := []services.ExchangeProvider{
		exchangeratehost.New(),
		cbr.New(),
	}

	exchangeService := services.NewExchangeService(providers, ratesCache)

	favoritesRepo := postgres.NewFavoritesRepository(a.db)

	botHandler := handlers.NewBotHandler(a.bot, exchangeService, favoritesRepo)

	return botHandler, nil
}

func (a *App) Run() error {
	ctx := context.Background()

	logger.S.Info("Starting application...")

	if err := a.initDB(ctx); err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}
	defer a.db.Close()

	if err := a.initBot(); err != nil {
		return fmt.Errorf("bot initialization failed: %w", err)
	}

	botHandler, err := a.initServices()
	if err != nil {
		return fmt.Errorf("services initialization failed: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := a.bot.GetUpdatesChan(u)

	logger.S.Info("Bot is now running. Press Ctrl+C to exit.")

	for update := range updates {
		botHandler.HandleUpdate(update)
	}

	return nil
}
